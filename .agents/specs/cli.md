# autohttp CLI & User Workflow

Date: 2026-06-23

## Commands

```text
autohttp record <url> [flags]
autohttp analyze [flags]
autohttp generate [flags]
autohttp verify [flags]
autohttp inspect [flags]
```

## Record

Start a browser recording session.

```bash
autohttp record <url> \
  --browser camoufox|cloak \
  --endpoints "/send-vcc" "/checksum" "/redeem" \
  --completion network-idle|response-only|url:/path|timeout:10s \
  --session <name> \
  --proxy <url> \
  --geoip
```

Flags:

- `--browser` (default `camoufox`) — which Python browser adapter to use.
- `--endpoints` — ordered list of URL path patterns. Each match advances the milestone. The final entry is terminal.
- `--completion` — terminal settle policy. Default `network-idle` with a 5 second quiet window.
- `--session` — optional session name. Defaults to a timestamp.
- `--proxy` — optional HTTP or SOCKS5 proxy URL applied to the browser launch.
- `--geoip` — auto-detect timezone and locale from the proxy exit IP.

Behavior:

- The Go CLI spawns the Python browser worker as a subprocess.
- The CLI opens the bidirectional gRPC stream and sends `StartRecording` with the configured options.
- The user drives the browser manually. The Python worker streams browser events to Go.
- Live endpoint matching happens in the Python worker. Go receives phase events and runs incremental analysis.
- On terminal settle, the CLI sends `CancelRecording`. The Python worker closes the browser and exits.
- The CLI writes the `RecordedSession` artifact to `.autohttp/sessions/<session-id>/`.

## Analyze

Run deterministic analysis on a persisted session.

```bash
autohttp analyze \
  --session <name|path> \
  --endpoints "/send-vcc" "/checksum" "/redeem" \
  --no-ai
```

Flags:

- `--session` — session name (looks under `.autohttp/sessions/`) or a direct path to a session directory.
- `--endpoints` — optional override of the endpoint list. The recording-time list is used by default.
- `--no-ai` — disable AI escalation entirely.

Behavior:

- Reads the session artifact directory.
- Runs tree parsing, value indexing, dependency discovery, dynamic field classification, noise filtering, and logical operation grouping.
- Emits accepted, rejected, and unresolved candidates with confidence scores and evidence paths.
- Writes updated analysis artifacts and the final graph to the session directory.

## Generate

Emit a standalone replay script from a session's graph.

```bash
autohttp generate \
  --session <name|path> \
  --target go|python \
  --output <path>
```

Flags:

- `--target` — `go` (default) or `python`.
- `--output` — destination file path. Defaults to a generated name in the session directory.

Behavior:

- Loads the analyzed graph from the session artifact.
- Emits a deterministic, template-based script.
- The script is pure HTTP. Unresolved dynamic values become explicit user-override stub functions.
- The script bundles only the required runtime helpers.

## Verify

Run a generated script against the live target and check it reaches the success condition.

```bash
autohttp verify \
  --script <path> \
  --target go|python \
  --success-url "https://target/dashboard"
```

Flags:

- `--script` — path to the generated script.
- `--target` — which runtime to use for execution.
- `--success-url` — final URL the script must reach for verification to pass.

Behavior:

- Executes the generated script in a fresh process.
- Watches the script's outbound network for the terminal redirect chain and final response.
- Compares the final URL and any extracted terminal values against the expected success condition.
- Reports pass/fail and writes a verification report to the session directory.

## Inspect

Inspect a session's captures, dependencies, and unresolved regions.

```bash
autohttp inspect \
  --session <name|path> \
  --format json|tree|graph|value-index
```

Flags:

- `--session` — session name or path.
- `--format` — output format. `json` for raw, `tree` for parsed tree, `graph` for executable graph, `value-index` for the inverted value index.

Behavior:

- Reads the session artifact directory.
- Prints a structured view of the requested component.
- Useful for diagnosing unresolved regions and tuning endpoint definitions.

## Global Flags

- `--no-color` — disable ANSI color output.
- `--log-level debug|info|warn|error` — default `info`.
- `--session-root <path>` — root directory for session artifacts. Default `.autohttp/sessions/`.

## Exit Codes

- `0` — success
- `1` — generic error
- `2` — invalid arguments
- `3` — recording error (browser crash, network failure)
- `4` — analysis error (unresolvable graph, missing session)
- `5` — generation error (target unsupported, template failure)
- `6` — verification failure (script did not reach success condition)
