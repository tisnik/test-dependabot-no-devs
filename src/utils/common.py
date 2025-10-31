"""Common utilities for the project."""

from typing import Any, List, cast
from logging import Logger

from llama_stack_client import LlamaStackClient, AsyncLlamaStackClient

from llama_stack.distribution.library_client import (
    AsyncLlamaStackAsLibraryClient,
)

from client import LlamaStackClientHolder, AsyncLlamaStackClientHolder
from models.config import Configuration, ModelContextProtocolServer


# TODO(lucasagomes): implement this function to retrieve user ID from auth
def retrieve_user_id(auth: Any) -> str:  # pylint: disable=unused-argument
    """
    Return a placeholder user ID string.
    
    This function is intended to extract a user ID from an authentication handler, but currently returns a fixed placeholder value.
    """
    return "user_id_placeholder"


async def register_mcp_servers_async(
    logger: Logger, configuration: Configuration
) -> None:
    """
    Asynchronously registers all configured Model Context Protocol (MCP) servers with the LlamaStack client.
    
    If no MCP servers are specified in the configuration, the function exits early. Depending on the configuration, it uses either the asynchronous library client or the synchronous service client to perform the registration.
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
        # Service client - use sync interface
        client = LlamaStackClientHolder().get_client()
        _register_mcp_toolgroups_sync(client, configuration.mcp_servers, logger)


async def _register_mcp_toolgroups_async(
    client: AsyncLlamaStackClient,
    mcp_servers: List[ModelContextProtocolServer],
    logger: Logger,
) -> None:
    """
    Asynchronously registers MCP toolgroups with the provided async client for any MCP servers not already registered.
    
    For each MCP server in the list, checks if its name is absent from the currently registered toolgroups and, if so, registers it using the client's async registration method.
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


def _register_mcp_toolgroups_sync(
    client: LlamaStackClient,
    mcp_servers: List[ModelContextProtocolServer],
    logger: Logger,
) -> None:
    """
    Register MCP toolgroups with the LlamaStack client for each MCP server not already registered.
    
    For each MCP server in the provided list, checks if its name is absent from the currently registered toolgroups. If so, registers the toolgroup with the client using the server's details.
    """
    # Get registered tool groups
    registered_toolgroups = client.toolgroups.list()
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

            client.toolgroups.register(**registration_params)
            logger.debug("MCP server %s registered successfully", mcp.name)
