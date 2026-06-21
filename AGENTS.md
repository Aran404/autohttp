# AGENTS.md — autohttp

This file guides AI coding agents working on the `autohttp` codebase.

## Project Overview

`autohttp` is an open-source, hybrid Go and Python project that records browser workflows (via CamoFox), reconstructs the functional HTTP request graph using deterministic analysis, and generates standalone Go/Python scripts that replay the workflow with fresh dynamic state.

**Core principle:** Deterministic-first. LLMs are optional ambiguity resolvers, not the main engine.

## Architecture

```
Go CLI/Orchestrator → CamoFox (Node.js) → Recorder → Session Model → Trees → Value Index → Deterministic Analyzer → Dependency Graph → Code Generator → Standalone Script
                          ↓
                    Optional: Python AI Worker (gRPC) → g4f Providers
```

Key packages:
- `cmd/autohttp` — CLI entrypoint
- `internal/camofox` — CamoFox process manager
- `internal/record` — Recording abstraction
- `internal/tree` — Typed tree parser
- `internal/index` — Value index
- `internal/analyze` — Dependency analyzer
- `internal/graph` — Executable graph
- `internal/generate` — Code generator
- `internal/inspect` — CLI inspection
- `pkg/session` — Public session types
- `pkg/runtime/go` — Go runtime for generated scripts
- `proto/autohttp/v1/` — Protobuf contracts

## Dev Environment

```bash
# Go version
go version go1.22+

# Protobuf
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate protobufs
make proto

# Build
go build -o bin/autohttp ./cmd/autohttp

# Run
./bin/autohttp --help
```

## Testing

```bash
# Unit tests
go test ./...

# With race detector
go test -race ./...

# Specific package
go test ./internal/analyze/... -v

# Run generated script against test target
./bin/autohttp verify --script testdata/generated/login.go --success-url "https://example.com/dashboard"
```

## Coding Conventions

- Go: standard `gofmt`, `golangci-lint` if available
- Protobuf: contract-first, Go and Python share `.proto` files
- Tests: table-driven, prefer golden fixtures in `testdata/fixtures/`
- Generated scripts: no external deps beyond stdlib + `pkg/runtime/go`
- Internal packages: not imported by generated scripts
- Deterministic algorithms only in core path; AI is advisory

## Project Structure

```
autohttp/
  cmd/autohttp/           # CLI
  internal/               # Private implementation
  pkg/                    # Public API (session, runtime)
  proto/autohttp/v1/      # Protobuf contracts
  python/autohttp_ai/     # Optional Python AI worker
  runtime/go/             # Go runtime for generated scripts
  testdata/fixtures/      # Golden fixtures
  .agents/
    specs/                # Design specs
    plans/                # Implementation plans
  testdata/
    fixtures/
```

## Key Commands

| Command | Purpose |
|---------|---------|
| `go build -o bin/autohttp ./cmd/autohttp` | Build CLI |
| `make proto` | Generate protobuf code |
| `go test ./...` | Run all tests |
| `go test -race ./...` | Race detection |
| `./bin/autohttp record <url>` | Start recording |
| `./bin/autohttp analyze` | Run deterministic analysis |
| `./bin/autohttp generate --target go --mode api` | Generate Go script |
| `./bin/autohttp verify --script <path>` | Verify generated script |

## Common Patterns

### Adding a new internal package
1. Create `internal/<name>/<name>.go` with types/functions
2. Add tests in `internal/<name>/<name>_test.go`
3. Export minimal API from `pkg/` if needed externally

### Modifying the session model
1. Update `.proto` files in `proto/autohttp/v1/`
2. Run `make proto` to regenerate Go bindings
3. Update Go code that uses the model

### Adding a test fixture
1. Add JSON to `testdata/fixtures/<name>.json`
2. Add golden expectations in `testdata/fixtures/<name>.golden.json`
4. Test loader in `internal/record/fixture_test.go`

## Dependencies

- Go: stdlib + `google.golang.org/protobuf`, `google.golang.org/grpc`
- CamoFox: external Node process (`@askjo/camofox-browser`)
- Python AI: isolated behind gRPC, optional
- `g4f`: only in Python worker, not in generated scripts

## PR Checklist

- [ ] `go fmt ./...` passes
- [ ] `go test ./...` passes
- [ ] `go test -race ./...` passes
- [ ] `make proto` works
- [ ] Generated code compiles
- [ ] No AI deps in generated runtime
- [ ] Deterministic path works with `--no-ai`

## Reference Files

AI assistants should consult these files for design context and external references:

- `.agents/specs/` — Design specifications covering architecture, data flow, analysis strategy, contracts, runtime, CLI, testing, and trust boundaries. `design.md` is the entrypoint with links to all sub-specs.
- `.agents/links.md` — External references (CamoFox, g4f, protobuf docs, project conventions).
- `.agents/plans/` — Implementation plans for each milestone.
- `AGENTS.md` (this file) — Project overview, conventions, and commands.