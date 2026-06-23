"""Unit tests for EndpointMatcher.

Coverage:
- OPTIONS and HEAD are ignored for matching.
- Ordered cursor advancement.
- URL pattern matching (exact path, query string, sub-path).
- Method matching (when goal method is set).
- Out-of-order matches do not advance.
- Terminal endpoint detection and mark_terminal_settled.
- State is updated correctly through the lifecycle.
"""

from __future__ import annotations

import sys
from pathlib import Path

import pytest

ROOT = Path(__file__).resolve().parent.parent.parent.parent
sys.path.insert(0, str(ROOT / "python"))

from autohttp.v1 import browser_pb2
from autohttp_worker.endpoint import EndpointMatcher


def _request_started(request_id: str, method: str, url: str) -> browser_pb2.BrowserEvent:
    return browser_pb2.BrowserEvent(
        request_started=browser_pb2.RequestStarted(
            request_id=request_id,
            method=method,
            url=url,
        )
    )


def _response_headers(request_id: str, status: int) -> browser_pb2.BrowserEvent:
    return browser_pb2.BrowserEvent(
        response_headers=browser_pb2.ResponseHeaders(
            request_id=request_id,
            status=status,
        )
    )


def _goal(
    endpoint_id: str,
    url_pattern: str,
    *,
    method: str = "",
    terminal: bool = False,
) -> browser_pb2.EndpointGoal:
    return browser_pb2.EndpointGoal(
        id=endpoint_id,
        url_pattern=url_pattern,
        method=method,
        terminal=terminal,
    )


def _is_started(ev: browser_pb2.BrowserEvent, endpoint_id: str, request_id: str) -> bool:
    s = ev.endpoint_request_started
    return s.endpoint_id == endpoint_id and s.request_id == request_id


def _is_completed(ev: browser_pb2.BrowserEvent, endpoint_id: str, request_id: str, status: int) -> bool:
    c = ev.endpoint_response_completed
    return (
        c.endpoint_id == endpoint_id
        and c.request_id == request_id
        and c.status == status
    )


class TestOptionsAndHeadIgnored:
    def test_options_request_does_not_emit_phase_event(self) -> None:
        m = EndpointMatcher([_goal("e0", "/login")])
        out = m.match_event(_request_started("r1", "OPTIONS", "https://example.com/login"))
        assert out == []
        assert m.state.cursor == 0
        assert m.state.matched_count == 0
        assert m.state.inflight_count == 0

    def test_head_request_does_not_emit_phase_event(self) -> None:
        m = EndpointMatcher([_goal("e0", "/login")])
        out = m.match_event(_request_started("r1", "HEAD", "https://example.com/login"))
        assert out == []
        assert m.state.cursor == 0

    def test_lowercase_options_is_also_ignored(self) -> None:
        m = EndpointMatcher([_goal("e0", "/login")])
        out = m.match_event(_request_started("r1", "options", "https://example.com/login"))
        assert out == []


class TestOrderedAdvancement:
    def test_first_endpoint_match_emits_request_started(self) -> None:
        m = EndpointMatcher(
            [_goal("e0", "/login"), _goal("e1", "/done", terminal=True)]
        )
        out = m.match_event(_request_started("r1", "GET", "https://example.com/login"))
        assert len(out) == 1
        assert _is_started(out[0], "e0", "r1")
        assert m.state.cursor == 0
        assert m.state.matched_count == 0
        assert m.state.inflight_count == 1

    def test_response_completion_advances_cursor(self) -> None:
        m = EndpointMatcher(
            [_goal("e0", "/login"), _goal("e1", "/done", terminal=True)]
        )
        m.match_event(_request_started("r1", "GET", "https://example.com/login"))
        out = m.match_event(_response_headers("r1", 200))
        assert len(out) == 1
        assert _is_completed(out[0], "e0", "r1", 200)
        assert m.state.cursor == 1
        assert m.state.matched_count == 1
        assert m.state.inflight_count == 0

    def test_full_sequence_completes_terminal(self) -> None:
        m = EndpointMatcher(
            [_goal("e0", "/login"), _goal("e1", "/done", terminal=True)]
        )
        m.match_event(_request_started("r1", "GET", "https://example.com/login"))
        m.match_event(_response_headers("r1", 200))
        m.match_event(_request_started("r2", "GET", "https://example.com/done"))
        out = m.match_event(_response_headers("r2", 200))
        assert _is_completed(out[0], "e1", "r2", 200)
        assert m.state.cursor == 2
        assert m.state.matched_count == 2
        assert m.state.terminal_reached is True

    def test_later_endpoint_url_does_not_advance_when_earlier_unmatched(self) -> None:
        m = EndpointMatcher(
            [_goal("e0", "/login"), _goal("e1", "/done", terminal=True)]
        )
        out = m.match_event(_request_started("r1", "GET", "https://example.com/done"))
        assert out == []
        assert m.state.cursor == 0
        assert m.state.matched_count == 0

    def test_response_for_non_matched_request_is_silently_ignored(self) -> None:
        m = EndpointMatcher([_goal("e0", "/login")])
        out = m.match_event(_response_headers("unknown", 200))
        assert out == []
        assert m.state.cursor == 0


