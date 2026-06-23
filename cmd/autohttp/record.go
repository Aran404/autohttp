package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	pb "github.com/autohttp/autohttp/gen/autohttp/v1"
	"github.com/autohttp/autohttp/internal/record"
)

type stringSliceFlag struct {
	values []string
}

func (s *stringSliceFlag) String() string {
	return strings.Join(s.values, ",")
}

func (s *stringSliceFlag) Set(value string) error {
	s.values = append(s.values, value)
	return nil
}

func (s *stringSliceFlag) Values() []string {
	return s.values
}

func browserDisplayName(b pb.Browser) string {
	switch b {
	case pb.Browser_BROWSER_CAMOUFOX:
		return "camoufox"
	case pb.Browser_BROWSER_CLOAK:
		return "cloak"
	default:
		return pb.Browser_name[int32(b)]
	}
}

func parseCompletionPolicy(raw string) (*pb.CompletionPolicy, error) {
	parts := strings.SplitN(raw, ":", 2)
	switch parts[0] {
	case "network-idle":
		quietMs := int64(5000)
		if len(parts) == 2 {
			d, err := time.ParseDuration(parts[1])
			if err != nil {
				return nil, fmt.Errorf("parse network-idle duration %q: %w", parts[1], err)
			}
			quietMs = d.Milliseconds()
		}
		return &pb.CompletionPolicy{
			Policy: &pb.CompletionPolicy_NetworkIdle{
				NetworkIdle: &pb.NetworkIdleSettle{QuietWindowMs: quietMs},
			},
		}, nil
	case "response-only":
		return &pb.CompletionPolicy{
			Policy: &pb.CompletionPolicy_ResponseOnly{
				ResponseOnly: &pb.ResponseOnlySettle{},
			},
		}, nil
	case "url":
		if len(parts) != 2 || parts[1] == "" {
			return nil, fmt.Errorf("url completion requires url:<path>")
		}
		return &pb.CompletionPolicy{
			Policy: &pb.CompletionPolicy_UrlSettle{
				UrlSettle: &pb.UrlSettle{Url: parts[1]},
			},
		}, nil
	case "timeout":
		if len(parts) != 2 {
			return nil, fmt.Errorf("timeout completion requires timeout:<duration>")
		}
		d, err := time.ParseDuration(parts[1])
		if err != nil {
			return nil, fmt.Errorf("parse timeout duration %q: %w", parts[1], err)
		}
		return &pb.CompletionPolicy{
			Policy: &pb.CompletionPolicy_Timeout{
				Timeout: &pb.TimeoutSettle{DurationMs: d.Milliseconds()},
			},
		}, nil
	default:
		return nil, fmt.Errorf("unknown completion policy %q (want network-idle, response-only, url:<path>, or timeout:<duration>)", parts[0])
	}
}

func buildEndpointGoals(patterns []string, terminalCompletion *pb.CompletionPolicy) []*pb.EndpointGoal {
	goals := make([]*pb.EndpointGoal, len(patterns))
	for i, p := range patterns {
		goals[i] = &pb.EndpointGoal{
			Id:         "endpoint-" + strconv.Itoa(i),
			UrlPattern: p,
			Terminal:   i == len(patterns)-1,
		}
		if i == len(patterns)-1 && terminalCompletion != nil {
			goals[i].Completion = terminalCompletion
		}
	}
	return goals
}

func defaultSessionName() string {
	return "session-" + strconv.FormatInt(time.Now().Unix(), 10)
}

// splitArgs separates flags and positionals from a CLI arg vector
// so Go's stdlib flag package (which stops parsing at the first
// positional) can be used with the documented call shape
// `record <url> --flag value ...`. The caller supplies the set of
// bool flag names because the stdlib does not expose that info.
func splitArgs(args []string, boolFlagNames map[string]bool) (flagArgs []string, positional []string) {
	flagName := func(token string) (name string, hasValue bool, value string) {
		body := strings.TrimLeft(token, "-")
		if eq := strings.IndexByte(body, '='); eq >= 0 {
			return body[:eq], true, body[eq+1:]
		}
		return body, false, ""
	}
	for i := 0; i < len(args); i++ {
		token := args[i]
		if !strings.HasPrefix(token, "-") {
			positional = append(positional, token)
			continue
		}
		name, hasInlineValue, inlineValue := flagName(token)
		flagArgs = append(flagArgs, token)
		if hasInlineValue {
			_ = inlineValue
			continue
		}
		if boolFlagNames[name] {
			continue
		}
		if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
			flagArgs = append(flagArgs, args[i+1])
			i++
		}
	}
	return flagArgs, positional
}

