"""Unit tests for the pure ASGI middlewares in main.py."""

import json
from typing import cast

import pytest
from fastapi import HTTPException, status
from pytest_mock import MockerFixture
from starlette.types import Message, Receive, Scope, Send

from app.main import GlobalExceptionMiddleware, RestApiMetricsMiddleware
from models.responses import InternalServerErrorResponse


def _make_scope(path: str = "/test") -> dict:
    """Build a minimal HTTP ASGI scope."""
    return {
        "type": "http",
        "method": "GET",
        "path": path,
        "query_string": b"",
        "headers": [],
    }


async def _noop_receive() -> dict:
    """
    ASGI receive callable that immediately provides an empty HTTP request message.
    
    Returns:
        dict: ASGI message with `"type": "http.request"` and an empty `body` (`b""`).
    """
    return {"type": "http.request", "body": b""}


class _ResponseCollector:
    """Accumulate ASGI messages so tests can inspect them."""

    def __init__(self) -> None:
        """
        Initialize the response collector.
        
        Creates an instance with an empty `messages` list used to accumulate ASGI `Message` objects received from an ASGI application.
        """
        self.messages: list[Message] = []

    async def __call__(self, message: Message) -> None:
        """
        Append an incoming ASGI message to the collector.
        
        Parameters:
            message (Message): The ASGI message to record (e.g., `{"type": "http.response.start", ...}` or `{"type": "http.response.body", "body": b"...", ...}`).
        """
        self.messages.append(message)

    @property
    def status_code(self) -> int:
        """
        Get the HTTP status code from the collected ASGI response messages.
        
        Returns:
            status_code (int): The status value from the first message with `"type" == "http.response.start"`.
        
        Raises:
            AssertionError: If no `http.response.start` message is present in `self.messages`.
        """
        for msg in self.messages:
            if msg["type"] == "http.response.start":
                return msg["status"]
        raise AssertionError("No http.response.start message")

    @property
    def body_json(self) -> dict:
        """
        Decode and return the collected HTTP response body as a JSON object.
        
        Returns:
            dict: The parsed JSON body built from the concatenated `body` fields of
            collected `http.response.body` ASGI messages.
        """
        body = b""
        for msg in self.messages:
            if msg["type"] == "http.response.body":
                body += msg.get("body", b"")
        return json.loads(body)


# ---------------------------------------------------------------------------
# GlobalExceptionMiddleware
# ---------------------------------------------------------------------------


@pytest.mark.asyncio
async def test_global_exception_middleware_catches_unexpected_exception() -> None:
    """Test that GlobalExceptionMiddleware catches unexpected exceptions."""

    async def failing_app(scope: Scope, receive: Receive, send: Send) -> None:
        """
        A minimal ASGI application that always raises a ValueError for testing.
        
        Raises:
            ValueError: "This is an unexpected error for testing".
        """
        raise ValueError("This is an unexpected error for testing")

    middleware = GlobalExceptionMiddleware(failing_app)
    collector = _ResponseCollector()

    await middleware(_make_scope(), _noop_receive, collector)

    assert collector.status_code == status.HTTP_500_INTERNAL_SERVER_ERROR

    detail = collector.body_json["detail"]
    assert isinstance(detail, dict)

    expected_response = InternalServerErrorResponse.generic()
    expected_detail = expected_response.model_dump()["detail"]
    detail_dict = cast(dict[str, str], detail)
    assert detail_dict["response"] == expected_detail["response"]
    assert detail_dict["cause"] == expected_detail["cause"]


@pytest.mark.asyncio
async def test_global_exception_middleware_passes_through_http_exception() -> None:
    """Test that GlobalExceptionMiddleware passes through HTTPException."""

    async def http_error_app(scope: Scope, receive: Receive, send: Send) -> None:
        """
        ASGI application that immediately raises an HTTP 400 error with a fixed test detail.
        
        Raises:
            HTTPException: with status_code 400 and detail {"response": "Test error", "cause": "This is a test"}.
        """
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail={"response": "Test error", "cause": "This is a test"},
        )

    middleware = GlobalExceptionMiddleware(http_error_app)
    collector = _ResponseCollector()

    with pytest.raises(HTTPException) as exc_info:
        await middleware(_make_scope(), _noop_receive, collector)

    assert exc_info.value.status_code == status.HTTP_400_BAD_REQUEST
    detail = cast(dict[str, str], exc_info.value.detail)
    assert detail["response"] == "Test error"
    assert detail["cause"] == "This is a test"


