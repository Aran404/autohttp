# autohttp Deterministic Analysis & AI Escalation

Date: 2026-06-23

## Deterministic Analysis Strategy

This is the core of `autohttp`. The analyzer must do as much as possible without AI.

### Analysis Inputs

The analyzer consumes:

- Canonical `RecordedSession` produced by Go's normalizer from browser events
- Parsed trees
- Value index
- Optional repeated recordings
- Optional user overrides
- Optional recorder metadata such as initiator and timing

### Live Endpoint Matching

The Python browser worker performs live endpoint matching during recording. The analyzer trusts the resulting endpoint phase events and uses them to drive incremental analysis.

Phase events:

- `EndpointRequestStarted` — emitted on request start.
- `EndpointResponseCompleted` — emitted on response completion. The ordered milestone advances on this event.
- `EndpointSettled` — emitted for the terminal endpoint after bounded browser/network settle.

Go runs incremental analysis after each `EndpointResponseCompleted`. The incremental pass operates on all captured evidence up to and including the matched exchange.

After the terminal endpoint settles, the analyzer runs the final full-session pass, which can supersede incremental results if later evidence changed an earlier assumption.

### Value Matching

The first pass discovers direct value propagation.

Examples:

- Exact match: upstream response value appears in downstream request.
- Decoded match: URL/base64/HTML-decoded value appears downstream.
- Structured match: JWT payload claim appears in later request.
- Cookie match: `Set-Cookie` value appears in later `Cookie` header.
- Storage match: localStorage/sessionStorage value appears in request.
- HTML match: hidden form input appears in later form body.
- Redirect match: redirect query parameter appears in later exchange.

Direct matches should usually have high confidence.

### Shape And Entropy Classification

The second pass classifies dynamic-looking values even without a known source.

Signals:

- High entropy
- Long opaque value
- UUID shape
- JWT shape
- Timestamp or expiry shape
- Hash-like string
- CSRF/nonce/token/session/fingerprint field names
- Anti-bot parameter names
- Values that appear only once
- Values that change across repeated recordings

These signals do not automatically create dependencies. They mark fields as dynamic candidates.

### Sequence Reasoning

The analyzer respects order.

A downstream field can depend only on:

- Earlier responses
- Earlier cookies/storage mutations
- Earlier redirects
- Earlier browser observations
- Stable user-provided inputs

This prevents impossible AI-style guesses.

### Redirect Handling

Each redirect hop is recorded as a separate `HttpExchange` with an explicit `RedirectEdge` relationship. The analyzer does not collapse redirects into a single node because intermediate responses often carry `Set-Cookie`, auth codes, CSRF refreshes, or other required evidence.

The generated replay also preserves each hop and disables automatic `Location` follow.

### Noise Filtering

Noise detection is rule-based first.

Likely noise categories:

- Static assets: CSS, JavaScript bundles, fonts, images
- Analytics: GA, Segment, Amplitude, Mixpanel
- Ads/tracking
- Telemetry/error reporting
- Preload/prefetch
- Favicons/manifests
- Long-polling or repeated heartbeat requests unless they feed dependencies

Noise should be soft-deleted, not removed. Users can inspect and restore it.

### Repeated Recording Comparison

If the user provides multiple recordings of the same workflow, this becomes the strongest dynamic detector.

The analyzer compares tree paths across runs:

- Same path, same value: likely static
- Same path, different value: likely dynamic
- Same semantic role, different path: possible site variation
- Same request shape with changed token fields: dependency candidate
- Request appears in one run only: optional, noise, or challenge branch

This is cheaper and more reliable than LLM analysis.

### Unresolved Bindings

If a field has no deterministically discoverable source, the analyzer marks it as unresolved and the generator emits a user-override stub function with a highly explicit name (e.g., `computeHeaderXSignature`).

The unresolved binding includes:

- The exact target field path
- The observed value at recording time
- The reason it is unresolved (high entropy with no source, no upstream, etc.)

The user is responsible for implementing the stub. Generated scripts fail loudly with a clear error if a stub is left unimplemented.

### Confidence Scoring

Every candidate gets a confidence score and reason code.

| Scenario | Confidence |
|----------|-----------|
| Set-Cookie reused in later Cookie header | 1.00 |
| Exact response JSON value reused downstream | 0.98 |
| Hidden input reused in form body | 0.95 |
| Redirect code reused in token request | 0.95 |
| High-entropy field with token-like name | 0.70 |
| Name-only CSRF guess | 0.45 |
| AI suggestion without deterministic match | max 0.50 |
| Field with no source | unresolved (no confidence, requires user override) |

### AI Escalation Criteria

AI is called only when:

- A required graph region remains unresolved
- Multiple candidate dependencies have similar confidence
- Noise filtering impacts functional success
- Challenge/anti-bot classification is ambiguous
- Logical operation grouping is useful but unclear
- The user explicitly requests explanation or naming

The analyzer must support `--no-ai` and still produce the best deterministic result.

### Output

The analyzer emits:

- Accepted candidates
- Rejected candidates
- Unresolved candidates (with user-override binding metadata)
- Confidence scores
- Evidence paths
- Required user decisions
- Optional AI escalation packets

The graph engine consumes only accepted candidates and user-approved unresolved candidates.

## AI-Minimal Escalation Strategy

The AI layer exists only to resolve ambiguity that deterministic analysis cannot cheaply solve.

### Default Mode

By default, `autohttp analyze` runs mostly deterministic.

Recommended defaults:

- AI disabled unless the session has unresolved required graph regions.
- AI receives only minimized ambiguity packets.
- AI has a strict call budget.
- AI output is advisory and must be validated by Go.
- AI cannot directly generate executable code.

### AI Provider Boundary

`python/autohttp_ai` exposes a provider-neutral gRPC interface.

`g4f` is the default open-source provider, but the contract should not mention `g4f` directly.

This allows future providers:

- `g4f`
- OpenAI-compatible endpoints
- Local models
- Anthropic/Gemini wrappers
- Commercial hosted analyzer

### Ambiguity Packet

Instead of sending the full session, Go sends a small packet:

- Problem type: dependency, noise, challenge, operation naming, or explanation
- Candidate tree paths
- Candidate values, redacted when possible
- Deterministic confidence scores
- Evidence snippets
- Exact question
- Required output schema

Example question:

> Downstream field `request[7].headers.x-csrf-token` has three possible upstream sources. Pick the most likely source path or return unresolved.

### Output Contract

AI returns structured annotations only:

- Chosen candidate ID
- Confidence
- Reason
- Required evidence path
- Whether user approval is recommended

Go rejects responses that:

- Refer to paths not in the packet
- Invent values
- Exceed max allowed confidence
- Contradict deterministic evidence
- Do not match the protobuf schema

### Caching

AI calls should be cached by a deterministic hash of:

- Packet type
- Candidate paths
- Redacted values
- Model/provider config
- Prompt version

This prevents paying repeatedly for the same ambiguity.

### Budgets

CLI controls:

- `--no-ai`
- `--ai auto|off|always`
- `--ai-budget <n>`
- `--ai-threshold <confidence>`
- `--ai-provider <name>`
- `--ai-cache-dir <path>`

### Design Rule

AI should never be required for the happy path. A deterministic-only run may be less polished, but it must still produce an inspectable graph or a clear list of unresolved decisions.
