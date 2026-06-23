# AutoHTTP

Record browser workflows, reconstruct HTTP dependency graphs, and generate standalone replay scripts in Go or Python with fresh dynamic state.

## What it does

You record a browser workflow once (login, form submission, API call chain) via Camoufox (Firefox) or CloakBrowser (Chromium). autohttp reconstructs the functional HTTP dependency graph using deterministic analysis — tree parsing, value indexing, and dependency matching — then generates a standalone script that replays the workflow with fresh values at runtime.

Browser control is delegated to a per-recording Python subprocess that drives the chosen engine and streams canonical `BrowserEvent`s back to the Go orchestrator over a single bidirectional gRPC stream.

No AI dependency on the critical path. The analyzer uses tree comparison, entropy classification, and repeated-recording diffing to infer dependencies. AI is only called for remaining ambiguity, and its output is validated by Go before acceptance.

## Pipeline

```
Browser (Python worker via gRPC) → Record → Normalize → 
Parse Trees → Build Value Index → Deterministic Analysis → 
Dependency Graph → Code Generation
```

Generated scripts are self-contained — no dependency on autohttp, g4f, or gRPC workers. They are pure HTTP and never drive a browser.

## Status

**Work in progress.** Skeleton and contracts are in place; live browser capture is not.

| Component | Status |
|-----------|--------|
| CLI (`record`, `analyze`, `generate`, `verify`, `inspect`) | Scaffolded, `record` runs end-to-end with stub browser |
| Design specs | Done |
| Protobuf contracts (`session`, `tree`, `analysis`, `graph`, `ai`, `browser`) | Done |
| Tree parser | Implemented, tested |
| Value index | Implemented, tested |
| Deterministic analyzer | Implemented, tested |
| Go code generator | Implemented, tested |
| Python browser worker | Scaffolded (gRPC server, endpoint matcher, adapter stubs) |
| Browser adapters (Camoufox, CloakBrowser) | Stubs only; live launch pending |
| Live recording flow | Not started (depends on live browser launch) |
| Python AI worker | Stub only |
