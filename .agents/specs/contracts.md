# autohttp — Core Data Contracts

Date: 2026-06-21

## Core Data Contracts

The data model should be contract-first using Protocol Buffers. Go and Python must not define separate hand-written structures.

### Protobuf Packages

Recommended contract layout:

- `proto/autohttp/v1/session.proto`
- `proto/autohttp/v1/tree.proto`
- `proto/autohttp/v1/analysis.proto`
- `proto/autohttp/v1/graph.proto`
- `proto/autohttp/v1/ai.proto`

Generated bindings:

- Go: `gen/autohttp/v1`
- Python: `python/autohttp_ai/gen/autohttp/v1`

### Recorded Session

`RecordedSession` is the normalized capture artifact.

It contains:

- Session metadata: target URL, started/stopped timestamps, recorder backend, and CamoFox config fingerprint
- Ordered `HttpExchange` list
- Browser storage snapshots
- Cookie mutations
- User actions if available
- Trace/source artifact references
- Recorder warnings

### HTTP Exchange

Each `HttpExchange` represents one request/response pair.

It contains:

- Stable ID
- Request method, URL, headers, cookies, body, and body type
- Response status, headers, cookies, body, and body type
- Redirect parent/child IDs
- Timing
- Initiator metadata where available
- Capture completeness flags
- Source references back to raw trace/network export

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

Each candidate includes:

- Confidence score
- Evidence paths
- Reason code
- Source: deterministic, AI, user override, or runtime observed
- Status: accepted, rejected, or unresolved

### Executable Graph

The graph is the intermediate representation used by code generation.

It contains:

- Nodes: HTTP, extract, bind, cookie update, storage update, JavaScript evaluate, captcha solve, and browser fallback
- Edges: data dependencies and execution ordering
- Runtime requirements
- Success conditions
- Unresolved or low-confidence regions

### Artifact Format

Each recording should produce a project-local artifact directory:

```text
.autohttp/sessions/<session-id>/
  raw/
    camofox-trace.zip
    network-events.jsonl
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