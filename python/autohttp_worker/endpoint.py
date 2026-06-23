"""Live endpoint matching for browser events.

The Python browser worker runs this matcher against the canonical
BrowserEvent stream. The matcher is stateful: it tracks an ordered
cursor over the user-defined EndpointGoal list, ignores OPTIONS and
HEAD for matching purposes, and emits the three phase events
(EndpointRequestStarted, EndpointResponseCompleted, EndpointSettled)
as the user-defined milestones are reached.
"""

from __future__ import annotations

from collections.abc import Sequence
from dataclasses import dataclass
from typing import Final
from urllib.parse import urlsplit

from autohttp.v1 import browser_pb2

_IGNORED_METHODS: Final[frozenset[str]] = frozenset({"OPTIONS", "HEAD"})


@dataclass(frozen=True)
class EndpointState:
    """Read-only view of the matcher's progress."""

    cursor: int
    matched_count: int
    inflight_count: int
    terminal_reached: bool


def _started(request_id: str, endpoint_id: str) -> browser_pb2.BrowserEvent:
    return browser_pb2.BrowserEvent(
        endpoint_request_started=browser_pb2.EndpointRequestStarted(
            request_id=request_id,
            endpoint_id=endpoint_id,
        )
    )


def _completed(request_id: str, endpoint_id: str, status: int) -> browser_pb2.BrowserEvent:
    return browser_pb2.BrowserEvent(
        endpoint_response_completed=browser_pb2.EndpointResponseCompleted(
            request_id=request_id,
            endpoint_id=endpoint_id,
            status=status,
        )
    )


def _settled(endpoint_id: str, reason: str) -> browser_pb2.BrowserEvent:
    return browser_pb2.BrowserEvent(
        endpoint_settled=browser_pb2.EndpointSettled(
            endpoint_id=endpoint_id,
            settle_reason=reason,
        )
    )


def _url_matches(url: str, pattern: str) -> bool:
    path = urlsplit(url).path
    if path == pattern:
        return True
    if path.startswith(pattern + "/") or path.startswith(pattern + "?"):
        return True
    return False


def _method_matches(req_method: str, goal_method: str) -> bool:
    if not goal_method:
        return True
    return req_method.upper() == goal_method.upper()


class EndpointMatcher:
    """Live endpoint matcher over a canonical BrowserEvent stream.

    The matcher advances an internal cursor through the ordered list
    of EndpointGoals. Only the request whose URL and (optional) method
    match the current expected goal emits a request-started phase
    event. The response to that request emits a response-completed
    phase event and advances the cursor. Earlier or later matches are
    observed but do not advance the cursor. OPTIONS and HEAD requests
    never participate in matching.
    """

    def __init__(self, goals: Sequence[browser_pb2.EndpointGoal]) -> None:
        self._goals: list[browser_pb2.EndpointGoal] = list(goals)
        self._cursor: int = 0
        self._inflight: dict[str, str] = {}
        self._matched_count: int = 0
        self._terminal_reached: bool = False

    @property
    def state(self) -> EndpointState:
        return EndpointState(
            cursor=self._cursor,
            matched_count=self._matched_count,
            inflight_count=len(self._inflight),
            terminal_reached=self._terminal_reached,
        )

    @property
    def terminal_reached(self) -> bool:
        return self._terminal_reached

    def goals(self) -> list[browser_pb2.EndpointGoal]:
        return list(self._goals)

    def mark_terminal_settled(self, reason: str) -> browser_pb2.BrowserEvent | None:
        """Record that the terminal endpoint has settled.

        The matcher itself does not wait for network-idle, URL, or
        timeout conditions; that is the responsibility of the server.
        The server calls this once the active completion policy fires
        and emits the resulting EndpointSettled event.
        """
        if not self._terminal_reached or not self._goals:
            return None
        return _settled(self._goals[-1].id, reason)

    def match_event(self, event: browser_pb2.BrowserEvent) -> list[browser_pb2.BrowserEvent]:
        if event.HasField("request_started"):
            return self._on_request(event.request_started)
        if event.HasField("response_headers"):
            return self._on_response(event.response_headers)
        return []

    def _on_request(self, req: browser_pb2.RequestStarted) -> list[browser_pb2.BrowserEvent]:
        if req.method.upper() in _IGNORED_METHODS:
            return []
        if self._cursor >= len(self._goals):
            return []
        goal = self._goals[self._cursor]
        if not _url_matches(req.url, goal.url_pattern):
            return []
        if not _method_matches(req.method, goal.method):
            return []
        self._inflight[req.request_id] = goal.id
        return [_started(req.request_id, goal.id)]

    def _on_response(self, resp: browser_pb2.ResponseHeaders) -> list[browser_pb2.BrowserEvent]:
        endpoint_id = self._inflight.pop(resp.request_id, None)
        if endpoint_id is None:
            return []
        self._cursor += 1
        self._matched_count += 1
        if self._cursor >= len(self._goals):
            self._terminal_reached = True
        return [_completed(resp.request_id, endpoint_id, resp.status)]
