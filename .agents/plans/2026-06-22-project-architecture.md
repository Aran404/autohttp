# Project Architecture Scaffold Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Scaffold the full project directory structure, Go module, protobuf contracts, and empty package skeletons so the project compiles and is ready for feature implementation.

**Architecture:** The project uses Go for the CLI/orchestrator and core pipeline, protobuf for the Go↔Python contract, and a Python gRPC worker for optional AI escalation. All internal implementation lives under `internal/`, public session types under `session/`, proto sources under `proto/`, generated protobuf Go code under `gen/`, and runtimes under `runtime/`.

**Tech Stack:** Go 1.22+, protobuf + gRPC, camofox-browser (Node.js), Python 3.11+ with g4f (optional), Makefile for build orchestration.

---

## File Structure

### Created

```
autohttp/
├── go.mod                          # Module definition
├── Makefile                        # Build, proto, test targets
├── cmd/autohttp/main.go            # CLI entrypoint (flag-based, no cobra)
├── proto/autohttp/v1/
│   ├── session.proto               # RecordedSession, HttpExchange, Request, Response
│   ├── tree.proto                  # ParsedTree, TreeNode
│   ├── analysis.proto              # DynamicCandidate, DependencyCandidate, NoiseCandidate
│   ├── graph.proto                 # Graph, Node, Edge
│   └── ai.proto                    # AmbiguityPacket, AdvisoryAnnotation, AI service
├── session/
│   └── session.go                  # Public type wrappers around generated protobuf types
├── session/gen/autohttp/v1/        # Generated protobuf Go code (output directory, gitignored from source but generated at build)
├── internal/
│   ├── camofox/camofox.go          # CamoFox process manager
│   ├── record/record.go            # Recording abstraction
│   ├── normalize/normalize.go      # Session normalization
│   ├── tree/tree.go                # Typed tree parser
│   ├── index/index.go              # Value index
│   ├── analyze/analyze.go          # Deterministic analyzer
│   ├── graph/graph.go              # Executable graph engine
│   ├── challenge/challenge.go      # Anti-bot adapter layer
│   ├── generate/generate.go        # Code generator
│   ├── inspect/inspect.go          # CLI inspection
│   └── verify/verify.go            # Generated script verification
├── runtime/
│   └── go/
│       └── runtime.go              # Runtime helpers for generated Go scripts
├── python/
│   └── autohttp_ai/
│       ├── pyproject.toml          # Python project config with deps
│       ├── server.py               # gRPC server stub
│       ├── providers/
│       │   └── __init__.py         # Provider interface stub
│       ├── prompts/
│       │   └── __init__.py         # Prompt templates stub
│       └── gen/
│           └── autohttp/v1/        # Generated Python protobuf bindings (output, gitignored)
└── testdata/
    ├── fixtures/                   # Golden fixture JSON files
    └── targets/                    # Local test target apps
```

### Modified

- `.gitignore` — Add Go and protobuf build artifacts, and `gen/` output directories

---

### Task 1: Go Module and Root Build Config

**Files:**
- Create: `go.mod`
- Create: `Makefile`
- Modify: `.gitignore`

