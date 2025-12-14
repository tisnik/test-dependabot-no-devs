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
    """
    Register configured Model Context Protocol (MCP) servers with the project's LlamaStack client.
    
    If `configuration.mcp_servers` is empty the function returns immediately. Chooses a library client when
    `configuration.llama_stack.use_as_library_client` is true (initializing it before registration) or a
    service client otherwise, and registers any MCP servers not already present in the client's toolgroups.
    Exceptions raised by the LlamaStack client (initialization or registration errors) propagate to the caller.
    
    Parameters:
        configuration (Configuration): Configuration containing `mcp_servers` and `llama_stack.use_as_library_client`.
    """
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
    """
    Register MCP (Model Context Protocol) toolgroups that are not already present on the client.
    
    Checks which toolgroups the client already exposes and registers any servers from `mcp_servers` whose `name` is not present; network errors from the client propagate to the caller.
    
    Parameters:
        client (AsyncLlamaStackClient): The LlamaStack async client used to query and register toolgroups.
        mcp_servers (List[ModelContextProtocolServer]): MCP server descriptors to ensure are registered.
        logger (Logger): Logger used for debug messages about registration progress.
    """
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
    """
    Return a decorator that ensures an async function is executed at most once and that all callers await the same result.
    Returns:
    	decorator (Callable): A decorator which, when applied to an async function, schedules it on first call and returns the same awaited result (or propagated exception) to subsequent callers.
    """
    """
    Execute the wrapped async function's single cached task and return its awaited result to every caller.
    Returns:
    	Any: The result produced by the wrapped coroutine, or the exception it raised propagated to callers.
    """
    task = None

    @wraps(func)
    async def wrapper(*args: Any, **kwargs: Any) -> Any:
        """
        Ensure the wrapped async callable is executed only once and subsequent calls await the same result.
        
        On the first invocation the wrapped coroutine is scheduled as an asyncio.Task on the current running event loop and cached; later calls await and return the cached task's result. Exceptions raised by the task propagate to callers. Requires an active running event loop on the first call.
        
        Returns:
            The awaited result of the wrapped coroutine.
        """
        nonlocal task
        if task is None:
            loop = asyncio.get_running_loop()
            task = loop.create_task(func(*args, **kwargs))
        return await task

    return wrapper