# autohttp Trust Boundaries & Error Handling

Date: 2026-06-23

## Trust Boundaries

### Process Boundary

The system is split across three process boundaries:

- Go CLI process
- Python browser worker subprocess
- Optional Python AI worker subprocess

Each boundary is a trust boundary. No Go process directly executes Python code or browser SDK code. No Python process directly imports from `autohttp`'s Go internals.

### gRPC Streaming Boundary

The Go CLI and the Python browser worker communicate exclusively over the local gRPC stream. The Python worker is a per-recording subprocess spawned by the Go CLI. There is no persistent daemon, no shared filesystem state during recording, and no port reuse.

Commands and events are defined in `proto/autohttp/v1/browser.proto`. Both sides validate the protobuf schema. Events with unknown fields or out-of-order phase events are rejected by Go.

### Browser Sandbox

The Python browser worker runs the browser through its adapter. Browser-specific quirks (Camoufox vs CloakBrowser) are isolated inside the adapter. The gRPC stream only carries the canonical `BrowserEvent` vocabulary.

The browser itself is launched with the anti-detection settings provided by its native engine (Camoufox patches Firefox at the C++ level, CloakBrowser patches Chromium at the C++ level). The Python adapter does not inject page-level JS for fingerprinting or detection evasion.

### AI Worker Isolation

`python/autohttp_ai` is a separate gRPC worker. It only receives minimized ambiguity packets. It cannot directly generate executable code. It cannot directly modify the session artifact. Its output is advisory and must be validated by Go before acceptance.

The AI provider boundary is provider-neutral. `g4f` is the default open-source provider, not an architectural dependency.

### Generated Script Boundary

Generated scripts are pure HTTP. They do not import from `autohttp`, do not connect to gRPC, do not drive a browser, and do not depend on AI providers. They run standalone. If a workflow requires a value the analyzer could not derive, the generated script fails loudly with a clear error message naming the missing user-override function.

## Error Handling

### Browser Crash

If the browser process crashes during recording:

- The Python adapter detects the crash and emits `BrowserCrashed` on the gRPC stream.
- Go marks the session as incomplete and writes a partial `RecordedSession` artifact.
- The CLI exits with code `3` and prints the failure reason.
- The user can rerun `autohttp record` and resume from the same session directory if desired, or start a new session.

### Network Failure

If the gRPC stream between Go and the Python worker fails:

- Go logs the failure and signals the Python worker to terminate.
- The Python worker closes the browser and exits.
- Go writes a partial `RecordedSession` and exits with code `3`.

### Analysis Failure

If `autohttp analyze` cannot produce a useful graph:

- The analyzer writes a partial result with rejected and unresolved candidates.
- The CLI exits with code `4` and prints the unresolved regions.
- The user can inspect the session, supply overrides, or rerun the analysis.

### Generation Failure

If the code generator cannot produce a target-language script:

- The generator writes an error report and exits with code `5`.
- Unsupported targets exit with code `5` before any work begins.

### Verification Failure

If the generated script does not reach the success URL:

- `autohttp verify` exits with code `6`.
- The verification report includes the script's final URL, status, and any extracted terminal values.
- The report is written to the session directory for inspection.

### Unresolved Required Values

If the analyzer produces a graph with unresolved required values:

- The generator emits a script that calls user-override stub functions.
- Running the script without implementing the stubs fails loudly with a clear error message naming the missing function.
- The user is expected to inspect the session, identify why the value is unresolved, and implement the stub.

## Confidence Policy

Every deterministic decision carries a confidence score. The user can override any decision through `overrides.yaml` in the session directory.

Overrides are versioned, inspectable, and reusable across recordings. Overrides are not retroactive: a new recording may need new overrides if the upstream sources change.

## User Overrides

User overrides are stored in `overrides.yaml` at the session root. The schema is YAML for human readability:

```yaml
bindings:
  - target: request[8].headers.x-signature
    source: user_function
    function: computeHeaderXSignature

noise:
  - exchange: 12
    reason: "internal telemetry endpoint, not part of functional flow"

endpoints:
  - pattern: "/custom-endpoint"
    method: POST
    status: 200
```

Overrides are applied at analysis time. They become part of the graph and are reflected in the generated script.

## Privacy

`autohttp` does not ship with telemetry. The Go CLI, the Python browser worker, and the Python AI worker do not report to any external endpoint.

Recorded sessions are stored under the user's project-local `.autohttp/sessions/` directory. The user is responsible for `.gitignore`-ing or otherwise protecting that directory if the workflow contains sensitive data.

Generated scripts contain only the request/response shapes and observed dynamic values necessary to replay the workflow. They do not contain cookies, sessions, or credentials from the recording.
