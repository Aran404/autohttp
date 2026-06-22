package importer

import (
	"os"
	"testing"

	"github.com/autohttp/autohttp/session"
)

func TestImportFixture(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
		check   func(t *testing.T, s *session.Session)
	}{
		{
			name: "simple exchange",
			content: `{"exchanges": [{"id": "1", "request": {"method": "GET", "url": "http://example.com"}}]}`,
			wantErr: false,
			check: func(t *testing.T, s *session.Session) {
				if len(s.Exchanges) != 1 {
					t.Fatalf("expected 1 exchange, got %d", len(s.Exchanges))
				}
				if s.Exchanges[0].Request.Method != "GET" {
					t.Errorf("expected method GET, got %s", s.Exchanges[0].Request.Method)
				}
				if s.Exchanges[0].Request.URL != "http://example.com" {
					t.Errorf("expected url http://example.com, got %s", s.Exchanges[0].Request.URL)
				}
			},
		},
		{
			name: "empty exchanges",
			content: `{"exchanges": []}`,
			wantErr: false,
			check: func(t *testing.T, s *session.Session) {
				if len(s.Exchanges) != 0 {
					t.Fatalf("expected 0 exchanges, got %d", len(s.Exchanges))
				}
			},
		},
		{
			name:    "invalid json",
			content: `not json`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmp, err := os.CreateTemp("", "fixture*.json")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmp.Name())

			if _, err := tmp.WriteString(tt.content); err != nil {
				t.Fatal(err)
			}
			tmp.Close()

			got, err := ImportFixture(tmp.Name())
			if (err != nil) != tt.wantErr {
				t.Fatalf("ImportFixture() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			tt.check(t, got)
		})
	}
}
