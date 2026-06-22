package session

import (
	"testing"

	pb "github.com/autohttp/autohttp/session/gen/autohttp/v1"
)

func TestFromProtoNil(t *testing.T) {
	if got := FromProto(nil); got != nil {
		t.Errorf("FromProto(nil) = %v, want nil", got)
	}
}

func TestFromProtoBasic(t *testing.T) {
	pb := &pb.RecordedSession{
		Id:        "test-1",
		TargetUrl: "https://example.com",
		Exchanges: []*pb.HttpExchange{
			{
				Id: "req-1",
				Request: &pb.Request{
					Method: "GET",
					Url:    "https://example.com/login",
					Headers: []*pb.Header{
						{Key: "Accept", Value: "text/html"},
					},
					Cookies: []*pb.CookieMutation{
						{Name: "session", Value: "abc"},
					},
				},
				Response: &pb.Response{
					Status: 200,
				},
			},
		},
	}
	s := FromProto(pb)
	if s == nil {
		t.Fatal("FromProto returned nil")
	}
	if s.ID != "test-1" {
		t.Errorf("s.ID = %q, want %q", s.ID, "test-1")
	}
	if len(s.Exchanges) != 1 {
		t.Fatalf("len(s.Exchanges) = %d, want 1", len(s.Exchanges))
	}
	e := s.Exchanges[0]
	if e.ID != "req-1" {
		t.Errorf("e.ID = %q, want %q", e.ID, "req-1")
	}
	if e.Request.Method != "GET" {
		t.Errorf("e.Request.Method = %q, want %q", e.Request.Method, "GET")
	}
	if e.Request.Headers["Accept"] != "text/html" {
		t.Errorf("e.Request.Headers[Accept] = %q, want %q", e.Request.Headers["Accept"], "text/html")
	}
	if e.Request.Cookies["session"] != "abc" {
		t.Errorf("e.Request.Cookies[session] = %q, want %q", e.Request.Cookies["session"], "abc")
	}
	if e.Response.Status != 200 {
		t.Errorf("e.Response.Status = %d, want %d", e.Response.Status, 200)
	}
}