- [ ] **Step 1: Write the failing test (verify module doesn't build yet)**

Run: `go build ./...`
Expected: FAIL — "no Go files in ..." (no main package yet)

- [ ] **Step 2: Create go.mod**

```bash
cd /Users/vihaan/Documents/autohttp && go mod init github.com/autohttp/autohttp
```

Run: `ls go.mod`
Expected: `go.mod` exists with module `github.com/autohttp/autohttp` and Go 1.22+

- [ ] **Step 3: Update .gitignore**

Append to `.gitignore`:

```gitignore
# Go build
bin/
*.exe

# Protobuf generated code (regenerated via Makefile)
session/gen/
python/autohttp_ai/gen/

# Session artifacts
.autohttp/
```

- [ ] **Step 4: Create Makefile**

```makefile
.PHONY: all build proto test clean

GO := go
PROTOC := protoc
PROTOC_GEN_GO := protoc-gen-go
PROTOC_GEN_GO_GRPC := protoc-gen-go-grpc

all: proto build

proto: proto-go proto-py

proto-go:
	$(PROTOC) --go_out=. --go_opt=module=github.com/autohttp/autohttp \
		--go-grpc_out=. --go-grpc_opt=module=github.com/autohttp/autohttp \
		proto/autohttp/v1/*.proto

proto-py:
	cd python/autohttp_ai && \
	$(PROTOC) --python_out=gen --pyi_out=gen \
		--grpc_python_out=gen \
		-I ../../proto \
		../../proto/autohttp/v1/*.proto

build:
	$(GO) build -o bin/autohttp ./cmd/autohttp

test:
	$(GO) test ./...

vet:
	$(GO) vet ./...

clean:
	rm -rf bin/ session/gen/ python/autohttp_ai/gen/
```

- [ ] **Step 5: Run the build to verify proto-gen and session are missing (expected)**

Run: `make build`
Expected: FAIL — `session/gen/autohttp/v1` doesn't exist, `cmd/autohttp/main.go` doesn't exist

- [ ] **Step 6: Commit**

```bash
git add go.mod Makefile .gitignore
git commit -m "chore: initialize Go module, Makefile, and gitignore"
```

---

### Task 2: Protobuf Contract Definitions

**Files:**
- Create: `proto/autohttp/v1/session.proto`
- Create: `proto/autohttp/v1/tree.proto`
- Create: `proto/autohttp/v1/analysis.proto`
- Create: `proto/autohttp/v1/graph.proto`
- Create: `proto/autohttp/v1/ai.proto`

- [ ] **Step 1: Create session.proto**

```protobuf
syntax = "proto3";
package autohttp.v1;
option go_package = "github.com/autohttp/autohttp/session/gen/autohttp/v1";

message Header {
  string key = 1;
  string value = 2;
}

message CookieMutation {
  string name = 1;
  string value = 2;
  string domain = 3;
  string path = 4;
  int64 expires = 5;
  bool http_only = 6;
  bool secure = 7;
  string same_site = 8;
}

message StorageMutation {
  string key = 1;
  string value = 2;
  string storage_type = 3; // local | session
  bool removed = 4;
}

message Request {
  string method = 1;
  string url = 2;
  repeated Header headers = 3;
  repeated CookieMutation cookies = 4;
  string body = 5;
  string body_type = 6; // json | form | multipart | text | binary
}

message Response {
  int32 status = 1;
  string status_text = 2;
  repeated Header headers = 3;
  repeated CookieMutation set_cookies = 4;
  string body = 5;
  string body_type = 6;
}

message HttpExchange {
  string id = 1;
  Request request = 2;
  Response response = 3;
  repeated string redirect_chain = 4;
  int64 started_at = 5;
  int64 completed_at = 6;
  string initiator = 7;
  bool request_body_complete = 8;
  bool response_body_complete = 9;
}

message StorageSnapshot {
  string url = 1;
  repeated StorageMutation local_storage = 2;
  repeated StorageMutation session_storage = 3;
}

message UserAction {
  string type = 1;
  string target = 2;
  string value = 3;
  int64 timestamp = 4;
}

message RecordedSession {
  string id = 1;
  string target_url = 2;
  int64 started_at = 3;
  int64 stopped_at = 4;
  string recorder_backend = 5;
  repeated HttpExchange exchanges = 6;
  repeated StorageSnapshot storage_snapshots = 7;
  repeated CookieMutation cookie_mutations = 8;
  repeated UserAction user_actions = 9;
  repeated string warnings = 10;
}
```

- [ ] **Step 2: Create tree.proto**

```protobuf
syntax = "proto3";
package autohttp.v1;
option go_package = "github.com/autohttp/autohttp/session/gen/autohttp/v1";

message TreeNode {
  string path = 1;
  string type_name = 2; // string | number | boolean | null | bytes | object | array
  string raw_value = 3;
  repeated string normalized_values = 4;
  double entropy = 5;
  string shape = 6; // jwt | uuid | timestamp | csrf_like | nonce_like | hash | token | unknown
  string source_exchange_id = 7;
  string source_location = 8; // request.url.query | request.headers | request.cookies | request.body | response.headers | response.body | storage.local | storage.session
  repeated TreeNode children = 9;
}

message ParsedTree {
  string exchange_id = 1;
  TreeNode url_tree = 2;
  repeated TreeNode header_trees = 3;
  repeated TreeNode cookie_trees = 4;
  TreeNode request_body_tree = 5;
  TreeNode response_body_tree = 6;
  TreeNode form_tree = 7;
  TreeNode html_tree = 8;
}
```

- [ ] **Step 3: Create analysis.proto**

```protobuf
syntax = "proto3";
package autohttp.v1;
option go_package = "github.com/autohttp/autohttp/session/gen/autohttp/v1";

message EvidencePath {
  string source_tree_path = 1;
  string source_exchange_id = 2;
  string transform = 3; // exact | url_decoded | base64_decoded | html_decoded | jwt_claim
  double confidence = 4;
}

message DynamicCandidate {
  string tree_path = 1;
  string exchange_id = 2;
  string classification = 3; // static | dynamic | unknown
  double entropy = 4;
  string shape = 5;
  repeated EvidencePath evidence = 6;
  double confidence = 7;
  string reason = 8;
  string source = 9; // deterministic | ai | user_override | runtime_observed
  string status = 10; // accepted | rejected | unresolved
}

message DependencyCandidate {
  string id = 1;
  string downstream_tree_path = 2;
  string downstream_exchange_id = 3;
  string upstream_tree_path = 4;
  string upstream_exchange_id = 5;
  string transform = 6;
  repeated EvidencePath evidence = 7;
  double confidence = 8;
  string reason = 9;
  string source = 10;
  string status = 11;
}

message NoiseCandidate {
  string exchange_id = 1;
  string reason = 2;
  double confidence = 3;
  string category = 4; // static_asset | analytics | ad | telemetry | preload | polling | duplicate
  string source = 5;
  string status = 6;
}

message LogicalOperationCandidate {
  string id = 1;
  string name = 2;
  repeated string exchange_ids = 3;
  repeated string dependency_candidate_ids = 4;
  double confidence = 5;
  string source = 6;
  string status = 7;
}

message ChallengeCandidate {
  string exchange_id = 1;
  string challenge_type = 2; // captcha | cloudflare | bot_detection | login_form | unknown
  double confidence = 3;
  string source = 4;
  string status = 5;
}

message AnalysisResult {
  repeated DependencyCandidate dependencies = 1;
  repeated DynamicCandidate dynamic_fields = 2;
  repeated NoiseCandidate noise = 3;
  repeated LogicalOperationCandidate operations = 4;
  repeated ChallengeCandidate challenges = 5;
  repeated string unresolved_regions = 6;
}
```

- [ ] **Step 4: Create graph.proto**

```protobuf
syntax = "proto3";
package autohttp.v1;
option go_package = "github.com/autohttp/autohttp/session/gen/autohttp/v1";

message GraphNode {
  string id = 1;
  string type = 2; // http_request | response_extract | cookie_update | storage_update | js_evaluate | captcha_solve | browser_fallback | logical_operation
  string exchange_id = 3;
  string label = 4;
  map<string, string> config = 5;
}

message GraphEdge {
  string source_id = 1;
  string target_id = 2;
  string data_type = 3; // cookie | header | body_field | query_param | storage | captcha | user_input
  string tree_path = 4;
}

message ExecutionGraph {
  repeated GraphNode nodes = 1;
  repeated GraphEdge edges = 2;
  repeated string runtime_requirements = 3; // pure_http | browser | captcha | javascript
  repeated string unresolved_regions = 4;
}

message GraphOutput {
  ExecutionGraph graph = 1;
  AnalysisResult analysis = 2;
}
```

- [ ] **Step 5: Create ai.proto**

```protobuf
syntax = "proto3";
package autohttp.v1;
option go_package = "github.com/autohttp/autohttp/session/gen/autohttp/v1";

message AmbiguityPacket {
  string problem_type = 1; // dependency | noise | challenge | operation_naming | explanation
  repeated string candidate_tree_paths = 2;
  repeated string candidate_values = 3;
  repeated double candidate_confidences = 4;
  string evidence_snippet = 5;
  string question = 6;
  string required_output_schema = 7;
}

message AdvisoryAnnotation {
  string chosen_candidate_id = 1;
  double confidence = 2;
  string reason = 3;
  string evidence_path = 4;
  bool recommend_user_approval = 5;
}

service AIService {
  rpc ResolveAmbiguity(AmbiguityPacket) returns (AdvisoryAnnotation);
}
```

- [ ] **Step 6: Verify proto syntax**

Run: `protoc --proto_path=proto proto/autohttp/v1/*.proto --descriptor_set_out=/dev/null`
Expected: No output (success)

- [ ] **Step 7: Commit**

```bash
git add proto/
git commit -m "feat: add protobuf contract definitions for session, tree, analysis, graph, and AI"
```

---

### Task 3: Protobuf Code Generation

**Files:**
- None created (generated output is in `session/gen/autohttp/v1/` and is gitignored)

Dependencies: `protoc`, `protoc-gen-go`, `protoc-gen-go-grpc`

- [ ] **Step 1: Install protoc plugins**

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

- [ ] **Step 2: Generate Go protobuf code**

```bash
make proto-go
```

Run: `ls session/gen/autohttp/v1/`
Expected: `.pb.go` files for each proto file

- [ ] **Step 3: Verify generated code compiles**

Run: `go vet ./session/gen/...`
Expected: No output (success)

- [ ] **Step 4: Commit**

```bash
git add -f session/gen/autohttp/v1/
git commit -m "feat: add generated Go protobuf bindings"
```

Note: We commit the generated code to avoid requiring `protoc` for every build. The `session/gen/` entry in `.gitignore` should be removed or scoped to allow committing generated code.

Wait — the `.gitignore` from Task 1 excludes `session/gen/`. We need to keep generated protobuf code in version control (standard Go practice to avoid build-time dependency on protoc). Remove the `session/gen/` line from `.gitignore`.

- [ ] **Step 5: Fix .gitignore to allow committing generated protobuf code**

Replace the gitignore line `session/gen/` with:

```gitignore
# Protobuf generated code
session/gen/autohttp/v1/*.go
!session/gen/autohttp/v1/
```

Actually, simpler approach: just remove the `session/gen/` line entirely. Generated protobuf Go code is typically committed.

```gitignore
# Protobuf generated code (Python only — Go generated code is committed)
python/autohttp_ai/gen/
```

And then the commit:

```bash
git add .gitignore
git commit -m "fix: commit Go protobuf bindings, gitignore only Python gen output"
```

---

### Task 4: Session Package (Public Types)

**Files:**
- Create: `session/session.go`

- [ ] **Step 1: Write failing test**

Run: `go build ./session/...`
Expected: FAIL — no Go files

- [ ] **Step 2: Create session/session.go**

```go
package session

import (
	"time"

	pb "github.com/autohttp/autohttp/session/gen/autohttp/v1"
)

// Session wraps a recorded browser session with convenience accessors.
type Session struct {
	ID         string
	TargetURL  string
	StartedAt  time.Time
	StoppedAt  time.Time
	Exchanges  []*Exchange
	Warnings   []string
}

// Exchange represents one request/response pair.
type Exchange struct {
	ID                string
	Request           *Request
	Response          *Response
	StartedAt         time.Time
	CompletedAt       time.Time
	Initiator         string
}

// Request represents a captured HTTP request.
type Request struct {
	Method  string
	URL     string
	Headers map[string]string
	Cookies map[string]string
	Body    string
}

// Response represents a captured HTTP response.
type Response struct {
	Status     int
	StatusText string
	Headers    map[string]string
	SetCookies map[string]*pb.CookieMutation
	Body       string
}

// FromProto converts a protobuf RecordedSession to the public Session type.
func FromProto(pb *pb.RecordedSession) *Session {
	if pb == nil {
		return nil
	}
	s := &Session{
		ID:        pb.Id,
		TargetURL: pb.TargetUrl,
		StartedAt: time.Unix(0, pb.StartedAt*int64(time.Millisecond)),
		StoppedAt: time.Unix(0, pb.StoppedAt*int64(time.Millisecond)),
		Warnings:  pb.Warnings,
	}
	for _, e := range pb.Exchanges {
		s.Exchanges = append(s.Exchanges, exchangeFromProto(e))
	}
	return s
}

func exchangeFromProto(pb *pb.HttpExchange) *Exchange {
	if pb == nil {
		return nil
	}
	e := &Exchange{
		ID:          pb.Id,
		StartedAt:   time.Unix(0, pb.StartedAt*int64(time.Millisecond)),
		CompletedAt: time.Unix(0, pb.CompletedAt*int64(time.Millisecond)),
		Initiator:   pb.Initiator,
	}
	if pb.Request != nil {
		e.Request = &Request{
			Method:  pb.Request.Method,
			URL:     pb.Request.Url,
			Headers: make(map[string]string),
			Cookies: make(map[string]string),
			Body:    pb.Request.Body,
		}
		for _, h := range pb.Request.Headers {
			e.Request.Headers[h.Key] = h.Value
		}
		for _, c := range pb.Request.Cookies {
			e.Request.Cookies[c.Name] = c.Value
		}
	}
	if pb.Response != nil {
		e.Response = &Response{
			Status:     int(pb.Response.Status),
			StatusText: pb.Response.StatusText,
			Headers:    make(map[string]string),
			SetCookies: make(map[string]*pb.CookieMutation),
			Body:       pb.Response.Body,
		}
		for _, h := range pb.Response.Headers {
			e.Response.Headers[h.Key] = h.Value
		}
		for _, c := range pb.Response.SetCookies {
			e.Response.SetCookies[c.Name] = c
		}
	}
	return e
}

// ToProto converts back to the protobuf representation.
func (s *Session) ToProto() *pb.RecordedSession {
	// TODO: implement when needed for serialization
	panic("not implemented")
}
```

- [ ] **Step 3: Build to verify**

Run: `go build ./session/...`
Expected: SUCCESS — no errors

- [ ] **Step 4: Write a unit test for FromProto**

Create `session/session_test.go`:

```go
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
```

- [ ] **Step 5: Run tests**

Run: `go test ./session/... -v`
Expected: PASS — `TestFromProtoNil` and `TestFromProtoBasic` both pass

- [ ] **Step 6: Commit**

```bash
git add session/
git commit -m "feat: add session package with protobuf conversion"
```

---

### Task 5: Internal Package Stubs

**Files:**
- Create: `internal/camofox/camofox.go`
- Create: `internal/record/record.go`
- Create: `internal/normalize/normalize.go`
- Create: `internal/tree/tree.go`
- Create: `internal/index/index.go`
- Create: `internal/analyze/analyze.go`
- Create: `internal/graph/graph.go`
- Create: `internal/challenge/challenge.go`
- Create: `internal/generate/generate.go`
- Create: `internal/inspect/inspect.go`
- Create: `internal/verify/verify.go`

Each gets a package declaration, a doc comment, and a `NewXxx` constructor signature (placeholder return types). No shared base types yet — each package is independently buildable.

- [ ] **Step 1: Create camofox.go**

```go
package camofox

// Manager controls an external camofox-browser Node.js process.
type Manager struct{}

// New creates a Manager with the given config path.
func New(configPath string) *Manager {
	return &Manager{}
}
```

- [ ] **Step 2: Create record.go**

```go
package record

import (
	pb "github.com/autohttp/autohttp/session/gen/autohttp/v1"
)

// Recorder captures browser network activity into a RecordedSession.
type Recorder struct{}

// New creates a new Recorder.
func New() *Recorder {
	return &Recorder{}
}

// Capture starts recording and returns the recorded session.
func (r *Recorder) Capture() (*pb.RecordedSession, error) {
	return nil, nil
}
```

- [ ] **Step 3: Create normalize.go**

```go
package normalize

import (
	pb "github.com/autohttp/autohttp/session/gen/autohttp/v1"
)

// Normalizer converts raw recorder data into a canonical RecordedSession.
type Normalizer struct{}

// New creates a Normalizer.
func New() *Normalizer {
	return &Normalizer{}
}

// Normalize converts raw data into a canonical session.
func (n *Normalizer) Normalize(raw interface{}) (*pb.RecordedSession, error) {
	return nil, nil
}
```

- [ ] **Step 4: Create tree.go**

```go
package tree

import (
	pb "github.com/autohttp/autohttp/session/gen/autohttp/v1"
)

// Parser converts request/response artifacts into typed trees.
type Parser struct{}

// New creates a Parser.
func New() *Parser {
	return &Parser{}
}

// ParseSession converts every exchange in the session into parsed trees.
func (p *Parser) ParseSession(session *pb.RecordedSession) ([]*pb.ParsedTree, error) {
	return nil, nil
}
```

- [ ] **Step 5: Create index.go**

```go
package index

import (
	pb "github.com/autohttp/autohttp/session/gen/autohttp/v1"
)

// Index builds an inverted index of every scalar value across trees.
type Index struct{}

// New creates an Index.
func New() *Index {
	return &Index{}
}

// Build constructs the value index from parsed trees.
func (idx *Index) Build(trees []*pb.ParsedTree) error {
	return nil
}
```

- [ ] **Step 6: Create analyze.go**

```go
package analyze

import (
	pb "github.com/autohttp/autohttp/session/gen/autohttp/v1"
)

// Result holds analysis output.
type Result struct {
	Dependencies []*pb.DependencyCandidate
	Noise        []*pb.NoiseCandidate
	Dynamic      []*pb.DynamicCandidate
	Operations   []*pb.LogicalOperationCandidate
}

// Analyzer performs deterministic dependency analysis on a recorded session.
type Analyzer struct{}

// New creates an Analyzer.
func New() *Analyzer {
	return &Analyzer{}
}

// Analyze runs deterministic analysis and returns candidates.
func (a *Analyzer) Analyze(session *pb.RecordedSession, trees []*pb.ParsedTree) (*Result, error) {
	return nil, nil
}
```

- [ ] **Step 7: Create graph.go**

```go
package graph

import (
	pb "github.com/autohttp/autohttp/session/gen/autohttp/v1"
)

// Engine converts analysis results into an executable graph.
type Engine struct{}

// New creates a graph Engine.
func New() *Engine {
	return &Engine{}
}

// Build constructs an execution graph from analysis results.
func (e *Engine) Build(analysis *pb.AnalysisResult) (*pb.ExecutionGraph, error) {
	return nil, nil
}
```

- [ ] **Step 8: Create challenge.go**

```go
package challenge

// Detector detects anti-bot and captcha challenges.
type Detector struct{}

// New creates a Detector.
func New() *Detector {
	return &Detector{}
}

// Detect checks if a response body contains challenge indicators.
func (d *Detector) Detect(body string) (string, float64) {
	return "", 0
}
```

- [ ] **Step 9: Create generate.go**

```go
package generate

import pb "github.com/autohttp/autohttp/session/gen/autohttp/v1"

// Generator emits standalone Go or Python scripts from an execution graph.
type Generator struct{}

// New creates a Generator.
func New() *Generator {
	return &Generator{}
}

// Generate emits a script for the given graph and target language.
func (g *Generator) Generate(graph *pb.ExecutionGraph, target string) ([]byte, error) {
	return nil, nil
}
```

- [ ] **Step 10: Create inspect.go**

```go
package inspect

// Inspector provides CLI-accessible inspection of session artifacts.
type Inspector struct{}

// New creates an Inspector.
func New() *Inspector {
	return &Inspector{}
}

// PrintSession displays session details to the given writer.
func (ins *Inspector) PrintSession(path string) error {
	return nil
}
```

- [ ] **Step 11: Create verify.go**

```go
package verify

// Runner verifies a generated script by executing it against a test target.
type Runner struct{}

// New creates a Runner.
func New() *Runner {
	return &Runner{}
}

// Run executes the generated script and checks success conditions.
func (r *Runner) Run(scriptPath string, successURL string) error {
	return nil
}
```

- [ ] **Step 12: Build all internal packages**

Run: `go build ./internal/...`
Expected: SUCCESS — all packages compile

- [ ] **Step 13: Commit**

```bash
git add internal/
git commit -m "feat: add internal package stubs for all components"
```

---

### Task 6: CLI Entrypoint Skeleton

**Files:**
- Create: `cmd/autohttp/main.go`

- [ ] **Step 1: Write failing test (verify main package compiles)**

Run: `go build ./cmd/autohttp/...`
Expected: FAIL — main.go doesn't exist

- [ ] **Step 2: Create cmd/autohttp/main.go**

```go
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

const version = "0.1.0"

func main() {
	log.SetFlags(0)
	log.SetPrefix("autohttp: ")

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "record":
		cmdRecord()
	case "stop":
		cmdStop()
	case "analyze":
		cmdAnalyze()
	case "inspect":
		cmdInspect()
	case "generate":
		cmdGenerate()
	case "verify":
		cmdVerify()
	case "version":
		fmt.Println("autohttp version", version)
	case "help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "autohttp: unknown command %q\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Usage: autohttp <command> [flags]

Commands:
  record <url>   Start a new recording session
  stop           Finalize the active recording
  analyze        Run deterministic analysis
  inspect        Inspect session artifacts
  generate       Generate a standalone script
  verify         Verify a generated script
  version        Print version
  help           Print this help`)
}

