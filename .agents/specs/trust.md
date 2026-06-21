# autohttp — Trust Boundaries & Error Handling

Date: 2026-06-21

## Trust Boundaries And Error Handling

### Trust Boundaries

`autohttp` should treat every external component as untrusted or partially reliable:

- CamoFox is trusted for browser execution, but recorder output must be validated and normalized.
- Captured traffic is evidence, not truth. Malformed bodies, missing responses, redactions, and compressed payloads must be handled explicitly.
- Python/g4f AI is advisory only. Returned annotations must be grounded in observed data.
- Captcha providers are external services. Failures must be surfaced clearly and never silently bypassed.
- Generated scripts must be standalone and should not leak captured secrets beyond what the user explicitly generated.

### Error Handling Rules

The Go core should fail with actionable errors:

- If CamoFox fails to start, report command, port, stderr, and health-check result.
- If trace/network capture is incomplete, report which exchanges are missing request body, response body, headers, or timing.
- If dependency inference is ambiguous, preserve the ambiguity in `inspect` instead of guessing.
- If AI is disabled or budget exhausted, continue deterministic analysis and mark unresolved graph regions.
- If generation cannot produce a safe standalone script, fail before outputting misleading code.
- If verification fails, show which graph node failed and what extracted value or request differed.

### Confidence Policy

Every inferred decision should carry:

- `confidence`
- `reason`
- `evidence_paths`
- `source`, such as `deterministic`, `ai`, `user_override`, `captcha_provider`, or `runtime_observed`

Low-confidence decisions should not disappear. They should be visible in `autohttp inspect` and either require user approval or trigger optional AI escalation.

### User Overrides

Users should be able to correct the analyzer without editing generated code:

- Mark a request as functional or noise.
- Mark a field as static or dynamic.
- Manually bind one tree path to another.
- Add a success condition.
- Force pure HTTP or browser-assisted runtime for a node.

Overrides are stored as a session-sidecar config and applied before graph generation.