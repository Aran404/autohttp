# AGENTS.md — autohttp

This file guides AI coding agents working on the `autohttp` codebase.

## Project Overview

`autohttp` is an open-source, hybrid Go and Python project that records browser workflows, reconstructs the functional HTTP request graph using deterministic analysis, and generates standalone Go/Python scripts that replay the workflow with fresh dynamic state.

**Core principle:** Deterministic-first. LLMs are optional ambiguity resolvers, not the main engine.

## Architecture

```
Go CLI/Orchestrator
  ↕ gRPC bidi stream (Record)
Python Browser Worker (per-recording subprocess)
  → Camoufox adapter (Firefox, default) or
  → CloakBrowser adapter (Chromium)
  → Browser events → Go normalizer → Session Model → Trees → Value Index →
    Deterministic Analyzer → Dependency Graph → Code Generator →
    Pure HTTP Replay Script (Go or Python)
                                            ↓
                                      Optional: Python AI Worker (gRPC) → g4f Providers
```

Key packages:
- `cmd/autohttp` — CLI entrypoint
- `internal/record` — Python worker subprocess lifecycle
- `internal/normalize` — Browser events to `RecordedSession`
- `internal/tree` — Typed tree parser
- `internal/index` — Value index
- `internal/analyze` — Dependency analyzer
- `internal/graph` — Executable graph
- `internal/generate` — Code generator
- `internal/verify` — Live verification runner
- `internal/inspect` — CLI inspection
- `session` — Public session types
- `runtime/go` — Go runtime for generated scripts
- `runtime/python` — Python runtime for generated scripts
- `proto/autohttp/v1/` — Protobuf contracts (includes `browser.proto` for the streaming worker contract)
- `python/autohttp_worker/` — Per-recording browser worker (Camoufox and CloakBrowser adapters)
- `python/autohttp_ai/` — Optional Python AI worker (g4f by default)

## Dev Environment

```bash
# Go version
go version go1.22+

# Python version
python3 --version    # 3.11+

# Protobuf
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate protobufs (Go and Python)
make proto

# Install Python workers
pip install -e python/autohttp_worker
pip install -e python/autohttp_ai

# Browser Python deps
pip install camoufox cloakbrowser

# Build
go build -o bin/autohttp ./cmd/autohttp

# Run
./bin/autohttp --help
```

## Testing

```bash
# Go unit tests
go test ./...
go test -race ./...

# Python worker tests
pytest python/autohttp_worker/tests
pytest python/autohttp_ai/tests

# Specific Go package
go test ./internal/analyze/... -v

# Run generated script against test target
./bin/autohttp verify --script testdata/generated/login.go --target go --success-url "https://example.com/dashboard"
```

## Coding Conventions

- Go: standard `gofmt`, `golangci-lint` if available
- Python: `ruff` for linting and formatting, `mypy --strict` for types, `pytest` for tests
- Protobuf: contract-first, Go and Python share `.proto` files
- Tests: table-driven, prefer golden fixtures in `testdata/fixtures/`
- Generated scripts: pure HTTP only, no external deps beyond stdlib + `runtime/go` or `runtime/python`
- Internal packages: not imported by generated scripts
- Generated scripts never drive a browser
- Unresolved dynamic values become user-override stub functions with explicit names (e.g. `computeHeaderXSignature`)
- Generated scripts must disable HTTP `Location` auto-follow; each redirect hop is a separate request
- Deterministic algorithms only in core path; AI is advisory
- Python worker subprocess is a per-recording lifecycle; it is not a persistent daemon
- Each recording session uses exactly one browser engine
- `OPTIONS` and `HEAD` requests are ignored for endpoint matching

## Project Structure

```
autohttp/
  cmd/autohttp/              # CLI
  internal/                  # Private Go implementation
  session/                   # Public session types
  gen/autohttp/v1/           # Generated protobuf Go bindings
  proto/autohttp/v1/         # Protobuf contracts (session, tree, analysis, graph, ai, browser)
  python/
    autohttp_worker/         # Per-recording browser worker
      adapters/
        camoufox/
        cloakbrowser/
    autohttp_ai/             # Optional Python AI worker
  runtime/
    go/                      # Go runtime for generated scripts
    python/                  # Python runtime for generated scripts
  testdata/
    fixtures/                # Golden fixtures
    targets/                 # Local test targets
  .agents/
    specs/                   # Design specs
    plans/                   # Implementation plans
    links.md                 # External references
```

## Key Commands

| Command | Purpose |
|---------|---------|
| `go build -o bin/autohttp ./cmd/autohttp` | Build CLI |
| `make proto` | Generate protobuf code for Go and Python |
| `go test ./...` | Run all Go tests |
| `go test -race ./...` | Go tests with race detector |
| `pytest python/` | Run all Python tests |
| `./bin/autohttp record <url>` | Start a recording (spawns Python worker) |
| `./bin/autohttp analyze` | Run deterministic analysis on a session |
| `./bin/autohttp generate --target go\|python` | Generate a replay script |
| `./bin/autohttp verify --script <path> --target go\|python` | Verify generated script against target |
| `./bin/autohttp inspect` | Inspect a session's captures and graph |

## Common Patterns

### Adding a new internal package
1. Create `internal/<name>/<name>.go` with types/functions
2. Add tests in `internal/<name>/<name>_test.go`
3. Export minimal API from `session/` if needed externally

### Modifying the session model
1. Update `.proto` files in `proto/autohttp/v1/`
2. Run `make proto` to regenerate Go and Python bindings
3. Update Go and Python code that uses the model

### Adding a browser event
1. Add the field to `proto/autohttp/v1/browser.proto`
2. Run `make proto`
3. Update the Python adapter to translate the browser-specific event
4. Update the Go consumer to handle the event

### Adding a test fixture
1. Add JSON to `testdata/fixtures/<name>.json`
2. Add golden expectations in `testdata/fixtures/<name>.golden.json`
3. Test loader in `internal/record/fixture_test.go`

## Dependencies

- Go: stdlib + `google.golang.org/protobuf`, `google.golang.org/grpc`
- Camoufox: Python browser automation (`camoufox` pip package, Firefox-based)
- CloakBrowser: Python browser automation (`cloakbrowser` pip package, Chromium-based)
- Python AI: isolated behind gRPC, optional
- `g4f`: only in Python AI worker, not in generated scripts

## PR Checklist

- [ ] `go fmt ./...` passes
- [ ] `go test ./...` passes
- [ ] `go test -race ./...` passes
- [ ] `make proto` works
- [ ] `pytest python/` passes
- [ ] Generated Go and Python code compiles
- [ ] No AI deps in generated runtime
- [ ] Generated scripts are pure HTTP only
- [ ] No browser-assisted replay path introduced
- [ ] Deterministic path works with `--no-ai`

## Reference Files

AI assistants should consult these files for design context and external references:

- `.agents/specs/` — Design specifications. `design.md` is the entrypoint with links to all sub-specs (architecture, data-flow, analysis, contracts, runtime, cli, testing, trust).
- `.agents/links.md` — External references (Camoufox, CloakBrowser, g4f, protobuf, gRPC, Playwright, project conventions).
- `.agents/plans/` — Implementation plans for each milestone.
- `AGENTS.md` (this file) — Project overview, conventions, and commands.
