# autohttp Generated Script Runtime

Date: 2026-06-23

## Runtime Principle

Generated scripts are pure HTTP only. They never drive a browser, never call into `autohttp`, never connect to a gRPC service, and never depend on AI providers.

The runtime is a small, self-contained set of helpers bundled with the generated script. For Go, this lives in `runtime/go/`. For Python, this lives in `runtime/python/` and is included as a vendored package with the generated script.

## Runtime Capabilities

- HTTP client (no auto-redirect by default)
- Cookie jar
- Header/body templating
- Extractors for JSON, HTML, regex, cookies, and storage-like values
- User-override function hooks for unresolved values
- Optional TLS/client fingerprinting support (deferred; not part of the initial runtime)

The runtime must stay small. If a generated script only needs HTTP and JSON extraction, it should not include HTML or regex helpers.

## Go Runtime

`runtime/go/` provides:

- A `gort` package with a minimal `Client` type backed by `net/http`
- JSON path extraction
- HTML hidden input extraction
- Cookie jar handling
- `ReplayContext` passed to user-override functions
- Helpers for templating request bodies from extracted values

Initial dependencies: Go standard library only. No third-party HTTP client, no third-party TLS library.

## Python Runtime

`runtime/python/` provides:

- A `autohttp_runtime` package with a minimal `Client` class backed by `http.client` and `http.cookiejar`
- JSON path extraction
- HTML hidden input extraction
- Cookie jar handling
- `ReplayContext` passed to user-override functions
- Helpers for templating request bodies from extracted values

Initial dependencies: Python standard library only.

## Redirect Handling

The replay HTTP client must disable automatic `Location` follow.

Each redirect hop is preserved as a separate request node in the generated graph. The replay code issues a request, receives a `302/303/307/308`, applies the `Set-Cookie` headers, and then explicitly issues the next request to the `Location` URL. This avoids collapsing `302 -> 200/204` into a single opaque result and losing intermediate response evidence.

## User-Override Bindings

If the analyzer cannot derive a required dynamic value deterministically, the value becomes a user-overridable binding in the generated script. The generated stub uses a highly explicit function name that includes the target field name:

```go
// Computes request[8].headers.x-signature. Observed value at recording was "d4e5f6".
func computeHeaderXSignature(ctx ReplayContext) (string, error) {
    return "", errors.New("implement computeHeaderXSignature")
}
```

```python
# Computes request[8].headers.x-signature. Observed value at recording was "d4e5f6".
def compute_header_x_signature(ctx: ReplayContext) -> str:
    raise NotImplementedError("implement compute_header_x_signature")
```

The generated script calls these stubs at the appropriate place. The user is responsible for implementing them. If a script is run with unresolved stubs, it fails loudly with a clear error message naming the missing function.

## No Browser-Assisted Mode

The runtime never includes browser launch, browser automation, or JS evaluation helpers. If a workflow requires a value that cannot be observed in the request/response stream or computed from prior evidence, the user must provide the override. There is no "fallback to browser" option in the runtime.
