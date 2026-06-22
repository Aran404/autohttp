# AutoHTTP

Record browser workflows, reconstruct HTTP dependency graphs, and generate standalone replay scripts in Go or Python with fresh dynamic state.

## What it does

You record a browser workflow once (login, form submission, API call chain) via CamoFox. autohttp reconstructs the functional HTTP dependency graph using deterministic analysis — tree parsing, value indexing, and dependency matching — then generates a standalone script that replays the workflow with fresh values at runtime.

No AI dependency on the critical path. The analyzer uses tree comparison, entropy classification, and repeated-recording diffing to infer dependencies. AI is only called for remaining ambiguity, and its output is validated by Go before acceptance.

## Pipeline

```
Record → Normalize → Parse Trees → Build Value Index → 
Deterministic Analysis → Dependency Graph → Code Generation
```

Generated scripts are self-contained — no dependency on autohttp, g4f, or gRPC workers.

## Status

**Work in progress.** Not yet usable.

| Component | Status |
|-----------|--------|
| CLI scaffold | Done |
| Design specs | Done |
| Protobuf contracts | Defined |
| Tree parser | Not started |
| Value index | Not started |
| Analyzer | Not started |
| Code generator | Not started |
| CamoFox integration | Not started |