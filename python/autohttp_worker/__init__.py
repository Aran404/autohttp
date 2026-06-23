"""autohttp_worker — Python browser recording worker.

Spawned by the Go CLI as a per-recording subprocess. Drives a single
browser engine (Camoufox or CloakBrowser) and streams canonical
BrowserEvents back to Go over a single bidirectional gRPC stream.
"""

from __future__ import annotations

import os
import sys

_GEN_DIR = os.path.join(os.path.dirname(__file__), "gen")
if _GEN_DIR not in sys.path:
    sys.path.insert(0, _GEN_DIR)
