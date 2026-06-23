"""Camoufox (Firefox) browser adapter stub.

Wraps `camoufox.sync_api.Camoufox` behind the shared BrowserAdapter
Protocol. The live browser launch is not part of this skeleton —
launch() returns a placeholder handle and capture_events() yields no
events. The real camoufox import is added when the live recorder
wiring lands.
"""

from __future__ import annotations

from collections.abc import Iterator
from typing import ClassVar

from autohttp.v1 import browser_pb2
from autohttp_worker.adapters.base import BrowserHandle, LaunchConfig


class CamoufoxAdapter:
    name: ClassVar[str] = "camoufox"

    def launch(self, config: LaunchConfig) -> BrowserHandle:
        return BrowserHandle(engine=None)

    def capture_events(
        self, handle: BrowserHandle
    ) -> Iterator[browser_pb2.BrowserEvent]:
        return iter(())

    def close(self, handle: BrowserHandle) -> None:
        return None
