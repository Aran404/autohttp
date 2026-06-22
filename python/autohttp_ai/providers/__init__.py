"""AI provider interface — provider-neutral gRPC adapter layer."""

from __future__ import annotations

from collections.abc import Mapping
from dataclasses import dataclass


@dataclass
class ProviderResponse:
    content: str
    model: str
    provider: str


class AIProvider:
    """Base interface for AI providers.

    Subclasses implement send() for their specific provider API.
    """

    async def send(self, prompt: str, context: Mapping[str, str] | None = None) -> ProviderResponse:
        raise NotImplementedError
