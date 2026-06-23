# autohttp Testing & Verification

Date: 2026-06-23

## Test Layers

### Unit Tests

Standard unit tests for Go packages and Python modules. Go uses `go test ./...`. Python uses `pytest`.

Targets:

- `internal/tree` — typed tree parser correctness.
- `internal/index` — value index normalization and lookup.
- `internal/analyze` — deterministic dependency discovery, dynamic field classification, noise filtering.
- `internal/graph` — graph construction and integrity.
- `internal/generate` — Go and Python code templates.
- `python/autohttp_worker` — browser adapter translation, endpoint matching, gRPC stream handling.
- `runtime/go` and `runtime/python` — pure HTTP replay correctness, redirect handling, JSON/HTML extraction.

### Browser Adapter Tests

Mocked browser adapter tests. The adapter interface is mocked in Go, and the Python adapter is exercised against a synthetic browser driver that emits known events. This validates the gRPC contract end to end without launching a real browser.

### Golden Fixtures

Session fixtures under `testdata/fixtures/` that exercise specific patterns:

- Login with CSRF hidden input and cookie session
- OAuth-style redirect chain with `Set-Cookie` at each hop
- Form submission with JSON body, dependency on prior JSON response
- Multi-endpoint ordered flow with incremental analysis
- Session with unresolved dynamic value requiring user-override binding

Each fixture has:

- A JSON session input
- A JSON expected `RecordedSession` output
- A JSON expected graph output
- A JSON expected unresolved binding list

### Live Recording Tests

Slow, gated, network-dependent. They:

- Launch the Python browser worker
- Drive a small workflow against a local test target
- Verify the resulting session artifact
- Verify the generated replay script runs against the same local target

These tests are tagged `live` and skipped by default. They run in CI on a schedule or on demand.

### Generated Script Tests

For each supported target (Go, Python), the generated script is run against a local test target and the final state is compared to the recording-time success condition. The test target is a small HTTP server that simulates the recorded workflow.

## Verification Commands

```bash
# Unit tests
go test ./...
go test -race ./...
pytest python/autohttp_worker/tests
pytest python/autohttp_ai/tests

# Live recording against a test target
go run ./cmd/autohttp record http://localhost:9999 \
  --browser camoufox \
  --endpoints "/login" \
  --session test-login

# Analysis
go run ./cmd/autohttp analyze --session test-login

# Generation
go run ./cmd/autohttp generate --session test-login --target go --output test-login.go

# Verification
go run ./cmd/autohttp verify --script test-login.go --target go --success-url http://localhost:9999/dashboard
```

## Test Targets

`testdata/targets/` contains small Go programs that simulate real workflows:

- `login-server` — form login with CSRF hidden input and `Set-Cookie` session.
- `oauth-server` — multi-step redirect chain with bearer token at the end.
- `json-server` — JSON request/response with nested dependency.

Each test target can run in CI on a random local port.

## CI

CI runs:

- `go fmt ./...`
- `go vet ./...`
- `go test ./...`
- `go test -race ./...`
- `make proto`
- `pytest python/autohttp_worker/tests` (adapter unit tests only, not live recording)
- `pytest python/autohttp_ai/tests`

Live recording and live verification are not part of standard CI. They run on a schedule against dedicated test targets.

## Coverage

Coverage is reported per package. Tests must exist for:

- New public Go APIs
- New Python worker behavior
- New protobuf fields
- New runtime helpers
- New CLI flags

Coverage targets are not enforced as a hard gate. The rule is: if a behavior is testable, it must be tested.
