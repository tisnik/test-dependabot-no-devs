"""Unit tests for routers.py."""

from collections.abc import Callable, Sequence
from typing import Any, Optional

from fastapi import FastAPI

from app.endpoints import (
    a2a,
    authorized,
    config,
    conversations_v1,
    conversations_v2,
    feedback,
    health,
    info,
    mcp_auth,
    mcp_servers,
    metrics,
    models,
    providers,
    query,
    rags,
    responses,
    rlsapi_v1,
    root,
    shields,
    stream_interrupt,
    streaming_query,
    tools,
)
from app.routers import include_routers


class MockFastAPI(FastAPI):
    """Mock class for FastAPI."""

    def __init__(self) -> None:  # pylint: disable=super-init-not-called
        """
        Initialize the MockFastAPI router registry.
        
        Create a mock FastAPI-like app and set `self.routers` to an empty list that will store tuples `(router, prefix)`, where `prefix` is a route prefix string or `None`.
        """
        self.routers: list[tuple[Any, Optional[str]]] = []

    def include_router(  # pylint: disable=too-many-arguments
        self,
        router: Any,
        *,
        prefix: str = "",
        tags: Optional[list] = None,
        dependencies: Optional[Sequence] = None,
        responses: Optional[dict] = None,  # pylint: disable=redefined-outer-name
        deprecated: Optional[bool] = None,
        include_in_schema: Optional[bool] = None,
        default_response_class: Optional[Any] = None,
        callbacks: Optional[list] = None,
        generate_unique_id_function: Optional[Callable] = None,
    ) -> None:
        """Register new router.

        Register a router and its mount prefix on the mock FastAPI
        app for test inspection.

        Parameters:
        ----------
            router (Any): Router object to register.
            prefix (str): Mount prefix to associate with the router.

        Notes:
        -----
            Accepts additional FastAPI-compatible parameters for
            API compatibility but ignores them; only the (router,
            prefix) pair is recorded.
        """
        self.routers.append((router, prefix))

    def get_routers(self) -> list[Any]:
        """Retrieve all routers defined in mocked REST API.

        Returns:
            routers (list[Any]): List of registered router objects in the order they were added.
        """
        return [r[0] for r in self.routers]

    def get_router_prefix(self, router: Any) -> Optional[str]:
        """Retrieve router prefix configured for mocked REST API.

        Get the prefix associated with a registered router in the mock FastAPI.

        Parameters:
        ----------
            router (Any): Router object to look up.

        Returns:
        -------
            Optional[str]: The prefix string for the router, or `None` if the
            router was registered without a prefix.

        Raises:
        ------
            IndexError: If the router is not registered in the mock app.
        """
        return list(filter(lambda r: r[0] == router, self.routers))[0][1]


def test_include_routers() -> None:
    """
    Verify that include_routers registers all expected endpoint routers on the FastAPI app.
    
    Asserts that exactly 22 routers are recorded on the mock app and that each expected endpoint router (e.g., root, info, models, tools, mcp_auth, mcp_servers, shields, providers, query, streaming_query, config, feedback, health, authorized, conversations_v1, conversations_v2, metrics, rlsapi_v1, a2a, stream_interrupt, responses) has been included. One conversations-related assertion is intentionally commented out.
    """
    app = MockFastAPI()
    include_routers(app)

    # are all routers added?
    assert len(app.routers) == 22
    assert root.router in app.get_routers()
    assert info.router in app.get_routers()
    assert models.router in app.get_routers()
    assert tools.router in app.get_routers()
    assert mcp_auth.router in app.get_routers()
    assert mcp_servers.router in app.get_routers()
    assert shields.router in app.get_routers()
    assert providers.router in app.get_routers()
    assert query.router in app.get_routers()
    assert streaming_query.router in app.get_routers()
    assert config.router in app.get_routers()
    assert feedback.router in app.get_routers()
    assert health.router in app.get_routers()
    assert authorized.router in app.get_routers()
    # assert conversations.router in app.get_routers()
    assert conversations_v2.router in app.get_routers()
    assert conversations_v1.router in app.get_routers()
    assert metrics.router in app.get_routers()
    assert rlsapi_v1.router in app.get_routers()
    assert a2a.router in app.get_routers()
    assert stream_interrupt.router in app.get_routers()
    assert responses.router in app.get_routers()


def test_check_prefixes() -> None:
    """
    Verify include_routers registers the expected routers with their configured URL prefixes.
    
    Asserts that 22 routers are registered on a MockFastAPI instance and that each router's mount prefix matches the expected value: empty string for root, health, authorized, metrics, and a2a; "/v1" for the majority of API routers; and "/v2" for conversations_v2.
    """
    app = MockFastAPI()
    include_routers(app)

    # are all routers added?
    assert len(app.routers) == 22
    assert app.get_router_prefix(root.router) == ""
    assert app.get_router_prefix(info.router) == "/v1"
    assert app.get_router_prefix(models.router) == "/v1"
    assert app.get_router_prefix(tools.router) == "/v1"
    assert app.get_router_prefix(mcp_auth.router) == "/v1"
    assert app.get_router_prefix(mcp_servers.router) == "/v1"
    assert app.get_router_prefix(shields.router) == "/v1"
    assert app.get_router_prefix(providers.router) == "/v1"
    assert app.get_router_prefix(rags.router) == "/v1"
    assert app.get_router_prefix(query.router) == "/v1"
    assert app.get_router_prefix(streaming_query.router) == "/v1"
    assert app.get_router_prefix(config.router) == "/v1"
    assert app.get_router_prefix(feedback.router) == "/v1"
    assert app.get_router_prefix(health.router) == ""
    assert app.get_router_prefix(authorized.router) == ""
    # assert app.get_router_prefix(conversations.router) == "/v1"
    assert app.get_router_prefix(conversations_v2.router) == "/v2"
    assert app.get_router_prefix(conversations_v1.router) == "/v1"
    assert app.get_router_prefix(metrics.router) == ""
    assert app.get_router_prefix(rlsapi_v1.router) == "/v1"
    assert app.get_router_prefix(a2a.router) == ""
    assert app.get_router_prefix(stream_interrupt.router) == "/v1"
    assert app.get_router_prefix(responses.router) == "/v1"
