package generate

import (
	"bytes"
	"go/parser"
	"go/token"
	"testing"

	"github.com/autohttp/autohttp/internal/analyze"
	"github.com/autohttp/autohttp/session"
)

func TestGoScriptReturnsValidGoCode(t *testing.T) {
	sess := &session.Session{
		ID:        "test-session",
		TargetURL: "https://example.com",
		Exchanges: []*session.Exchange{
			{ID: "ex1"},
		},
	}
	analysis := &analyze.Result{}

	code, err := GoScript(sess, analysis)
	if err != nil {
		t.Fatalf("GoScript failed: %v", err)
	}

	if !bytes.Contains(code, []byte("package main")) {
		t.Fatal("generated code does not contain 'package main'")
	}

	_, err = parser.ParseFile(token.NewFileSet(), "", code, parser.PackageClauseOnly)
	if err != nil {
		t.Fatalf("generated code is not valid Go:\n%v\n\nGenerated code:\n%s", err, string(code))
	}
}

func TestGeneratedCodeContainsSessionID(t *testing.T) {
	sessID := "sess-uniq-123"
	sess := &session.Session{
		ID:        sessID,
		TargetURL: "https://example.com",
		Exchanges: []*session.Exchange{
			{ID: "ex1"},
		},
	}
	analysis := &analyze.Result{}

	code, err := GoScript(sess, analysis)
	if err != nil {
		t.Fatalf("GoScript failed: %v", err)
	}

	if !bytes.Contains(code, []byte(sessID)) {
		t.Fatalf("expected generated code to contain session ID %q\n\nGenerated code:\n%s", sessID, string(code))
	}
}

func TestGeneratedCodeContainsDependencyValues(t *testing.T) {
	sess := &session.Session{
		ID:        "test-sess",
		TargetURL: "https://example.com",
		Exchanges: []*session.Exchange{
			{ID: "ex1"},
			{ID: "ex2"},
		},
	}
	analysis := &analyze.Result{
		Dependencies: []*analyze.Dependency{
			{
				From:       "ex1",
				To:         "ex2",
				Value:      "token_abc_123",
				Path:       "response.body.token",
				TargetPath: "request.url.query.token",
			},
		},
	}

	code, err := GoScript(sess, analysis)
	if err != nil {
		t.Fatalf("GoScript failed: %v", err)
	}

	for _, want := range []string{"ex1", "ex2", "token_abc_123"} {
		if !bytes.Contains(code, []byte(want)) {
			t.Errorf("expected generated code to contain %q", want)
		}
	}
}

func TestGoScriptNoDependencies(t *testing.T) {
	sess := &session.Session{
		ID:        "no-dep-session",
		TargetURL: "https://example.com",
		Exchanges: []*session.Exchange{
			{ID: "ex1"},
		},
	}
	analysis := &analyze.Result{}

	code, err := GoScript(sess, analysis)
	if err != nil {
		t.Fatalf("GoScript failed: %v", err)
	}

	if !bytes.Contains(code, []byte("(none)")) {
		t.Errorf("expected generated code to contain '(none)' for empty dependencies")
	}
}

func TestGoScriptNilDependencies(t *testing.T) {
	sess := &session.Session{
		ID:        "nil-dep-session",
		TargetURL: "https://example.com",
		Exchanges: []*session.Exchange{
			{ID: "ex1"},
		},
	}
	analysis := &analyze.Result{
		Dependencies: nil,
	}

	code, err := GoScript(sess, analysis)
	if err != nil {
		t.Fatalf("GoScript failed: %v", err)
	}

	if !bytes.Contains(code, []byte("(none)")) {
		t.Errorf("expected generated code to contain '(none)' for nil dependencies")
	}
}
