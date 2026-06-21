# autohttp — Relevant Links

## Core Dependencies

- [camofox-browser](https://github.com/jo-inc/camofox-browser) — Stealth headless browser for AI agents, wraps Camoufox patched Firefox
- [Camoufox](https://camoufox.com) — Firefox fork with C++ anti-detection
- [gpt4free (g4f)](https://github.com/xtekky/gpt4free) — Free LLM provider aggregation library (Python)
- [Protocol Buffers](https://protobuf.dev/) — Contract language for Go/Python interop
- [gRPC](https://grpc.io/) — Communication between Go core and Python AI worker

## References

- [Playwright Trace Viewer](https://playwright.dev/docs/trace-viewer) — Format used by CamoFox tracing
- [OpenAPI spec for camofox-browser](https://raw.githubusercontent.com/jo-inc/camofox-browser/master/openapi.json) — REST API reference
- [g4f docs](https://g4f.dev/docs) — GPT4Free documentation
- [g4f providers](https://g4f.dev/docs/providers-and-models) — Available model providers

## Conventions

- Specs live in `.agents/specs/` — see [AGENTS.md issue #71](https://github.com/agentsmd/agents.md/issues/71)
- Go follows standard project layout (`cmd/`, `internal/`, `pkg/`)
- Python follows `python/<package>/` layout with `pyproject.toml`