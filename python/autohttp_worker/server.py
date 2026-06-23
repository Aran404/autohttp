"""gRPC server for the Python browser worker.

The Go CLI spawns this server as a per-recording subprocess, opens
one bidirectional gRPC stream (the Record RPC), and exchanges
BrowserCommand and BrowserEvent messages until the terminal endpoint
settles. The Python side drives the browser through a BrowserAdapter
and runs an EndpointMatcher over the produced event stream.
"""

from __future__ import annotations

import logging
from collections.abc import Iterable, Iterator
from concurrent import futures
from typing import Final

import grpc

from autohttp.v1 import browser_pb2, browser_pb2_grpc
from autohttp_worker.endpoint import EndpointMatcher

log = logging.getLogger("autohttp.worker")

_DEFAULT_PORT: Final[int] = 0


class BrowserWorkerServicer(browser_pb2_grpc.BrowserWorkerServicer):
    """Handles the bidirectional Record RPC for a single session.

    The servicer stops its gRPC server when the RPC ends, so the
    Python worker subprocess exits cleanly after the session.
    """

    def __init__(self, server: grpc.Server) -> None:
        self._server = server
        self._cancelled: bool = False
        self._matcher: EndpointMatcher | None = None

    def Record(
        self,
        request_iterator: Iterable[browser_pb2.BrowserCommand],
        context: grpc.ServicerContext,
    ) -> Iterator[browser_pb2.BrowserEvent]:
        log.debug("Record RPC opened")
        try:
            for command in request_iterator:
                if self._cancelled:
                    break
                kind = command.WhichOneof("command")
                if kind == "start_recording":
                    yield from self._handle_start(command.start_recording)
                elif kind == "cancel_recording":
                    self._cancelled = True
                    yield from self._handle_cancel(command.cancel_recording)
                    break
                elif kind == "update_settings":
                    log.debug("update_settings received: %s", command.update_settings.settings)
                else:
                    log.warning("unknown command: %r", kind)
        finally:
            log.debug("Record RPC closed (cancelled=%s), stopping server", self._cancelled)
            self._server.stop(grace=0)

    def _handle_start(
        self, start: browser_pb2.StartRecording
    ) -> Iterator[browser_pb2.BrowserEvent]:
        log.info(
            "start_recording: browser=%s url=%s endpoints=%d",
            browser_pb2.Browser.Name(start.browser),
            start.target_url,
            len(start.endpoints),
        )
        self._matcher = EndpointMatcher(start.endpoints)
        yield browser_pb2.BrowserEvent(
            browser_launched=browser_pb2.BrowserLaunched(browser=start.browser)
        )

    def _handle_cancel(
        self, cancel: browser_pb2.CancelRecording
    ) -> Iterator[browser_pb2.BrowserEvent]:
        log.info("cancel_recording: reason=%s", cancel.reason)
        yield browser_pb2.BrowserEvent(
            session_finalized=browser_pb2.SessionFinalized()
        )


def serve(port: int = _DEFAULT_PORT) -> int:
    """Start the gRPC server and block until the recording session ends.

    Returns the actual bound port (useful when port=0 for ephemeral
    binding). The Go CLI connects to the port the worker prints to
    stdout on startup. The server stops itself when the Record RPC
    closes, so the Python worker subprocess exits cleanly.
    """
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=2))
    browser_pb2_grpc.add_BrowserWorkerServicer_to_server(
        BrowserWorkerServicer(server), server
    )
    address = f"127.0.0.1:{port}"
    bound_port = server.add_insecure_port(address)
    server.start()
    print(bound_port, flush=True)
    log.info("autohttp-worker listening on %s", bound_port)
    server.wait_for_termination()
    log.info("autohttp-worker stopped")
    return bound_port


if __name__ == "__main__":
    logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(name)s: %(message)s")
    serve()
