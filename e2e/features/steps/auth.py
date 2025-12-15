"""Implementation of common test steps."""

import requests
from behave import given, when  # pyright: ignore[reportAttributeAccessIssue]
from behave.runner import Context
from tests.e2e.utils.utils import normalize_endpoint


@given("I set the Authorization header to {header_value}")
def set_authorization_header_custom(context: Context, header_value: str) -> None:
    """
    Set the Authorization header value in the Behave test context.
    
    If the context has no `auth_headers` dictionary, one is created and the given value is assigned to the `Authorization` key.
    
    Parameters:
        header_value (str): The value to set for the `Authorization` header.
    """
    if not hasattr(context, "auth_headers"):
        context.auth_headers = {}
    context.auth_headers["Authorization"] = header_value
    print(f"🔑 Set Authorization header to: {header_value}")


@given("I remove the auth header")  # type: ignore
def remove_authorization_header(context: Context) -> None:
    """
    Remove the "Authorization" header from context.auth_headers if it exists.
    """
    if hasattr(context, "auth_headers") and "Authorization" in context.auth_headers:
        del context.auth_headers["Authorization"]


@when("I access endpoint {endpoint} using HTTP POST method with user_id {user_id}")
def access_rest_api_endpoint_post(
    context: Context, endpoint: str, user_id: str
) -> None:
    """
    Send a POST request to the given endpoint including the provided user_id as a query parameter.
    
    The request URL is built from context.hostname and context.port, and the endpoint is normalized before use. The HTTP response is stored in context.response and the request uses headers from context.auth_headers (created as an empty dict if missing).
    
    Parameters:
        endpoint (str): Endpoint path to call; will be normalized.
        user_id (str): Value used for the `user_id` query parameter (surrounding quotes are removed).
    """
    endpoint = normalize_endpoint(endpoint)
    user_id = user_id.replace('"', "")
    base = f"http://{context.hostname}:{context.port}"
    path = f"{endpoint}?user_id={user_id}".replace("//", "/")
    url = base + path

    if not hasattr(context, "auth_headers"):
        context.auth_headers = {}

    # perform REST API call
    context.response = requests.post(
        url, json="", headers=context.auth_headers, timeout=10
    )


@when("I access endpoint {endpoint} using HTTP POST method without user_id")
def access_rest_api_endpoint_post_without_param(
    context: Context, endpoint: str
) -> None:
    """
    Send a POST request to the given endpoint without query parameters and store the HTTP response on `context.response`.
    """
    endpoint = normalize_endpoint(endpoint)
    base = f"http://{context.hostname}:{context.port}"
    path = f"{endpoint}".replace("//", "/")
    url = base + path

    if not hasattr(context, "auth_headers"):
        context.auth_headers = {}

    # perform REST API call
    context.response = requests.post(
        url, json="", headers=context.auth_headers, timeout=10
    )