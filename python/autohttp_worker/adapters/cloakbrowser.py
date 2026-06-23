"""CloakBrowser (Chromium) browser adapter stub.

Wraps `cloakbrowser.launch` behind the shared BrowserAdapter
Protocol. The live browser launch is not part of this skeleton —
launch() returns a placeholder handle and capture_events() yields no
events. The real cloakbrowser import is added when the live recorder
wiring lands.
"""

from __future__ import annotations

from collections.abc import Iterator
from typing import ClassVar

from autohttp.v1 import browser_pb2
from autohttp_worker.adapters.base import BrowserHandle, LaunchConfig


class CloakBrowserAdapter:
    name: ClassVar[str] = "cloak"

    def launch(self, config: LaunchConfig) -> BrowserHandle:
        return BrowserHandle(engine=None)

    def capture_events(
        self, handle: BrowserHandle
    ) -> Iterator[browser_pb2.BrowserEvent]:
        return iter(())

    def close(self, handle: BrowserHandle) -> None:
        return None