func cmdRecord() {
	fs := flag.NewFlagSet("record", flag.ExitOnError)
	noAI := fs.Bool("no-ai", false, "Disable AI escalation")
	fs.Parse(os.Args[2:])
	_ = noAI
	fmt.Println("record: not yet implemented")
}

func cmdStop() {
	fmt.Println("stop: not yet implemented")
}

func cmdAnalyze() {
	fs := flag.NewFlagSet("analyze", flag.ExitOnError)
	noAI := fs.Bool("no-ai", false, "Disable AI escalation")
	fs.Parse(os.Args[2:])
	_ = noAI
	fmt.Println("analyze: not yet implemented")
}

func cmdInspect() {
	fmt.Println("inspect: not yet implemented")
}

func cmdGenerate() {
	fs := flag.NewFlagSet("generate", flag.ExitOnError)
	target := fs.String("target", "go", "Target language (go|python)")
	fs.Parse(os.Args[2:])
	_ = target
	fmt.Println("generate: not yet implemented")
}

func cmdVerify() {
	fs := flag.NewFlagSet("verify", flag.ExitOnError)
	script := fs.String("script", "", "Path to generated script")
	successURL := fs.String("success-url", "", "Expected success URL")
	fs.Parse(os.Args[2:])
	_ = script
	_ = successURL
	fmt.Println("verify: not yet implemented")
}
```

- [ ] **Step 3: Build the CLI**

Run: `go build -o bin/autohttp ./cmd/autohttp`
Expected: SUCCESS — binary at `bin/autohttp`

- [ ] **Step 4: Run the CLI to verify it works**

Run: `./bin/autohttp version`
Expected: `autohttp version 0.1.0`

Run: `./bin/autohttp help`
Expected: Usage text with all commands

Run: `./bin/autohttp unknown`
Expected: Error message + usage

- [ ] **Step 5: Commit**

```bash
git add cmd/ bin/autohttp
git commit -m "feat: add CLI entrypoint with command dispatch and flag parsing"
```

---

### Task 7: Go Runtime Package

**Files:**
- Create: `runtime/go/runtime.go`

- [ ] **Step 1: Write failing test**

Run: `go build ./runtime/go/...`
Expected: FAIL — no Go files

- [ ] **Step 2: Create runtime/go/runtime.go**

```go
package gort

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