@pytest.mark.asyncio
async def test_global_exception_middleware_reraises_when_response_started() -> None:
    """Test that exceptions after response headers are sent are re-raised."""

    async def partial_response_app(
        _scope: Scope, _receive: Receive, send: Send
    ) -> None:
        """
        ASGI application that sends an HTTP 200 response start and then raises a runtime error.
        
        This callable sends a single `http.response.start` message with status 200 and then raises RuntimeError to simulate an error occurring after response headers have been sent.
        
        Raises:
            RuntimeError: Always raised after sending the response start message.
        """
        await send({"type": "http.response.start", "status": 200, "headers": []})
        raise RuntimeError("error after headers sent")

    middleware = GlobalExceptionMiddleware(partial_response_app)
    collector = _ResponseCollector()

    with pytest.raises(RuntimeError, match="error after headers sent"):
        await middleware(_make_scope(), _noop_receive, collector)


@pytest.mark.asyncio
async def test_global_exception_middleware_skips_non_http() -> None:
    """Test that non-HTTP scopes pass through untouched."""
    called = False

    async def inner_app(_scope: Scope, _receive: Receive, _send: Send) -> None:
        """
        Marks that the inner ASGI application was invoked by setting the enclosing `called` flag to True.
        
        Parameters:
            _scope (Scope): ASGI connection scope (ignored).
            _receive (Receive): ASGI receive callable (ignored).
            _send (Send): ASGI send callable (ignored).
        """
        nonlocal called
        called = True

    middleware = GlobalExceptionMiddleware(inner_app)
    await middleware({"type": "websocket"}, _noop_receive, _ResponseCollector())
    assert called


# ---------------------------------------------------------------------------
# RestApiMetricsMiddleware
# ---------------------------------------------------------------------------


@pytest.mark.asyncio
async def test_rest_api_metrics_skips_non_http() -> None:
    """Test that non-HTTP scopes pass through untouched."""
    called = False

    async def inner_app(_scope: Scope, _receive: Receive, _send: Send) -> None:
        """
        Marks that the inner ASGI application was invoked by setting the enclosing `called` flag to True.
        
        Parameters:
            _scope (Scope): ASGI connection scope (ignored).
            _receive (Receive): ASGI receive callable (ignored).
            _send (Send): ASGI send callable (ignored).
        """
        nonlocal called
        called = True

    middleware = RestApiMetricsMiddleware(inner_app)
    await middleware({"type": "websocket"}, _noop_receive, _ResponseCollector())
    assert called


@pytest.mark.asyncio
async def test_rest_api_metrics_increments_counter_on_exception(
    mocker: MockerFixture,
) -> None:
    """Counter must be incremented even when the inner app raises."""
    mocker.patch("app.main.app_routes_paths", ["/v1/infer"])
    mock_metrics = mocker.patch("app.main.metrics")

    async def failing_app(_scope: Scope, _receive: Receive, _send: Send) -> None:
        """
        ASGI application callable that immediately raises a RuntimeError.
        
        Raises:
            RuntimeError: Always raised with the message "boom".
        """
        raise RuntimeError("boom")

    middleware = RestApiMetricsMiddleware(failing_app)

    with pytest.raises(RuntimeError, match="boom"):
        await middleware(_make_scope("/v1/infer"), _noop_receive, _ResponseCollector())

    mock_metrics.response_duration_seconds.labels.assert_called_once_with("/v1/infer")
    mock_metrics.rest_api_calls_total.labels.assert_called_once_with("/v1/infer", 500)
    mock_metrics.rest_api_calls_total.labels.return_value.inc.assert_called_once()