func cmdRecord() {
	fs := flag.NewFlagSet("record", flag.ExitOnError)
	browser := fs.String("browser", "camoufox", "Browser engine: camoufox or cloak")
	completion := fs.String(
		"completion",
		"network-idle:5s",
		"Terminal completion policy: network-idle[:dur], response-only, url:<path>, timeout:<duration>",
	)
	sessionName := fs.String("session", "", "Session name (default: timestamp)")
	noAI := fs.Bool("no-ai", false, "Disable AI escalation")
	proxy := fs.String("proxy", "", "Optional proxy URL applied to the browser launch")
	geoip := fs.Bool("geoip", false, "Auto-detect timezone and locale from the proxy exit IP")

	var endpoints stringSliceFlag
	fs.Var(&endpoints, "endpoints", "Endpoint URL path; repeat for ordered milestones (the last one is terminal)")

	boolFlags := map[string]bool{
		"no-ai": true,
		"geoip": true,
	}
	flagArgs, positional := splitArgs(os.Args[2:], boolFlags)
	if err := fs.Parse(flagArgs); err != nil {
		os.Exit(2)
	}
	_ = noAI

	if len(positional) < 1 {
		fmt.Fprintln(os.Stderr, "autohttp: record <url> requires a URL argument")
		os.Exit(2)
	}
	targetURL := positional[0]

	if len(endpoints.values) == 0 {
		fmt.Fprintln(os.Stderr, "autohttp: --endpoints is required (one or more URL paths)")
		os.Exit(2)
	}

	browserEnum, err := record.BrowserFromName(*browser)
	if err != nil {
		fmt.Fprintln(os.Stderr, "autohttp:", err)
		os.Exit(2)
	}

	policy, err := parseCompletionPolicy(*completion)
	if err != nil {
		fmt.Fprintln(os.Stderr, "autohttp:", err)
		os.Exit(2)
	}

	goals := buildEndpointGoals(endpoints.values, policy)
	name := *sessionName
	if name == "" {
		name = defaultSessionName()
	}

	fmt.Printf("autohttp: session=%s target=%s browser=%s\n", name, targetURL, browserDisplayName(browserEnum))
	for i, g := range goals {
		terminal := ""
		if g.Terminal {
			terminal = " (terminal)"
		}
		fmt.Printf("  endpoint[%d] id=%s pattern=%s%s\n", i, g.Id, g.UrlPattern, terminal)
	}

	worker, err := record.StartWorker(record.StartConfig{
		Browser:   browserEnum,
		TargetURL: targetURL,
		Endpoints: goals,
		ProxyURL:  *proxy,
		GeoIP:     *geoip,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "autohttp: starting worker:", err)
		os.Exit(3)
	}
	fmt.Printf("autohttp: worker listening on port %d (send SIGINT to stop)\n", worker.Port())

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Fprintln(os.Stderr, "autohttp: interrupt, cancelling recording")
		_ = worker.SendCommand(&pb.BrowserCommand{
			Command: &pb.BrowserCommand_CancelRecording{
				CancelRecording: &pb.CancelRecording{Reason: "user interrupt"},
			},
		})
	}()

	for {
		event, err := worker.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, "autohttp: worker stream error:", err)
			break
		}
		switch {
		case event.GetBrowserLaunched() != nil:
			fmt.Println("autohttp: browser launched")
		case event.GetBrowserCrashed() != nil:
			c := event.GetBrowserCrashed()
			fmt.Fprintf(os.Stderr, "autohttp: browser crashed: %s\n", c.Reason)
		case event.GetEndpointRequestStarted() != nil:
			e := event.GetEndpointRequestStarted()
			fmt.Printf("autohttp: endpoint %s request started (request_id=%s)\n", e.EndpointId, e.RequestId)
		case event.GetEndpointResponseCompleted() != nil:
			e := event.GetEndpointResponseCompleted()
			fmt.Printf("autohttp: endpoint %s response completed (status=%d)\n", e.EndpointId, e.Status)
		case event.GetEndpointSettled() != nil:
			e := event.GetEndpointSettled()
			fmt.Printf("autohttp: terminal endpoint %s settled (%s)\n", e.EndpointId, e.SettleReason)
		case event.GetSessionFinalized() != nil:
			fmt.Println("autohttp: session finalized")
		case event.GetError() != nil:
			e := event.GetError()
			fmt.Fprintf(os.Stderr, "autohttp: error: %s\n", e.Message)
		default:
			fmt.Printf("autohttp: event: %v\n", event)
		}
		if event.GetSessionFinalized() != nil {
			break
		}
	}

	if err := worker.Stop("main exit"); err != nil {
		fmt.Fprintln(os.Stderr, "autohttp: stopping worker:", err)
	}
}
