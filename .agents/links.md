# autohttp — Relevant Links

## Core Dependencies

### Browser engines (driven from Python)

- [Camoufox](https://github.com/daijro/camoufox) — Firefox fork with C++ anti-detection. Default browser for `autohttp record` via `camoufox.sync_api.Camoufox`.
- [CloakBrowser](https://github.com/CloakHQ/CloakBrowser) — Chromium with C++ source-level fingerprint patches. Optional browser for `autohttp record` via `cloakbrowser.launch`.
- [CamoFox website](https://camoufox.com) — Documentation and stealth overview for Camoufox.

### AI escalation (optional)

- [gpt4free (g4f)](https://github.com/xtekky/gpt4free) — Free LLM provider aggregation library (Python). Default open-source provider for the AI worker; not an architectural dependency.
- [g4f docs](https://g4f.dev/docs) — GPT4Free documentation.
- [g4f providers](https://g4f.dev/docs/providers-and-models) — Available model providers.

### Interop

- [Protocol Buffers](https://protobuf.dev/) — Contract language for Go/Python interop.
- [gRPC](https://grpc.io/) — Bidirectional streaming between Go CLI and the Python browser worker.
- [Playwright](https://playwright.dev/) — Underlying automation API used by Camoufox and CloakBrowser adapters.

## References

- [Camoufox Python interface](https://github.com/daijro/camoufox/tree/main/pythonlib) — Python bindings for Camoufox.
- [CloakBrowser API](https://github.com/CloakHQ/CloakBrowser#api) — Launch options, persistent contexts, fingerprint flags.
- [Camoufox Stealth Overview](https://camoufox.com/stealth) — How fingerprint injection is done at the C++ level.
- [Playwright Trace Viewer](https://playwright.dev/docs/trace-viewer) — Used for inspecting recorded browser sessions.

## Conventions

- Specs live in `.agents/specs/`
- Go follows standard project layout (`cmd/`, `internal/`, domain-specific public packages)
- Python follows `python/<package>/` layout with `pyproject.toml`
- Generated scripts are pure HTTP and use stdlib only (Go and Python runtimes live in `runtime/go/` and `runtime/python/`)
