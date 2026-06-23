package record

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/autohttp/autohttp/gen/autohttp/v1"
)

// PythonPackageDir returns the absolute path to the python/ directory
// in the source tree. It walks up from the current working directory
// looking for go.mod so the CLI works whether invoked from the project
// root or a subdirectory.
func PythonPackageDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getwd: %w", err)
	}
	dir := cwd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return filepath.Join(dir, "python"), nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found above %s", cwd)
		}
		dir = parent
	}
}

// BrowserFromName maps the CLI-friendly browser name to the
// Browser enum value. Unknown names return an error.
func BrowserFromName(name string) (pb.Browser, error) {
	switch strings.ToLower(name) {
	case "camoufox", "":
		return pb.Browser_BROWSER_CAMOUFOX, nil
	case "cloak", "cloakbrowser":
		return pb.Browser_BROWSER_CLOAK, nil
	default:
		return pb.Browser_BROWSER_UNSPECIFIED, fmt.Errorf("unknown browser %q", name)
	}
}

// StartConfig holds the parameters for a recording session sent to
// the Python worker in the StartRecording command. Completion
// policies live on each EndpointGoal; the CLI applies the user's
// --completion default when building the goal list.
type StartConfig struct {
	Browser   pb.Browser
	TargetURL string
	Endpoints []*pb.EndpointGoal
	ProxyURL  string
	GeoIP     bool
	Settings  map[string]string
}

// Worker manages the lifecycle of a Python browser worker subprocess
// and its bidirectional gRPC stream. One Worker is created per
// `autohttp record` invocation. Close must be called to release
// resources.
type Worker struct {
	cmd    *exec.Cmd
	conn   *grpc.ClientConn
	stream pb.BrowserWorker_RecordClient
	port   int
}

// StartWorker spawns the Python browser worker subprocess, waits for
// it to print its bound port, opens the gRPC stream, and sends
// StartRecording. The returned Worker is ready to receive events.
// The Worker manages its own subprocess and gRPC lifetime; call
// Stop to release them.
func StartWorker(cfg StartConfig) (*Worker, error) {
	pythonDir, err := PythonPackageDir()
	if err != nil {
		return nil, fmt.Errorf("locate python package: %w", err)
	}

	port, cmd, err := spawnPythonWorker(pythonDir)
	if err != nil {
		return nil, err
	}

	conn, err := grpc.NewClient(
		fmt.Sprintf("127.0.0.1:%d", port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return nil, fmt.Errorf("dial worker: %w", err)
	}

	client := pb.NewBrowserWorkerClient(conn)
	stream, err := client.Record(context.Background())
	if err != nil {
		_ = conn.Close()
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return nil, fmt.Errorf("open record stream: %w", err)
	}

	start := &pb.StartRecording{
		Browser:   cfg.Browser,
		TargetUrl: cfg.TargetURL,
		Endpoints: cfg.Endpoints,
		ProxyUrl:  cfg.ProxyURL,
		Geoip:     cfg.GeoIP,
		Settings:  cfg.Settings,
	}

	if err := stream.Send(&pb.BrowserCommand{
		Command: &pb.BrowserCommand_StartRecording{StartRecording: start},
	}); err != nil {
		_ = stream.CloseSend()
		_ = conn.Close()
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return nil, fmt.Errorf("send start_recording: %w", err)
	}

	return &Worker{cmd: cmd, conn: conn, stream: stream, port: port}, nil
}

// Port returns the port the worker subprocess bound to.
func (w *Worker) Port() int {
	return w.port
}

// Recv blocks until the next BrowserEvent arrives or the stream
// ends. io.EOF means the worker has closed the stream cleanly.
func (w *Worker) Recv() (*pb.BrowserEvent, error) {
	return w.stream.Recv()
}

// SendCommand sends a BrowserCommand upstream. The most common
// non-start command is CancelRecording.
func (w *Worker) SendCommand(cmd *pb.BrowserCommand) error {
	return w.stream.Send(cmd)
}

// Stop sends CancelRecording, waits for the worker subprocess to
// exit, and releases the gRPC connection. It is safe to call
// multiple times; subsequent calls are no-ops.
func (w *Worker) Stop(reason string) error {
	if w == nil || w.cmd == nil {
		return nil
	}
	if err := w.stream.Send(&pb.BrowserCommand{
		Command: &pb.BrowserCommand_CancelRecording{
			CancelRecording: &pb.CancelRecording{Reason: reason},
		},
	}); err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("send cancel_recording: %w", err)
	}
	_ = w.stream.CloseSend()
	_ = w.conn.Close()

	waitErr := w.cmd.Wait()
	if waitErr != nil {
		var exitErr *exec.ExitError
		if errors.As(waitErr, &exitErr) {
			return fmt.Errorf("worker exited with status %d: %w", exitErr.ExitCode(), waitErr)
		}
		return fmt.Errorf("worker wait: %w", waitErr)
	}
	return nil
}

// spawnPythonWorker launches `python3 -m autohttp_worker` with
// PYTHONPATH pointing at the project's python/ directory. It reads
// the first line of stdout, which is the bound port.
func spawnPythonWorker(pythonDir string) (int, *exec.Cmd, error) {
	cmd := exec.Command("python3", "-m", "autohttp_worker")
	cmd.Env = append(os.Environ(), "PYTHONPATH="+pythonDir)
	cmd.Stderr = os.Stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return 0, nil, fmt.Errorf("stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return 0, nil, fmt.Errorf("start worker: %w", err)
	}

	reader := bufio.NewReader(stdout)
	line, err := reader.ReadString('\n')
	if err != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return 0, nil, fmt.Errorf("read worker port: %w", err)
	}
	port, err := strconv.Atoi(strings.TrimSpace(line))
	if err != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return 0, nil, fmt.Errorf("parse worker port %q: %w", line, err)
	}
	return port, cmd, nil
}
