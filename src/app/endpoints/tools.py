"""Handler for REST API call to list available tools from MCP servers."""

import logging
from typing import Annotated, Any

from fastapi import APIRouter, Depends, HTTPException, Request, status
from llama_stack_client import APIConnectionError

from authentication import get_auth_dependency
from authentication.interface import AuthTuple
from authorization.middleware import authorize
from client import AsyncLlamaStackClientHolder
from configuration import configuration
from models.config import Action
from models.responses import ToolsResponse
from utils.endpoints import check_configuration_loaded
from utils.tool_formatter import format_tools_list

logger = logging.getLogger(__name__)
router = APIRouter(tags=["tools"])


tools_responses: dict[int | str, dict[str, Any]] = {
    200: {
        "description": "Successful Response",
        "content": {
            "application/json": {
                "example": {
                    "tools": [
                        {
                            "identifier": "",
                            "description": "",
                            "parameters": [
                                {
                                    "name": "",
                                    "description": "",
                                    "parameter_type": "",
                                    "required": "True/False",
                                    "default": "null",
                                }
                            ],
                            "provider_id": "",
                            "toolgroup_id": "",
                            "server_source": "",
                            "type": "tool",
                        }
                    ]
                }
            }
        },
    },
    500: {"description": "Connection to Llama Stack is broken or MCP server error"},
}


@router.get("/tools", responses=tools_responses)
@authorize(Action.GET_TOOLS)
async def tools_endpoint_handler(
    request: Request,
    auth: Annotated[AuthTuple, Depends(get_auth_dependency())],
) -> ToolsResponse:
    """
    Handle requests to the /tools endpoint.

    Process GET requests to the /tools endpoint, returning a consolidated list of
    available tools from all configured MCP servers.

    Raises:
        HTTPException: If unable to connect to the Llama Stack server or if
        tool retrieval fails for any reason.

    Returns:
        ToolsResponse: An object containing the consolidated list of available tools
        with metadata including tool name, description, parameters, and server source.
    """
    # Used only by the middleware
    _ = auth

    # Nothing interesting in the request
    _ = request

    check_configuration_loaded(configuration)

    try:
        # Get Llama Stack client
        client = AsyncLlamaStackClientHolder().get_client()

        consolidated_tools = []
        mcp_server_names = (
            {mcp_server.name for mcp_server in configuration.mcp_servers}
            if configuration.mcp_servers
            else set()
        )

        # Get all available toolgroups
        try:
            logger.debug("Retrieving tools from all toolgroups")
            toolgroups_response = await client.toolgroups.list()

            for toolgroup in toolgroups_response:
                try:
                    # Get tools for each toolgroup
                    tools_response = await client.tools.list(
                        toolgroup_id=toolgroup.identifier
                    )

                    # Convert tools to dict format
                    tools_count = 0
                    server_source = "unknown"

                    for tool in tools_response:
                        tool_dict = dict(tool)

                        # Determine server source based on toolgroup type
                        if toolgroup.identifier in mcp_server_names:
                            # This is an MCP server toolgroup
                            mcp_server = next(
                                (
                                    s
                                    for s in configuration.mcp_servers
                                    if s.name == toolgroup.identifier
                                ),
                                None,
                            )
                            tool_dict["server_source"] = (
                                mcp_server.url if mcp_server else toolgroup.identifier
                            )
                        else:
                            # This is a built-in toolgroup
                            tool_dict["server_source"] = "builtin"

                        consolidated_tools.append(tool_dict)
                        tools_count += 1
                        server_source = tool_dict["server_source"]

                    logger.debug(
                        "Retrieved %d tools from toolgroup %s (source: %s)",
                        tools_count,
                        toolgroup.identifier,
                        server_source,
                    )

                except Exception as e:  # pylint: disable=broad-exception-caught
                    # Catch any exception from individual toolgroup failures to allow
                    # processing of other toolgroups to continue (partial failure scenario)
                    logger.warning(
                        "Failed to retrieve tools from toolgroup %s: %s",
                        toolgroup.identifier,
                        e,
                    )
                    continue

        except APIConnectionError as e:
            logger.warning("Failed to retrieve tools from toolgroups: %s", e)
            raise
        except (ValueError, AttributeError) as e:
            logger.warning("Failed to retrieve tools from toolgroups: %s", e)

        logger.info(
            "Retrieved total of %d tools (%d from built-in toolgroups, %d from MCP servers)",
            len(consolidated_tools),
            len([t for t in consolidated_tools if t.get("server_source") == "builtin"]),
            len([t for t in consolidated_tools if t.get("server_source") != "builtin"]),
        )

        # Format tools with structured description parsing
        formatted_tools = format_tools_list(consolidated_tools)

        return ToolsResponse(tools=formatted_tools)

    # Connection to Llama Stack server
    except APIConnectionError as e:
        logger.error("Unable to connect to Llama Stack: %s", e)
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail={
                "response": "Unable to connect to Llama Stack",
                "cause": str(e),
            },
        ) from e
    # Any other exception that can occur during tool listing
    except Exception as e:
        logger.error("Unable to retrieve list of tools: %s", e)
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail={
                "response": "Unable to retrieve list of tools",
                "cause": str(e),
            },
        ) from e
