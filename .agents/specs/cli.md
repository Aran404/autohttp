# autohttp — CLI & User Workflow

Date: 2026-06-21

## CLI And User Workflow

### Primary User Workflow

The core workflow should be simple:

```bash
autohttp record https://example.com/login
# user completes workflow in CamoFox
autohttp stop
autohttp inspect
autohttp generate --target go --mode api
```

### Commands

**`autohttp record <url>`** — starts a new recording session.

Flags:

- `--session <name>`
- `--proxy <url>`
- `--profile <path>`
- `--recorder trace|network`
- `--vnc`
- `--headful`
- `--no-ai`
- `--recordings <n>`

**`autohttp stop`** — finalizes the active recording.

Responsibilities:

- Close and flush CamoFox tracing or network export
- Fetch storage state
- Write raw artifacts
- Normalize session

**`autohttp analyze`** — runs deterministic analysis.

Flags:

- `--no-ai`
- `--ai auto|off|always`
- `--ai-budget <n>`
- `--ai-threshold <score>`
- `--compare <session-a>,<session-b>`
- `--apply-overrides`

**`autohttp inspect`** — provides interactive or textual inspection of:

- Captured requests
- Parsed trees
- Dynamic fields
- Dependency edges
- Confidence scores
- Noise classifications
- Required user decisions
- AI escalation packets

**`autohttp generate`** — generates a standalone script.

Flags:

- `--target go|python`
- `--mode replay|api`
- `--output <path>`
- `--runtime pure-http|browser-assisted|auto`
- `--include-noise`
- `--fail-on-unresolved`

**`autohttp verify`** — runs a generated script and checks success conditions.

Flags:

- `--script <path>`
- `--success-url <pattern>`
- `--success-status <code>`
- `--success-json-path <path>`
- `--timeout <duration>`

### UX Principle

The tool should never hide uncertainty.

If the analyzer is unsure, it should show:

- What is ambiguous
- Why it matters
- What evidence exists
- Whether AI can help
- What user override would resolve it