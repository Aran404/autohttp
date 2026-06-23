"""Browser engine adapters for autohttp_worker."""

from __future__ import annotations

from autohttp_worker.adapters.base import BrowserAdapter, BrowserHandle, LaunchConfig
from autohttp_worker.adapters.camoufox import CamoufoxAdapter
from autohttp_worker.adapters.cloakbrowser import CloakBrowserAdapter

__all__ = [
    "BrowserAdapter",
    "BrowserHandle",
    "CamoufoxAdapter",
    "CloakBrowserAdapter",
    "LaunchConfig",
]
