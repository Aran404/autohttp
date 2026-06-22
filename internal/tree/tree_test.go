package tree

import (
	"testing"

	pb "github.com/autohttp/autohttp/gen/autohttp/v1"
)

func TestParseURL(t *testing.T) {
	exchange := &pb.HttpExchange{
		Id: "test-1",
		Request: &pb.Request{
			Method: "GET",
			Url:    "https://example.com:8080/path?key=value&foo=bar",
		},
	}

	parser := New()
	trees := parser.Parse(exchange)

	if len(trees) != 1 {
		t.Fatalf("expected 1 tree, got %d", len(trees))
	}

	tree := trees[0]

	host := tree.Find("request.url.host")
	if host == nil {
		t.Fatal("expected to find request.url.host")
	}
	if host.Value != "example.com:8080" {
		t.Errorf("expected host 'example.com:8080', got %v", host.Value)
	}

	scheme := tree.Find("request.url.scheme")
	if scheme == nil {
		t.Fatal("expected to find request.url.scheme")
	}
	if scheme.Value != "https" {
		t.Errorf("expected scheme 'https', got %v", scheme.Value)
	}

	path := tree.Find("request.url.path")
	if path == nil {
		t.Fatal("expected to find request.url.path")
	}
	if path.Value != "/path" {
		t.Errorf("expected path '/path', got %v", path.Value)
	}

	key := tree.Find("request.url.query.key")
	if key == nil {
		t.Fatal("expected to find request.url.query.key")
	}
	if key.Value != "value" {
		t.Errorf("expected query value 'value', got %v", key.Value)
	}

	foo := tree.Find("request.url.query.foo")
	if foo == nil {
		t.Fatal("expected to find request.url.query.foo")
	}
	if foo.Value != "bar" {
		t.Errorf("expected query value 'bar', got %v", foo.Value)
	}
}

func TestParseHeaders(t *testing.T) {
	exchange := &pb.HttpExchange{
		Id: "test-1",
		Request: &pb.Request{
			Url: "https://example.com",
			Headers: []*pb.Header{
				{Key: "Content-Type", Value: "application/json"},
				{Key: "Authorization", Value: "Bearer token123"},
			},
		},
	}

	parser := New()
	trees := parser.Parse(exchange)
	tree := trees[0]

	ct := tree.Find("request.headers.Content-Type")
	if ct == nil {
		t.Fatal("expected to find Content-Type header")
	}
	if ct.Value != "application/json" {
		t.Errorf("expected 'application/json', got %v", ct.Value)
	}

	auth := tree.Find("request.headers.Authorization")
	if auth == nil {
		t.Fatal("expected to find Authorization header")
	}
	if auth.Value != "Bearer token123" {
		t.Errorf("expected 'Bearer token123', got %v", auth.Value)
	}
}

func TestParseJSONBody(t *testing.T) {
	body := `{"user": {"name": "John", "age": 30}, "active": true, "score": null}`
	exchange := &pb.HttpExchange{
		Id: "test-1",
		Request: &pb.Request{
			Url:      "https://example.com",
			Body:     []byte(body),
			BodyType: "json",
		},
	}

	parser := New()
	trees := parser.Parse(exchange)
	tree := trees[0]

	name := tree.Find("request.body.user.name")
	if name == nil {
		t.Fatal("expected to find request.body.user.name")
	}
	if name.Value != "John" {
		t.Errorf("expected 'John', got %v", name.Value)
	}

	age := tree.Find("request.body.user.age")
	if age == nil {
		t.Fatal("expected to find request.body.user.age")
	}
	if age.Value != float64(30) {
		t.Errorf("expected 30, got %v", age.Value)
	}

	active := tree.Find("request.body.active")
	if active == nil {
		t.Fatal("expected to find request.body.active")
	}
	if active.Value != true {
		t.Errorf("expected true, got %v", active.Value)
	}

	score := tree.Find("request.body.score")
	if score == nil {
		t.Fatal("expected to find request.body.score")
	}
	if score.Type != "null" {
		t.Errorf("expected type 'null', got %s", score.Type)
	}
}

func TestParseResponse(t *testing.T) {
	exchange := &pb.HttpExchange{
		Id: "test-1",
		Response: &pb.Response{
			Status:     200,
			StatusText: "OK",
			Headers: []*pb.Header{
				{Key: "Content-Type", Value: "text/html"},
			},
			Body:     []byte("<html>hello</html>"),
			BodyType: "text",
		},
	}

	parser := New()
	trees := parser.Parse(exchange)
	tree := trees[0]

	status := tree.Find("response.status")
	if status == nil {
		t.Fatal("expected to find response.status")
	}
	if status.Value != 200 {
		t.Errorf("expected 200, got %v", status.Value)
	}

	statusText := tree.Find("response.status_text")
	if statusText == nil {
		t.Fatal("expected to find response.status_text")
	}
	if statusText.Value != "OK" {
		t.Errorf("expected 'OK', got %v", statusText.Value)
	}

	ct := tree.Find("response.headers.Content-Type")
	if ct == nil {
		t.Fatal("expected to find Content-Type header")
	}
	if ct.Value != "text/html" {
		t.Errorf("expected 'text/html', got %v", ct.Value)
	}

	body := tree.Find("response.body")
	if body == nil {
		t.Fatal("expected to find response.body")
	}
	if body.Value != "<html>hello</html>" {
		t.Errorf("expected '<html>hello</html>', got %v", body.Value)
	}
}

