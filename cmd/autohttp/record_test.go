package main

import (
	"reflect"
	"testing"

	pb "github.com/autohttp/autohttp/gen/autohttp/v1"
)

func TestSplitArgs(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		boolFlags map[string]bool
		wantFlags []string
		wantPos   []string
	}{
		{
			name:      "url first then flags",
			args:      []string{"https://example.com", "--browser", "camoufox", "--endpoints", "/login"},
			boolFlags: map[string]bool{"no-ai": true, "geoip": true},
			wantFlags: []string{"--browser", "camoufox", "--endpoints", "/login"},
			wantPos:   []string{"https://example.com"},
		},
		{
			name:      "flags first then url",
			args:      []string{"--browser", "camoufox", "https://example.com"},
			boolFlags: map[string]bool{"no-ai": true},
			wantFlags: []string{"--browser", "camoufox"},
			wantPos:   []string{"https://example.com"},
		},
		{
			name:      "bool flag without value does not consume next",
			args:      []string{"--geoip", "https://example.com"},
			boolFlags: map[string]bool{"geoip": true},
			wantFlags: []string{"--geoip"},
			wantPos:   []string{"https://example.com"},
		},
		{
			name:      "inline equals value",
			args:      []string{"https://example.com", "--browser=camoufox"},
			boolFlags: map[string]bool{"no-ai": true},
			wantFlags: []string{"--browser=camoufox"},
			wantPos:   []string{"https://example.com"},
		},
		{
			name:      "repeated endpoints",
			args:      []string{"--endpoints", "/a", "--endpoints", "/b", "https://example.com"},
			boolFlags: map[string]bool{"no-ai": true},
			wantFlags: []string{"--endpoints", "/a", "--endpoints", "/b"},
			wantPos:   []string{"https://example.com"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFlags, gotPos := splitArgs(tt.args, tt.boolFlags)
			if !reflect.DeepEqual(gotFlags, tt.wantFlags) {
				t.Errorf("flags = %#v, want %#v", gotFlags, tt.wantFlags)
			}
			if !reflect.DeepEqual(gotPos, tt.wantPos) {
				t.Errorf("positional = %#v, want %#v", gotPos, tt.wantPos)
			}
		})
	}
}

func TestParseCompletionPolicy(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantErr    bool
		wantPolicy *pb.CompletionPolicy
	}{
		{
			name:  "default network-idle",
			input: "network-idle:5s",
			wantPolicy: &pb.CompletionPolicy{
				Policy: &pb.CompletionPolicy_NetworkIdle{
					NetworkIdle: &pb.NetworkIdleSettle{QuietWindowMs: 5000},
				},
			},
		},
		{
			name:  "network-idle default duration",
			input: "network-idle",
			wantPolicy: &pb.CompletionPolicy{
				Policy: &pb.CompletionPolicy_NetworkIdle{
					NetworkIdle: &pb.NetworkIdleSettle{QuietWindowMs: 5000},
				},
			},
		},
		{
			name:  "response-only",
			input: "response-only",
			wantPolicy: &pb.CompletionPolicy{
				Policy: &pb.CompletionPolicy_ResponseOnly{
					ResponseOnly: &pb.ResponseOnlySettle{},
				},
			},
		},
		{
			name:  "url",
			input: "url:/dashboard",
			wantPolicy: &pb.CompletionPolicy{
				Policy: &pb.CompletionPolicy_UrlSettle{
					UrlSettle: &pb.UrlSettle{Url: "/dashboard"},
				},
			},
		},
		{
			name:  "timeout",
			input: "timeout:10s",
			wantPolicy: &pb.CompletionPolicy{
				Policy: &pb.CompletionPolicy_Timeout{
					Timeout: &pb.TimeoutSettle{DurationMs: 10000},
				},
			},
		},
		{name: "unknown policy", input: "bogus", wantErr: true},
		{name: "url without path", input: "url", wantErr: true},
		{name: "bad duration", input: "timeout:not-a-duration", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseCompletionPolicy(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got policy = %v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseCompletionPolicy(%q): %v", tt.input, err)
			}
			if !reflect.DeepEqual(got, tt.wantPolicy) {
				t.Errorf("policy = %v, want %v", got, tt.wantPolicy)
			}
		})
	}
}

func TestBuildEndpointGoals(t *testing.T) {
	goals := buildEndpointGoals([]string{"/login", "/checksum", "/done"}, nil)
	if len(goals) != 3 {
		t.Fatalf("len(goals) = %d, want 3", len(goals))
	}
	if goals[0].Id != "endpoint-0" {
		t.Errorf("goals[0].Id = %q, want endpoint-0", goals[0].Id)
	}
	if goals[0].Terminal {
		t.Errorf("goals[0].Terminal = true, want false")
	}
	if !goals[2].Terminal {
		t.Errorf("goals[2].Terminal = false, want true")
	}
	if goals[2].Completion != nil {
		t.Errorf("goals[2].Completion should be nil when no policy provided")
	}
}
