---
name: python-conventions
description: Use when editing Python files, pyproject.toml, Python typing, pathlib, exceptions, async code, Pydantic models, Ruff, mypy, or pyright settings.
---

# Python Conventions

Apply these rules for Python changes in this project.

## Rules

Use explicit checks before exceptions-as-control-flow. Prefer `if key in mapping` over catching `KeyError` for normal branching.

Never use bare `except:` or `except Exception: pass`. Let unexpected failures surface.

Magic methods such as `__len__`, `__bool__`, and `__contains__` must be O(1).

Check `Path.exists()` before `Path.resolve()` or `Path.is_relative_to()` when the path may not exist.

Avoid import-time computation and side effects. Lazily compute expensive or environment-dependent values, using `functools.cache` when appropriate.

Assert before `typing.cast()` unless the code is a measured hot path.

Use `Literal` types for fixed string sets instead of plain `str`.

Declare variables close to their first use.

Use keyword-only arguments for functions with five or more parameters.

Do not add default parameter values. Make caller intent explicit.

Use strict static typing with `mypy --strict` or strict `pyright` when available.

Use Ruff for linting and formatting when available.

Use Pydantic models for data crossing process, network, config, or serialization boundaries.

## Common Mistakes

Do not return `None` silently on failure.

Do not hide filesystem errors behind broad exception handlers.

Do not add `async` for sequential code without actual concurrency.
