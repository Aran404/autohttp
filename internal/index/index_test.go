package index

import (
	"testing"

	"github.com/autohttp/autohttp/internal/tree"
)

func strPtr(s string) *string { return &s }

func TestAddAndLookup(t *testing.T) {
	idx := New()

	t1 := &tree.Tree{
		ExchangeID: "ex1",
		Root: &tree.Node{
			Path: "", Type: "object",
			Children: []*tree.Node{
				{
					Path: "request", Type: "object",
					Children: []*tree.Node{
						{Path: "method", Type: "string", RawValue: "GET"},
						{Path: "host", Type: "string", RawValue: "api.example.com"},
					},
				},
				{
					Path: "response", Type: "object",
					Children: []*tree.Node{
						{Path: "status", Type: "number", RawValue: "200"},
						{Path: "status_text", Type: "string", RawValue: "OK"},
					},
				},
			},
		},
	}

	idx.Add(t1)

	refs := idx.Lookup("GET")
	if len(refs) != 1 {
		t.Fatalf("expected 1 reference for 'GET', got %d", len(refs))
	}
	if refs[0].Tree != t1 {
		t.Fatal("reference Tree does not match")
	}
	if refs[0].NodePath != "request.method" {
		t.Fatalf("expected path 'request.method', got %q", refs[0].NodePath)
	}
	if refs[0].Node.RawValue != "GET" {
		t.Fatalf("expected node RawValue 'GET', got %q", refs[0].Node.RawValue)
	}

	refs = idx.Lookup("200")
	if len(refs) != 1 {
		t.Fatalf("expected 1 reference for '200', got %d", len(refs))
	}
	if refs[0].NodePath != "response.status" {
		t.Fatalf("expected path 'response.status', got %q", refs[0].NodePath)
	}

	refs = idx.Lookup("OK")
	if len(refs) != 1 {
		t.Fatalf("expected 1 reference for 'OK', got %d", len(refs))
	}
	if refs[0].NodePath != "response.status_text" {
		t.Fatalf("expected path 'response.status_text', got %q", refs[0].NodePath)
	}
}

func TestLookupMissing(t *testing.T) {
	idx := New()

	t1 := &tree.Tree{
		ExchangeID: "ex1",
		Root: &tree.Node{
			Path: "", Type: "object",
			Children: []*tree.Node{
				{Path: "method", Type: "string", RawValue: "GET"},
			},
		},
	}

	idx.Add(t1)

	refs := idx.Lookup("POST")
	if refs != nil {
		t.Fatalf("expected nil for missing value, got %d references", len(refs))
	}
}

func TestMultipleTrees(t *testing.T) {
	idx := New()

	t1 := &tree.Tree{
		ExchangeID: "ex1",
		Root: &tree.Node{
			Path: "", Type: "object",
			Children: []*tree.Node{
				{Path: "method", Type: "string", RawValue: "GET"},
			},
		},
	}

	t2 := &tree.Tree{
		ExchangeID: "ex2",
		Root: &tree.Node{
			Path: "", Type: "object",
			Children: []*tree.Node{
				{Path: "method", Type: "string", RawValue: "POST"},
			},
		},
	}

	t3 := &tree.Tree{
		ExchangeID: "ex3",
		Root: &tree.Node{
			Path: "", Type: "object",
			Children: []*tree.Node{
				{Path: "method", Type: "string", RawValue: "GET"},
			},
		},
	}

	idx.Add(t1)
	idx.Add(t2)
	idx.Add(t3)

	refs := idx.Lookup("GET")
	if len(refs) != 2 {
		t.Fatalf("expected 2 references for 'GET', got %d", len(refs))
	}

	ids := map[string]bool{}
	for _, r := range refs {
		ids[r.Tree.ExchangeID] = true
	}
	if !ids["ex1"] || !ids["ex3"] {
		t.Fatal("expected GET references from ex1 and ex3")
	}

	refs = idx.Lookup("POST")
	if len(refs) != 1 {
		t.Fatalf("expected 1 reference for 'POST', got %d", len(refs))
	}
}

