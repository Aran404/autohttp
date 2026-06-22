package analyze

import (
	"testing"
	"time"

	"github.com/autohttp/autohttp/internal/tree"
	"github.com/autohttp/autohttp/session"
)

func TestValueFlowsFromResponseToRequest(t *testing.T) {
	sess := &session.Session{
		Exchanges: []*session.Exchange{
			{ID: "ex1", StartedAt: time.UnixMilli(100)},
			{ID: "ex2", StartedAt: time.UnixMilli(200)},
		},
	}

	trees := []*tree.Tree{
		{
			ExchangeID: "ex1",
			Root: &tree.Node{
				Path: "", Type: "object",
				Children: []*tree.Node{
					{
						Path: "response", Type: "object",
						Children: []*tree.Node{
							{Path: "body", Type: "object",
								Children: []*tree.Node{
									{Path: "token", Type: "string", RawValue: "abc123", Source: "ex1", SourceField: "response.body"},
								},
							},
						},
					},
				},
			},
		},
		{
			ExchangeID: "ex2",
			Root: &tree.Node{
				Path: "", Type: "object",
				Children: []*tree.Node{
					{
						Path: "request", Type: "object",
						Children: []*tree.Node{
							{Path: "url", Type: "object",
								Children: []*tree.Node{
									{Path: "query", Type: "object",
										Children: []*tree.Node{
											{Path: "token", Type: "string", RawValue: "abc123", Source: "ex2", SourceField: "request.url.query"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	result := Analyze(sess, trees)

	if len(result.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(result.Dependencies))
	}

	dep := result.Dependencies[0]
	if dep.From != "ex1" {
		t.Fatalf("expected From='ex1', got %q", dep.From)
	}
	if dep.To != "ex2" {
		t.Fatalf("expected To='ex2', got %q", dep.To)
	}
	if dep.Value != "abc123" {
		t.Fatalf("expected Value='abc123', got %q", dep.Value)
	}
	if dep.Confidence != 1.0 {
		t.Fatalf("expected Confidence=1.0, got %f", dep.Confidence)
	}
	if dep.Reason != "exact match" {
		t.Fatalf("expected Reason='exact match', got %q", dep.Reason)
	}
}

func TestSelfReferencesNotDetected(t *testing.T) {
	sess := &session.Session{
		Exchanges: []*session.Exchange{
			{ID: "ex1", StartedAt: time.UnixMilli(100)},
		},
	}

	trees := []*tree.Tree{
		{
			ExchangeID: "ex1",
			Root: &tree.Node{
				Path: "", Type: "object",
				Children: []*tree.Node{
					{Path: "method", Type: "string", RawValue: "GET", Source: "ex1"},
				},
			},
		},
	}

	result := Analyze(sess, trees)

	if len(result.Dependencies) != 0 {
		t.Fatalf("expected 0 dependencies for single tree, got %d", len(result.Dependencies))
	}
}

func TestValuesNotSharedProduceNoDependency(t *testing.T) {
	sess := &session.Session{
		Exchanges: []*session.Exchange{
			{ID: "ex1", StartedAt: time.UnixMilli(100)},
			{ID: "ex2", StartedAt: time.UnixMilli(200)},
		},
	}

	trees := []*tree.Tree{
		{
			ExchangeID: "ex1",
			Root: &tree.Node{
				Path: "", Type: "object",
				Children: []*tree.Node{
					{Path: "method", Type: "string", RawValue: "GET", Source: "ex1"},
				},
			},
		},
		{
			ExchangeID: "ex2",
			Root: &tree.Node{
				Path: "", Type: "object",
				Children: []*tree.Node{
					{Path: "method", Type: "string", RawValue: "POST", Source: "ex2"},
				},
			},
		},
	}

	result := Analyze(sess, trees)

	if len(result.Dependencies) != 0 {
		t.Fatalf("expected 0 dependencies, got %d", len(result.Dependencies))
	}
}

func TestMultipleExchangesWithValuePropagation(t *testing.T) {
	sess := &session.Session{
		Exchanges: []*session.Exchange{
			{ID: "ex1", StartedAt: time.UnixMilli(100)},
			{ID: "ex2", StartedAt: time.UnixMilli(200)},
			{ID: "ex3", StartedAt: time.UnixMilli(300)},
		},
	}

	trees := []*tree.Tree{
		{
			ExchangeID: "ex1",
			Root: &tree.Node{
				Path: "", Type: "object",
				Children: []*tree.Node{
					{
						Path: "response", Type: "object",
						Children: []*tree.Node{
							{Path: "body", Type: "object",
								Children: []*tree.Node{
									{Path: "session_id", Type: "string", RawValue: "sess_001", Source: "ex1", SourceField: "response.body"},
									{Path: "token", Type: "string", RawValue: "tok_abc", Source: "ex1", SourceField: "response.body"},
								},
							},
						},
					},
				},
			},
		},
		{
			ExchangeID: "ex2",
			Root: &tree.Node{
				Path: "", Type: "object",
				Children: []*tree.Node{
					{
						Path: "request", Type: "object",
						Children: []*tree.Node{
							{Path: "url", Type: "object",
								Children: []*tree.Node{
									{Path: "query", Type: "object",
										Children: []*tree.Node{
											{Path: "session_id", Type: "string", RawValue: "sess_001", Source: "ex2", SourceField: "request.url.query"},
										},
									},
								},
							},
						},
					},
					{
						Path: "response", Type: "object",
						Children: []*tree.Node{
							{Path: "body", Type: "object",
								Children: []*tree.Node{
									{Path: "refresh", Type: "string", RawValue: "tok_def", Source: "ex2", SourceField: "response.body"},
								},
							},
						},
					},
				},
			},
		},
		{
			ExchangeID: "ex3",
			Root: &tree.Node{
				Path: "", Type: "object",
				Children: []*tree.Node{
					{
						Path: "request", Type: "object",
						Children: []*tree.Node{
							{Path: "url", Type: "object",
								Children: []*tree.Node{
									{Path: "query", Type: "object",
										Children: []*tree.Node{
											{Path: "token", Type: "string", RawValue: "tok_abc", Source: "ex3", SourceField: "request.url.query"},
											{Path: "refresh", Type: "string", RawValue: "tok_def", Source: "ex3", SourceField: "request.url.query"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	result := Analyze(sess, trees)

	if len(result.Dependencies) == 0 {
		t.Fatal("expected dependencies, got none")
	}

	deps := make(map[string]int)
	for _, d := range result.Dependencies {
		key := d.From + "->" + d.To + ":" + d.Value
		deps[key]++
		verifyDependency(t, d)
	}

	if deps["ex1->ex2:sess_001"] != 1 {
		t.Errorf("expected ex1->ex2:sess_001 dependency, deps=%v", deps)
	}
	if deps["ex1->ex3:tok_abc"] != 1 {
		t.Errorf("expected ex1->ex3:tok_abc dependency, deps=%v", deps)
	}
	if deps["ex2->ex3:tok_def"] != 1 {
		t.Errorf("expected ex2->ex3:tok_def dependency, deps=%v", deps)
	}
}

func verifyDependency(t *testing.T, d *Dependency) {
	t.Helper()
	if d.From == "" {
		t.Error("dependency From is empty")
	}
	if d.To == "" {
		t.Error("dependency To is empty")
	}
	if d.Value == "" {
		t.Error("dependency Value is empty")
	}
	if d.Path == "" {
		t.Error("dependency Path (source) is empty")
	}
	if d.TargetPath == "" {
		t.Error("dependency TargetPath is empty")
	}
	if d.Confidence != 1.0 {
		t.Errorf("expected Confidence=1.0, got %f", d.Confidence)
	}
	if d.Reason != "exact match" {
		t.Errorf("expected Reason='exact match', got %q", d.Reason)
	}
}

func TestEmptySession(t *testing.T) {
	result := Analyze(nil, nil)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Dependencies) != 0 {
		t.Fatalf("expected 0 dependencies, got %d", len(result.Dependencies))
	}
}

func TestEmptyExchanges(t *testing.T) {
	sess := &session.Session{}
	result := Analyze(sess, nil)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Dependencies) != 0 {
		t.Fatalf("expected 0 dependencies, got %d", len(result.Dependencies))
	}
}

func TestOrderIndependence(t *testing.T) {
	sess := &session.Session{
		Exchanges: []*session.Exchange{
			{ID: "ex1", StartedAt: time.UnixMilli(100)},
			{ID: "ex2", StartedAt: time.UnixMilli(200)},
			{ID: "ex3", StartedAt: time.UnixMilli(300)},
		},
	}

	trees := []*tree.Tree{
		{
			ExchangeID: "ex3",
			Root: &tree.Node{
				Path: "", Type: "object",
				Children: []*tree.Node{
					{Path: "request", Type: "object",
						Children: []*tree.Node{
							{Path: "key", Type: "string", RawValue: "val1", Source: "ex3", SourceField: "request"},
						},
					},
				},
			},
		},
		{
			ExchangeID: "ex2",
			Root: &tree.Node{
				Path: "", Type: "object",
				Children: []*tree.Node{
					{Path: "response", Type: "object",
						Children: []*tree.Node{
							{Path: "key", Type: "string", RawValue: "val1", Source: "ex2", SourceField: "response"},
						},
					},
				},
			},
		},
		{
			ExchangeID: "ex1",
			Root: &tree.Node{
				Path: "", Type: "object",
				Children: []*tree.Node{
					{Path: "response", Type: "object",
						Children: []*tree.Node{
							{Path: "key", Type: "string", RawValue: "val0", Source: "ex1", SourceField: "response"},
						},
					},
				},
			},
		},
	}

	result := Analyze(sess, trees)

	if len(result.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency (ex2->ex3), got %d", len(result.Dependencies))
	}
	if result.Dependencies[0].From != "ex2" {
		t.Fatalf("expected From='ex2', got %q", result.Dependencies[0].From)
	}
	if result.Dependencies[0].To != "ex3" {
		t.Fatalf("expected To='ex3', got %q", result.Dependencies[0].To)
	}
}

func TestDynamicClassificationByShape(t *testing.T) {
	sess := &session.Session{
		Exchanges: []*session.Exchange{
			{ID: "ex1", StartedAt: time.UnixMilli(100)},
		},
	}

	trees := []*tree.Tree{
		{
			ExchangeID: "ex1",
			Root: &tree.Node{
				Path: "", Type: "object",
				Children: []*tree.Node{
					{Path: "token", Type: "string", RawValue: "eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxIn0.MTIzNDU2", Shape: "jwt", Source: "ex1"},
					{Path: "id", Type: "string", RawValue: "550e8400-e29b-41d4-a716-446655440000", Shape: "uuid", Source: "ex1"},
					{Path: "name", Type: "string", RawValue: "Alice", Source: "ex1"},
				},
			},
		},
	}

	result := Analyze(sess, trees)

	if len(result.Dynamic) == 0 {
		t.Fatal("expected dynamic fields")
	}

	dynSet := make(map[string]bool)
	for _, d := range result.Dynamic {
		dynSet[d] = true
	}

	if !dynSet["token"] {
		t.Errorf("expected 'token' to be dynamic (jwt shape)")
	}
	if !dynSet["id"] {
		t.Errorf("expected 'id' to be dynamic (uuid shape)")
	}

	if len(result.Static) == 0 {
		t.Fatal("expected static fields")
	}

	staticSet := make(map[string]bool)
	for _, s := range result.Static {
		staticSet[s] = true
	}
	if !staticSet["name"] {
		t.Errorf("expected 'name' to be static")
	}
}

func TestDynamicClassificationByChangingValue(t *testing.T) {
	sess := &session.Session{
		Exchanges: []*session.Exchange{
			{ID: "ex1", StartedAt: time.UnixMilli(100)},
			{ID: "ex2", StartedAt: time.UnixMilli(200)},
		},
	}

	trees := []*tree.Tree{
		{
			ExchangeID: "ex1",
			Root: &tree.Node{
				Path: "", Type: "object",
				Children: []*tree.Node{
					{Path: "counter", Type: "number", RawValue: "1", Source: "ex1"},
				},
			},
		},
		{
			ExchangeID: "ex2",
			Root: &tree.Node{
				Path: "", Type: "object",
				Children: []*tree.Node{
					{Path: "counter", Type: "number", RawValue: "2", Source: "ex2"},
				},
			},
		},
	}

	result := Analyze(sess, trees)

	found := false
	for _, d := range result.Dynamic {
		if d == "counter" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'counter' to be dynamic (different values)")
	}
}
