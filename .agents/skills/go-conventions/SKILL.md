---
name: go-conventions
description: Use when editing Go files, go.mod, Go packages, goroutines, channels, error handling, interfaces, internal packages, or Go service shutdown behavior.
---

# Go Conventions

Apply these rules for Go changes in this project.

## Rules

Handle errors explicitly. Never swallow errors, and never panic for expected failures.

Wrap errors with context using `fmt.Errorf("...: %w", err)` when returning them across a boundary.

Accept interfaces and return structs. Define interfaces in the consuming package, not the implementing package.

Keep interfaces small and compose them instead of creating hierarchies.

Do not create `pkg/`, `utils/`, `helpers`, `common`, or other dumping-ground packages. Choose package boundaries by responsibility.

Use `internal/` subfolders for domain grouping when the package is private, such as `internal/browser` or `internal/scraping`.

Keep filenames single words unless an existing package convention requires otherwise.

Prefer the simplest concurrency primitive. Do not add goroutines, channels, or worker pools unless the task requires concurrency.

Use `errgroup` for bounded fan-out when concurrent work must return errors.

Never spawn unbounded goroutines.

For long-running services, handle shutdown deliberately: stop accepting new work, drain in-flight requests, then close dependencies.

After Go changes, run the available equivalent of `gofmt` or `goimports`, `go vet`, `golangci-lint`, and `go test -race`.

## Common Mistakes

Do not introduce producer-side interfaces for hypothetical mocking.

Do not choose channels or goroutines when straight-line code or a mutex is enough.

Do not close dependencies before the server or workers have drained.
