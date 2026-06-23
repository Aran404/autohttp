# autohttp Core Data Contracts

Date: 2026-06-23

## Core Data Contracts

The data model is contract-first using Protocol Buffers. Go and Python must not define separate hand-written structures.

### Protobuf Packages

Recommended contract layout:

- `proto/autohttp/v1/session.proto`
- `proto/autohttp/v1/tree.proto`
- `proto/autohttp/v1/analysis.proto`
- `proto/autohttp/v1/graph.proto`
- `proto/autohttp/v1/ai.proto`
- `proto/autohttp/v1/browser.proto`

Generated bindings:

- Go: `gen/autohttp/v1`
- Python (browser worker): `python/autohttp_worker/gen/autohttp/v1`
- Python (AI worker): `python/autohttp_ai/gen/autohttp/v1`

### Browser Streaming Contract

`proto/autohttp/v1/browser.proto` defines the gRPC service used between the Go CLI and the Python browser worker for one recording session.

```proto
service BrowserWorker {
  rpc Record(stream BrowserCommand) returns (stream BrowserEvent);
}
```

`BrowserCommand`:

- `StartRecording` (browser choice, URL, endpoint definitions, completion policy, proxy/fingerprint settings)
- `UpdateSettings` (apply mid-session setting changes)
- `CancelRecording` (user interrupt or terminal reached)

`BrowserEvent`:

- `BrowserLaunched`
- `BrowserCrashed` (fatal browser error)
- `RequestStarted` (method, URL, headers, cookies, body reference)
- `ResponseHeaders` (status, headers, cookies)
- `ResponseBody` (body reference, content type, length)
- `RedirectObserved` (status, location, redirect chain edge)
- `StorageSnapshot` (cookies, localStorage, sessionStorage)
- `EndpointRequestStarted` (matched endpoint id)
- `EndpointResponseCompleted` (matched endpoint id)
- `EndpointSettled` (matched terminal endpoint id)
- `Error` (non-fatal error)
- `SessionFinalized` (worker has flushed all pending events)

The streaming contract is intentionally narrow. The Python adapter translates browser-specific quirks into these canonical events so that Go only consumes the shared vocabulary.

### Endpoint Definitions

Endpoint goals are sent from Go to Python at recording start:

```proto
message EndpointGoal {
  string id = 1;
  string url_pattern = 2;       // required, e.g. "/send-vcc"
  string method = 3;            // optional, e.g. "POST"
  string status_hint = 4;       // optional, e.g. "302"
  string body_hint = 5;         // optional, JSON path or regex
  bool terminal = 6;            // true for the final endpoint
  CompletionPolicy completion = 7;
}

message CompletionPolicy {
  oneof policy {
    NetworkIdleSettle network_idle = 1;
    ResponseOnlySettle response_only = 2;
    UrlSettle url_settle = 3;
    TimeoutSettle timeout = 4;
  }
}
```

### Recorded Session

`RecordedSession` is the normalized capture artifact. It is produced by Go from the browser event stream.

It contains:

- Session metadata (target URL, started/stopped timestamps, browser choice, fingerprint config, completion policy, endpoint list)
- Ordered `HttpExchange` list
- Browser storage snapshots
- Cookie mutations
- Trace/source artifact references
- Recorder warnings

### HTTP Exchange

Each `HttpExchange` represents one request/response pair. Redirect hops are recorded as separate exchanges with explicit `RedirectEdge` relationships.

It contains:

- Stable ID
- Request method, URL, headers, cookies, body, and body type
- Response status, headers, cookies, body, and body type
- `RedirectEdge` parent/child IDs
- Timing
- Capture completeness flags
- Source references back to raw browser events

### Parsed Trees

Every request/response artifact is parsed into `ParsedTree`.

Each tree leaf includes:

- Path, such as `request.body.json.auth.csrf`
- Type, such as string, number, boolean, null, or bytes
- Raw value
- Normalized values
- Entropy score
- Shape classification, such as `jwt`, `uuid`, `timestamp`, `csrf_like`, or `nonce_like`
- Source exchange ID

This makes matching cheap and deterministic.

### Analysis Output

The analyzer emits:

- `DynamicCandidate`
- `DependencyCandidate`
- `NoiseCandidate`
- `LogicalOperationCandidate`
- `ChallengeCandidate`
- `UnresolvedBinding` (field with no deterministic source, requires user override)

Each candidate includes:

- Confidence score
- Evidence paths
- Reason code
- Source: deterministic, AI, user override, or runtime observed
- Status: accepted, rejected, or unresolved

### Executable Graph

The graph is the intermediate representation used by code generation.

It contains:

- Nodes: HTTP request, extract, bind, cookie update, storage update, redirect, user-override binding, logical operation
- Edges: data dependencies and execution ordering
- Runtime requirements
- Success conditions
- Unresolved or low-confidence regions

The graph is pure HTTP. It never includes browser-assisted nodes.

### Artifact Format

Each recording produces a project-local artifact directory:

```text
.autohttp/sessions/<session-id>/
  raw/
    browser-events.pb
  normalized/
    session.pb
    session.json
  analysis/
    trees.pb
    value-index.json
    candidates.json
  graph/
    graph.pb
    graph.json
  overrides.yaml
  warnings.json
```

The JSON files are for inspection and debugging. The protobuf files are the source of truth.