func TestFindNotFound(t *testing.T) {
	exchange := &pb.HttpExchange{
		Id: "test-1",
		Request: &pb.Request{
			Url: "https://example.com",
		},
	}

	parser := New()
	trees := parser.Parse(exchange)
	tree := trees[0]

	result := tree.Find("nonexistent.path")
	if result != nil {
		t.Errorf("expected nil for nonexistent path, got %v", result)
	}

	result = tree.Find("request.nonexistent")
	if result != nil {
		t.Errorf("expected nil for nonexistent path, got %v", result)
	}
}

func TestEmptyExchange(t *testing.T) {
	parser := New()

	trees := parser.Parse(nil)
	if trees != nil {
		t.Errorf("expected nil for nil exchange, got %v", trees)
	}

	exchange := &pb.HttpExchange{Id: "test-1"}
	trees = parser.Parse(exchange)
	if len(trees) != 1 {
		t.Fatalf("expected 1 tree, got %d", len(trees))
	}

	root := trees[0].Root
	if len(root.Children) != 0 {
		t.Errorf("expected no children for empty exchange, got %d", len(root.Children))
	}
}

func TestNonJSONBody(t *testing.T) {
	exchange := &pb.HttpExchange{
		Id: "test-1",
		Request: &pb.Request{
			Url:      "https://example.com",
			Body:     []byte("plain text body"),
			BodyType: "text",
		},
	}

	parser := New()
	trees := parser.Parse(exchange)
	tree := trees[0]

	body := tree.Find("request.body")
	if body == nil {
		t.Fatal("expected to find request.body")
	}
	if body.Value != "plain text body" {
		t.Errorf("expected 'plain text body', got %v", body.Value)
	}
	if body.Type != "string" {
		t.Errorf("expected type 'string', got %s", body.Type)
	}
}

func TestJSONArrayBody(t *testing.T) {
	body := `[{"id": 1}, {"id": 2}]`
	exchange := &pb.HttpExchange{
		Id: "test-1",
		Request: &pb.Request{
			Url:      "https://example.com",
			Body:     []byte(body),
			BodyType: "json",
		},
	}

	parser := New()
	trees := parser.Parse(exchange)
	tree := trees[0]

	rootBody := tree.Find("request.body")
	if rootBody == nil {
		t.Fatal("expected to find request.body")
	}
	if rootBody.Type != "array" {
		t.Errorf("expected type 'array', got %s", rootBody.Type)
	}

	id0 := tree.Find("request.body.0.id")
	if id0 == nil {
		t.Fatal("expected to find request.body.0.id")
	}
	if id0.Value != float64(1) {
		t.Errorf("expected 1, got %v", id0.Value)
	}

	id1 := tree.Find("request.body.1.id")
	if id1 == nil {
		t.Fatal("expected to find request.body.1.id")
	}
	if id1.Value != float64(2) {
		t.Errorf("expected 2, got %v", id1.Value)
	}
}

func TestExchangeID(t *testing.T) {
	exchange := &pb.HttpExchange{
		Id: "my-exchange-id",
		Request: &pb.Request{
			Url: "https://example.com",
		},
	}

	parser := New()
	trees := parser.Parse(exchange)

	if trees[0].ExchangeID != "my-exchange-id" {
		t.Errorf("expected 'my-exchange-id', got %s", trees[0].ExchangeID)
	}
}

func TestParseMethod(t *testing.T) {
	exchange := &pb.HttpExchange{
		Id: "test-1",
		Request: &pb.Request{
			Method: "POST",
			Url:    "https://example.com",
		},
	}

	parser := New()
	trees := parser.Parse(exchange)
	tree := trees[0]

	method := tree.Find("request.method")
	if method == nil {
		t.Fatal("expected to find request.method")
	}
	if method.Value != "POST" {
		t.Errorf("expected 'POST', got %v", method.Value)
	}
}

func TestInvalidURL(t *testing.T) {
	exchange := &pb.HttpExchange{
		Id: "test-1",
		Request: &pb.Request{
			Url: "://invalid-url",
		},
	}

	parser := New()
	trees := parser.Parse(exchange)
	tree := trees[0]

	url := tree.Find("request.url")
	if url == nil {
		t.Fatal("expected to find request.url")
	}
	if url.Type != "string" {
		t.Errorf("expected type 'string' for invalid URL, got %s", url.Type)
	}
	if url.Value != "://invalid-url" {
		t.Errorf("expected raw URL value, got %v", url.Value)
	}
}

func TestInvalidJSONBody(t *testing.T) {
	exchange := &pb.HttpExchange{
		Id: "test-1",
		Request: &pb.Request{
			Url:      "https://example.com",
			Body:     []byte("{invalid json}"),
			BodyType: "json",
		},
	}

	parser := New()
	trees := parser.Parse(exchange)
	tree := trees[0]

	body := tree.Find("request.body")
	if body == nil {
		t.Fatal("expected to find request.body")
	}
	if body.Type != "string" {
		t.Errorf("expected type 'string' for invalid JSON, got %s", body.Type)
	}
}

func TestURLFragment(t *testing.T) {
	exchange := &pb.HttpExchange{
		Id: "test-1",
		Request: &pb.Request{
			Url: "https://example.com/page#section",
		},
	}

	parser := New()
	trees := parser.Parse(exchange)
	tree := trees[0]

	fragment := tree.Find("request.url.fragment")
	if fragment == nil {
		t.Fatal("expected to find request.url.fragment")
	}
	if fragment.Value != "section" {
		t.Errorf("expected 'section', got %v", fragment.Value)
	}
}
