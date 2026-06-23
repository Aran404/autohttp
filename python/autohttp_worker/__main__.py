"""Entry point: `python -m autohttp_worker`."""

from __future__ import annotations

from autohttp_worker.server import serve

if __name__ == "__main__":
    serve()
