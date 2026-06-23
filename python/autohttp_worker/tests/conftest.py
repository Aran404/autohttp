"""Test configuration: makes autohttp_worker and its generated proto
modules importable when pytest runs from the package directory."""

from __future__ import annotations

import os
import sys

_PKG_ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
_GEN_DIR = os.path.join(_PKG_ROOT, "gen")
_REPO_PYTHON = os.path.dirname(_PKG_ROOT)

for entry in (_GEN_DIR, _REPO_PYTHON):
    if entry not in sys.path:
        sys.path.insert(0, entry)
