"""Common utilities for the project."""

import asyncio
from functools import wraps
from typing import Any, Callable, List, cast
from logging import Logger

from llama_stack_client import AsyncLlamaStackClient
from llama_stack import AsyncLlamaStackAsLibraryClient

from client import AsyncLlamaStackClientHolder
from models.config import Configuration, ModelContextProtocolServer


async def register_mcp_servers_async(
    logger: Logger, configuration: Configuration
) -> None:
    """Register Model Context Protocol (MCP) servers with the LlamaStack client (async)."""
    # Skip MCP registration if no MCP servers are configured
    if not configuration.mcp_servers:
        logger.debug("No MCP servers configured, skipping registration")
        return

    if configuration.llama_stack.use_as_library_client:
        # Library client - use async interface
        client = cast(
            AsyncLlamaStackAsLibraryClient, AsyncLlamaStackClientHolder().get_client()
        )
        await client.initialize()
        await _register_mcp_toolgroups_async(client, configuration.mcp_servers, logger)
    else:
        # Service client - also use async interface
        client = AsyncLlamaStackClientHolder().get_client()
        await _register_mcp_toolgroups_async(client, configuration.mcp_servers, logger)


async def _register_mcp_toolgroups_async(
    client: AsyncLlamaStackClient,
    mcp_servers: List[ModelContextProtocolServer],
    logger: Logger,
) -> None:
    """Async logic for registering MCP toolgroups."""
    # Get registered tools
    registered_toolgroups = await client.toolgroups.list()
    registered_toolgroups_ids = [
        tool_group.provider_resource_id for tool_group in registered_toolgroups
    ]
    logger.debug("Registered toolgroups: %s", registered_toolgroups_ids)

    # Register toolgroups for MCP servers if not already registered
    for mcp in mcp_servers:
        if mcp.name not in registered_toolgroups_ids:
            logger.debug("Registering MCP server: %s, %s", mcp.name, mcp.url)

            registration_params = {
                "toolgroup_id": mcp.name,
                "provider_id": mcp.provider_id,
                "mcp_endpoint": {"uri": mcp.url},
            }

            await client.toolgroups.register(**registration_params)
            logger.debug("MCP server %s registered successfully", mcp.name)


def run_once_async(func: Callable) -> Callable:
    """Decorate an async function to run only once."""
    task = None

    @wraps(func)
    async def wrapper(*args: Any, **kwargs: Any) -> Any:
        nonlocal task
        if task is None:
            loop = asyncio.get_running_loop()
            task = loop.create_task(func(*args, **kwargs))
        return await task

    return wrapper