class TestTerminalSettle:
    def test_terminal_reached_false_before_response(self) -> None:
        m = EndpointMatcher([_goal("e0", "/login", terminal=True)])
        assert m.terminal_reached is False

    def test_terminal_reached_true_after_response(self) -> None:
        m = EndpointMatcher([_goal("e0", "/login", terminal=True)])
        m.match_event(_request_started("r1", "GET", "https://example.com/login"))
        m.match_event(_response_headers("r1", 200))
        assert m.terminal_reached is True

    def test_mark_terminal_settled_emits_event(self) -> None:
        m = EndpointMatcher([_goal("e0", "/login", terminal=True)])
        m.match_event(_request_started("r1", "GET", "https://example.com/login"))
        m.match_event(_response_headers("r1", 200))
        settled = m.mark_terminal_settled("network_idle")
        assert settled is not None
        assert settled.endpoint_settled.endpoint_id == "e0"
        assert settled.endpoint_settled.settle_reason == "network_idle"

    def test_mark_terminal_settled_returns_none_before_terminal(self) -> None:
        m = EndpointMatcher([_goal("e0", "/login", terminal=True)])
        assert m.mark_terminal_settled("response") is None

    def test_mark_terminal_settled_returns_none_with_no_goals(self) -> None:
        m = EndpointMatcher([])
        assert m.mark_terminal_settled("response") is None


class TestUrlMatching:
    @pytest.mark.parametrize(
        "url,expected",
        [
            ("https://example.com/login", True),
            ("https://example.com/login?next=/x", True),
            ("https://example.com/login/sub", True),
            ("https://example.com/logout", False),
            ("https://example.com/login-extra", False),
        ],
    )
    def test_url_pattern(self, url: str, expected: bool) -> None:
        m = EndpointMatcher([_goal("e0", "/login")])
        out = m.match_event(_request_started("r1", "GET", url))
        assert (len(out) == 1) is expected


class TestMethodMatching:
    def test_method_mismatch_does_not_emit(self) -> None:
        m = EndpointMatcher([_goal("e0", "/login", method="POST")])
        out = m.match_event(_request_started("r1", "GET", "https://example.com/login"))
        assert out == []

    def test_method_match_emits(self) -> None:
        m = EndpointMatcher([_goal("e0", "/login", method="POST")])
        out = m.match_event(_request_started("r1", "POST", "https://example.com/login"))
        assert len(out) == 1
        assert _is_started(out[0], "e0", "r1")

    def test_method_match_is_case_insensitive(self) -> None:
        m = EndpointMatcher([_goal("e0", "/login", method="post")])
        out = m.match_event(_request_started("r1", "POST", "https://example.com/login"))
        assert len(out) == 1


class TestEventFiltering:
    def test_storage_snapshot_is_passthrough(self) -> None:
        m = EndpointMatcher([_goal("e0", "/login")])
        from autohttp.v1 import session_pb2
        ev = browser_pb2.BrowserEvent(
            storage_snapshot=session_pb2.StorageSnapshot(url="https://example.com")
        )
        assert m.match_event(ev) == []

    def test_browser_launched_is_passthrough(self) -> None:
        m = EndpointMatcher([_goal("e0", "/login")])
        ev = browser_pb2.BrowserEvent(
            browser_launched=browser_pb2.BrowserLaunched()
        )
        assert m.match_event(ev) == []
