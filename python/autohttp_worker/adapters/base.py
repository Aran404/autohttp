"""BrowserAdapter Protocol shared by all engine adapters."""

from __future__ import annotations

from collections.abc import Iterator
from dataclasses import dataclass, field
from typing import Protocol

from autohttp.v1 import browser_pb2


@dataclass(frozen=True)
class LaunchConfig:
    target_url: str
    proxy_url: str = ""
    geoip: bool = False
    settings: dict[str, str] = field(default_factory=dict)


@dataclass
class BrowserHandle:
    """Opaque handle returned by an adapter's launch().

    Adapters return whatever browser-specific object they need
    (Camoufox context, CloakBrowser page, etc.) wrapped here.
    """

    engine: object


class BrowserAdapter(Protocol):
    """Shared interface for browser engine adapters.

    Each adapter translates engine-specific events into canonical
    BrowserEvent protobufs. Go only consumes the canonical vocabulary.
    """

    name: str

    def launch(self, config: LaunchConfig) -> BrowserHandle:
        ...

    def capture_events(self, handle: BrowserHandle) -> Iterator[browser_pb2.BrowserEvent]:
        ...

    def close(self, handle: BrowserHandle) -> None:
        ...
