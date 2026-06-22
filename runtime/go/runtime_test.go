package gort

import (
	"testing"
)

func TestExtractJSON(t *testing.T) {
	body := []byte(`{"data":{"token":"abc123","user":{"id":42}}}`)

	tests := []struct {
		path     string
		want     string
		wantFail bool
	}{
		{"data.token", "abc123", false},
		{"data.user.id", "42", false},
		{"missing", "", true},
		{"data.missing", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got, err := ExtractJSON(body, tt.path)
			if tt.wantFail {
				if err == nil {
					t.Errorf("ExtractJSON(%q) = %q, nil; want error", tt.path, got)
				}
				return
			}
			if err != nil {
				t.Errorf("ExtractJSON(%q): %v", tt.path, err)
				return
			}
			if got != tt.want {
				t.Errorf("ExtractJSON(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}