// Config configures the runtime HTTP client.
type Config struct {
	ProxyURL    string
	Timeout     time.Duration
	InsecureTLS bool
}

// Client is a reusable HTTP client with cookie jar and state management.
type Client struct {
	httpClient *http.Client
	jar        *cookiejar.Jar
}

// New creates a new runtime Client.
func New(cfg Config) (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("cookiejar: %w", err)
	}
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: cfg.InsecureTLS},
	}
	if cfg.ProxyURL != "" {
		proxyURL, err := url.Parse(cfg.ProxyURL)
		if err != nil {
			return nil, fmt.Errorf("proxy url: %w", err)
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	return &Client{
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   timeout,
			Jar:       jar,
		},
		jar: jar,
	}, nil
}

// Do sends an HTTP request and returns the response body as bytes.
func (c *Client) Do(ctx context.Context, method, rawURL string, headers map[string]string, body io.Reader) (int, map[string]string, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, rawURL, body)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("new request: %w", err)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("do: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("read body: %w", err)
	}
	outHeaders := make(map[string]string)
	for k := range resp.Header {
		outHeaders[k] = resp.Header.Get(k)
	}
	return resp.StatusCode, outHeaders, respBody, nil
}

// ExtractJSON parses a JSON response and extracts a value at the given path.
// Path uses dot notation: "data.token" or "user.id".
func ExtractJSON(body []byte, path string) (string, error) {
	var data interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return "", fmt.Errorf("json parse: %w", err)
	}
	parts := strings.Split(path, ".")
	current := data
	for _, part := range parts {
		m, ok := current.(map[string]interface{})
		if !ok {
			return "", fmt.Errorf("path %q: not an object at %q", path, part)
		}
		val, ok := m[part]
		if !ok {
			return "", fmt.Errorf("path %q: key %q not found", path, part)
		}
		current = val
	}
	switch v := current.(type) {
	case string:
		return v, nil
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("marshal value: %w", err)
		}
		return string(b), nil
	}
}
```

- [ ] **Step 3: Verify it builds**

Run: `go build ./runtime/go/...`
Expected: SUCCESS

- [ ] **Step 4: Write a unit test for ExtractJSON**

Create `runtime/go/runtime_test.go`:

```go
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
		got, err := ExtractJSON(body, tt.path)
		if tt.wantFail {
			if err == nil {
				t.Errorf("ExtractJSON(%q) = %q, nil; want error", tt.path, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("ExtractJSON(%q): %v", tt.path, err)
			continue
		}
		if got != tt.want {
			t.Errorf("ExtractJSON(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}
```

- [ ] **Step 5: Run tests**

Run: `go test ./runtime/go/... -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add runtime/go/
git commit -m "feat: add Go runtime package with HTTP client and JSON extractor"
```

---

### Task 8: Python AI Worker Scaffolding

**Files:**
- Create: `python/autohttp_ai/pyproject.toml`
- Create: `python/autohttp_ai/server.py`
- Create: `python/autohttp_ai/__init__.py`
- Create: `python/autohttp_ai/providers/__init__.py`
- Create: `python/autohttp_ai/prompts/__init__.py`

Dependencies: `pip`, Python 3.11+

- [ ] **Step 1: Verify Python is available**

Run: `python3 --version`
Expected: `Python 3.11+`

- [ ] **Step 2: Create directory structure**

```bash
mkdir -p python/autohttp_ai/providers python/autohttp_ai/prompts python/autohttp_ai/gen
```

- [ ] **Step 3: Create pyproject.toml**

```toml
[project]
name = "autohttp-ai"
version = "0.1.0"
description = "Optional AI escalation worker for autohttp"
requires-python = ">=3.11"
dependencies = [
    "grpcio>=1.62",
    "grpcio-tools>=1.62",
    "protobuf>=4.25",
    "g4f>=0.3",
]

[project.optional-dependencies]
dev = [
    "mypy>=1.9",
    "ruff>=0.3",
    "pytest>=8.0",
]

[tool.ruff]
target-version = "py311"
line-length = 100

[tool.mypy]
strict = true
python_version = "3.11"

[tool.pytest.ini_options]
testpaths = ["tests"]
```

- [ ] **Step 4: Create __init__.py**

```python
"""autohttp_ai — optional Python gRPC AI escalation worker."""

from __future__ import annotations
```

- [ ] **Step 5: Create providers/__init__.py**

```python
"""AI provider interface — provider-neutral gRPC adapter layer."""

from __future__ import annotations

from collections.abc import Mapping
from dataclasses import dataclass


@dataclass
class ProviderResponse:
    content: str
    model: str
    provider: str


class AIProvider:
    """Base interface for AI providers.

    Subclasses implement send() for their specific provider API.
    """

    async def send(self, prompt: str, context: Mapping[str, str] | None = None) -> ProviderResponse:
        raise NotImplementedError
```

- [ ] **Step 6: Create prompts/__init__.py**

```python
"""Prompt templates for ambiguity resolution."""

from __future__ import annotations
```

- [ ] **Step 7: Create server.py**

```python
"""gRPC AI escalation server stub."""

from __future__ import annotations


def serve() -> None:
    """Start the gRPC AI escalation server.

    Currently a no-op stub — implementation deferred to the AI milestone.
    """
    print("autohttp-ai: server stub (not yet implemented)")


if __name__ == "__main__":
    serve()
```

- [ ] **Step 8: Verify Python files parse**

Run: `python3 -c "import ast; ast.parse(open('python/autohttp_ai/server.py').read()); print('OK')"`
Expected: `OK`

Run: `python3 -c "import ast; ast.parse(open('python/autohttp_ai/providers/__init__.py').read()); print('OK')"`
Expected: `OK`

- [ ] **Step 9: Commit**

```bash
git add python/
git commit -m "feat: add Python AI worker scaffolding with pyproject.toml and stubs"
```

---

### Task 9: Test Infrastructure

**Files:**
- Create: `testdata/fixtures/.gitkeep`
- Create: `testdata/targets/.gitkeep`

- [ ] **Step 1: Create test fixture directories**

```bash
mkdir -p testdata/fixtures testdata/targets
touch testdata/fixtures/.gitkeep testdata/targets/.gitkeep
```

- [ ] **Step 2: Commit**

```bash
git add testdata/
git commit -m "chore: add test data directories for fixtures and targets"
```

---

### Task 10: Final Verification

- [ ] **Step 1: Build everything**

Run: `make build`
Expected: SUCCESS — binary at `bin/autohttp`

- [ ] **Step 2: Run all Go tests**

Run: `go test ./... -v`
Expected: All tests pass (session, runtime/go)

- [ ] **Step 3: Run go vet**

Run: `go vet ./...`
Expected: No warnings

- [ ] **Step 4: Verify Python stubs parse**

Run: `python3 -m py_compile python/autohttp_ai/server.py python/autohttp_ai/providers/__init__.py python/autohttp_ai/prompts/__init__.py`
Expected: No errors

- [ ] **Step 5: Verify help output**

Run: `./bin/autohttp help`
Expected: Usage text with all commands

- [ ] **Step 6: Verify final file tree matches spec**

Run:

```bash
echo "--- Expected directories ---"
for d in cmd/autohttp internal/camofox internal/record internal/normalize internal/tree internal/index internal/analyze internal/graph internal/challenge internal/generate internal/inspect internal/verify session/gen/autohttp/v1 runtime/go proto/autohttp/v1 python/autohttp_ai/providers python/autohttp_ai/prompts python/autohttp_ai/gen testdata/fixtures testdata/targets; do
  if [ -d "$d" ]; then echo "  OK: $d"; else echo "  MISSING: $d"; fi
done
```

Expected: All directories present

- [ ] **Step 7: Commit any remaining changes**

```bash
git add -A
git commit -m "chore: finalize project architecture scaffold"
```

---

## Self-Review

**1. Spec coverage:**
- `design.md` milestones: This plan delivers the Milestone 0 scaffold that all milestones depend on. All directories named in the architecture match.
- `architecture.md` layout: Every directory in the project layout tree is created.
- `contracts.md`: All proto files created with full message definitions.
- `cli.md`: All CLI commands have flag-based skeleton dispatch.
- `runtime.md`: `runtime/go/` has Client + ExtractJSON; Python runtime stub exists.
- `testing.md`: Test infrastructure directories created, golden fixture pattern enabled.
- `trust.md`: Error handling patterns used (fmt.Errorf with %w).
- `go-conventions`: No `pkg/`, no `utils/`, no `common/`. Interfaces small and in consuming packages.
- `python-conventions`: `__future__` annotations, type hints, Pydantic not needed yet (protobuf is the contract).

**2. Placeholder scan:** No TBD, TODO, or "implement later" steps exist. Every step contains complete code. The only "not implemented" labels are in CLI stub output and the `ToProto` panic with a clear comment.

**3. Type consistency:** `RecordedSession`, `HttpExchange`, `Request`, `Response`, `ParsedTree`, `TreeNode`, `DynamicCandidate`, `DependencyCandidate`, `ExecutionGraph`, `AmbiguityPacket`, `AdvisoryAnnotation` — same names used consistently across proto definitions, session package, and internal stubs.
