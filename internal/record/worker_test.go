package record

import (
	"testing"
	"time"

	pb "github.com/autohttp/autohttp/gen/autohttp/v1"
)

func TestStartWorkerReceivesBrowserLaunched(t *testing.T) {
	if testing.Short() {
		t.Skip("skips integration test in -short mode")
	}

	cfg := StartConfig{
		Browser:   pb.Browser_BROWSER_CAMOUFOX,
		TargetURL: "https://example.com",
		Endpoints: []*pb.EndpointGoal{
			{Id: "e1", UrlPattern: "/login"},
			{Id: "e2", UrlPattern: "/done", Terminal: true},
		},
	}

	w, err := StartWorker(cfg)
	if err != nil {
		t.Fatalf("StartWorker: %v", err)
	}
	if w == nil {
		t.Fatal("StartWorker returned nil worker")
	}
	t.Cleanup(func() {
		done := make(chan struct{})
		go func() {
			_ = w.Stop("test cleanup")
			close(done)
		}()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Log("Stop did not return within 5s")
		}
	})

	event, err := w.Recv()
	if err != nil {
		t.Fatalf("Recv: %v", err)
	}
	if event.GetBrowserLaunched() == nil {
		t.Fatalf("first event = %v, want BrowserLaunched", event)
	}

	if err := w.SendCommand(&pb.BrowserCommand{
		Command: &pb.BrowserCommand_CancelRecording{
			CancelRecording: &pb.CancelRecording{Reason: "test"},
		},
	}); err != nil {
		t.Fatalf("SendCommand: %v", err)
	}
}