func TestShapeIndexing(t *testing.T) {
	idx := New()

	t1 := &tree.Tree{
		ExchangeID: "ex1",
		Root: &tree.Node{
			Path: "", Type: "object",
			Children: []*tree.Node{
				{
					Path: "auth", Type: "object",
					Children: []*tree.Node{
						{Path: "token", Type: "string", RawValue: "eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.RTQ2NjMxMjM0NTY3ODkwMTIzNDU2", Shape: "jwt"},
						{Path: "id", Type: "string", RawValue: "550e8400-e29b-41d4-a716-446655440000", Shape: "uuid"},
						{Path: "login_at", Type: "string", RawValue: "2024-01-15T10:30:00Z", Shape: "timestamp"},
					},
				},
			},
		},
	}

	idx.Add(t1)

	jwtRefs := idx.LookupByShape("jwt")
	if len(jwtRefs) != 1 {
		t.Fatalf("expected 1 jwt ref, got %d", len(jwtRefs))
	}
	if jwtRefs[0].Node.Path != "token" {
		t.Fatalf("expected jwt node path 'token', got %q", jwtRefs[0].Node.Path)
	}

	uuidRefs := idx.LookupByShape("uuid")
	if len(uuidRefs) != 1 {
		t.Fatalf("expected 1 uuid ref, got %d", len(uuidRefs))
	}
	if uuidRefs[0].Node.Path != "id" {
		t.Fatalf("expected uuid node path 'id', got %q", uuidRefs[0].Node.Path)
	}

	tsRefs := idx.LookupByShape("timestamp")
	if len(tsRefs) != 1 {
		t.Fatalf("expected 1 timestamp ref, got %d", len(tsRefs))
	}
	if tsRefs[0].Node.Path != "login_at" {
		t.Fatalf("expected timestamp node path 'login_at', got %q", tsRefs[0].Node.Path)
	}
}

func TestShapeDetectionFallback(t *testing.T) {
	idx := New()

	t1 := &tree.Tree{
		ExchangeID: "ex1",
		Root: &tree.Node{
			Path: "", Type: "object",
			Children: []*tree.Node{
				{Path: "jwt", Type: "string", RawValue: "eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxIn0.MTIzNDU2"},
				{Path: "uuid", Type: "string", RawValue: "550e8400-e29b-41d4-a716-446655440000"},
				{Path: "ts_nano", Type: "string", RawValue: "1705315230000000000"},
				{Path: "ts_sec", Type: "string", RawValue: "1705315230"},
				{Path: "ts_ms", Type: "string", RawValue: "1705315230000"},
			},
		},
	}

	idx.Add(t1)

	if refs := idx.LookupByShape("jwt"); len(refs) != 1 {
		t.Fatalf("expected 1 auto-detected jwt ref, got %d", len(refs))
	}
	if refs := idx.LookupByShape("uuid"); len(refs) != 1 {
		t.Fatalf("expected 1 auto-detected uuid ref, got %d", len(refs))
	}
	if refs := idx.LookupByShape("timestamp"); len(refs) != 3 {
		t.Fatalf("expected 3 auto-detected timestamp refs, got %d", len(refs))
	}
}

func TestDecodedIndex(t *testing.T) {
	idx := New()

	t1 := &tree.Tree{
		ExchangeID: "ex1",
		Root: &tree.Node{
			Path: "", Type: "object",
			Children: []*tree.Node{
				{Path: "email", Type: "string", RawValue: "User@Example.COM"},
			},
		},
	}

	idx.Add(t1)

	refs := idx.LookupDecoded("user@example.com")
	if len(refs) != 1 {
		t.Fatalf("expected 1 decoded ref, got %d", len(refs))
	}
	if refs[0].NodePath != "email" {
		t.Fatalf("expected path 'email', got %q", refs[0].NodePath)
	}
}

func TestNilTree(t *testing.T) {
	idx := New()
	idx.Add(nil)
	idx.Add(&tree.Tree{Root: nil})
	if refs := idx.Lookup("anything"); refs != nil {
		t.Fatal("expected nil for empty index")
	}
}

func TestBooleanAndNumberValues(t *testing.T) {
	idx := New()

	t1 := &tree.Tree{
		ExchangeID: "ex1",
		Root: &tree.Node{
			Path: "", Type: "object",
			Children: []*tree.Node{
				{Path: "active", Type: "boolean", RawValue: "true"},
				{Path: "count", Type: "number", RawValue: "42"},
			},
		},
	}

	idx.Add(t1)

	if refs := idx.Lookup("true"); len(refs) != 1 {
		t.Fatalf("expected 1 ref for boolean true, got %d", len(refs))
	}
	if refs := idx.Lookup("42"); len(refs) != 1 {
		t.Fatalf("expected 1 ref for number 42, got %d", len(refs))
	}
}

func TestConcurrentAccess(t *testing.T) {
	idx := New()

	t1 := &tree.Tree{
		ExchangeID: "ex1",
		Root: &tree.Node{
			Path: "", Type: "object",
			Children: []*tree.Node{
				{Path: "method", Type: "string", RawValue: "GET"},
			},
		},
	}

	done := make(chan bool, 2)

	go func() {
		for i := 0; i < 100; i++ {
			idx.Add(t1)
			idx.Lookup("GET")
			idx.LookupByShape("jwt")
			idx.LookupDecoded("get")
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			idx.Lookup("GET")
			idx.LookupByShape("jwt")
		}
		done <- true
	}()

	<-done
	<-done
}
