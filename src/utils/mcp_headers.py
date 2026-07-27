"""MCP headers handling."""

import json
import logging
from urllib.parse import urlparse

from fastapi import Request

from configuration import AppConfig


logger = logging.getLogger("app.endpoints.dependencies")


async def mcp_headers_dependency(_request: Request) -> dict[str, dict[str, str]]:
    """
    Asynchronously retrieves MCP headers from a FastAPI request for use with MCP servers.
    
    Returns:
        A dictionary mapping MCP URL paths or toolgroup names to their respective headers, or an empty dictionary if the headers are missing or invalid.
    """
    return extract_mcp_headers(_request)


def extract_mcp_headers(request: Request) -> dict[str, dict[str, str]]:
    """
    Extracts and parses the "MCP-HEADERS" HTTP header from a FastAPI request as a dictionary.
    
    If the header is missing, empty, not valid JSON, or not a dictionary, returns an empty dictionary.
    """
    mcp_headers_string = request.headers.get("MCP-HEADERS", "")
    mcp_headers = {}
    if mcp_headers_string:
        try:
            mcp_headers = json.loads(mcp_headers_string)
        except json.decoder.JSONDecodeError as e:
            logger.error("MCP headers decode error: %s", e)

        if not isinstance(mcp_headers, dict):
            logger.error(
                "MCP headers wrong type supplied (mcp headers must be a dictionary), "
                "but type %s was supplied",
                type(mcp_headers),
            )
            mcp_headers = {}
    return mcp_headers


def handle_mcp_headers_with_toolgroups(
    mcp_headers: dict[str, dict[str, str]], config: AppConfig
) -> dict[str, dict[str, str]]:
    """
    Convert MCP headers with toolgroup names to use MCP server URLs as keys.
    
    For each entry in the input dictionary, if the key is a valid HTTP/HTTPS URL, it is retained. If the key is a toolgroup name, it is replaced with the corresponding MCP server URL from the application configuration. Entries with unknown toolgroup names are omitted.
    
    Parameters:
        mcp_headers (dict[str, dict[str, str]]): MCP headers with keys as URLs or toolgroup names.
    
    Returns:
        dict[str, dict[str, str]]: MCP headers with only URLs as keys and their associated headers as values.
    """
    converted_mcp_headers = {}

    for key, item in mcp_headers.items():
        key_url_parsed = urlparse(key)
        if key_url_parsed.scheme in ("http", "https") and key_url_parsed.netloc:
            # a valid url is supplied, deliver it as is
            converted_mcp_headers[key] = item
        else:
            # assume the key is a toolgroup name
            # look for toolgroups name in mcp_servers configuration
            # if the mcp server is not found, the mcp header gets ignored
            for mcp_server in config.mcp_servers:
                if mcp_server.name == key and mcp_server.url:
                    converted_mcp_headers[mcp_server.url] = item
                    break

    return converted_mcp_headers
