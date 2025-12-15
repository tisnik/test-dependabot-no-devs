"""Common steps for HTTP-related operations."""

import json

import requests
from behave import then, when, step  # pyright: ignore[reportAttributeAccessIssue]
from behave.runner import Context
from tests.e2e.utils.utils import (
    normalize_endpoint,
    replace_placeholders,
    validate_json,
    validate_json_partially,
)

# default timeout for HTTP operations
DEFAULT_TIMEOUT = 10


@when(
    "I request the {endpoint} endpoint in {hostname:w}:{port:d} with {body} in the body"
)
def request_endpoint_with_body(
    context: Context, endpoint: str, hostname: str, port: int, body: str
) -> None:
    """
    Send an HTTP GET to the specified local endpoint using the provided raw body and save the response on the context.
    
    Sets context.response to the received requests.Response object.
    """
    # initial value
    context.response = None

    # perform REST API call
    context.response = requests.get(
        f"http://{hostname}:{port}/{endpoint}",
        data=body,
        timeout=DEFAULT_TIMEOUT,
    )


@when("I request the {endpoint} endpoint in {hostname:w}:{port:d} with JSON")
def request_endpoint_with_json(
    context: Context, endpoint: str, hostname: str, port: int
) -> None:
    """
    Send an HTTP GET to a local server endpoint using JSON parsed from context.text.
    
    Parses JSON from context.text and sends it as the request payload; stores the resulting Response on context.response.
    
    Raises:
        AssertionError: if context.text is None (payload must be provided).
    """
    # initial value
    context.response = None

    assert context.text is not None, "Payload needs to be specified"

    # perform REST API call
    context.response = requests.get(
        f"http://{hostname}:{port}/{endpoint}",
        json=json.loads(context.text),
        timeout=DEFAULT_TIMEOUT,
    )


@when(
    "I request the {endpoint} endpoint in {hostname:w}:{port:d} with following parameters"
)
def request_endpoint_with_url_params(
    context: Context, endpoint: str, hostname: str, port: int
) -> None:
    """
    Send an HTTP GET to the specified endpoint using query parameters built from context.table.
    
    Expects context.table to be present with rows containing 'param' and 'value' keys; those rows are converted into URL query parameters. The HTTP response is stored on context.response.
    """
    params = {}

    assert context.table is not None, "Request parameters needs to be specified"

    for row in context.table:
        name = row["param"]
        value = row["value"]
        params[name] = value

    # initial value
    context.response = None

    # perform REST API call
    context.response = requests.get(
        f"http://{hostname}:{port}/{endpoint}",
        params=params,
        timeout=DEFAULT_TIMEOUT,
    )


@when("I request the {endpoint} endpoint in {hostname:w}:{port:d} with path {path}")
def request_endpoint_with_url_path(
    context: Context, endpoint: str, hostname: str, port: int, path: str
) -> None:
    """
    Send an HTTP GET to the specified host and port, targeting the given endpoint with an additional path segment.
    
    Parameters:
        context (Context): Behave context used to store the response in `context.response`.
        endpoint (str): Endpoint path segment appended after the host and port (e.g., "api/v1/resource").
        hostname (str): Hostname of the target service.
        port (int): Port of the target service.
        path (str): Additional path segment appended after the endpoint.
    """
    # initial value
    context.response = None

    # perform REST API call
    context.response = requests.get(
        f"http://{hostname}:{port}/{endpoint}/{path}",
        timeout=DEFAULT_TIMEOUT,
    )


@when("I request the {endpoint} endpoint in {hostname:w}:{port:d}")
def request_endpoint(context: Context, endpoint: str, hostname: str, port: int) -> None:
    """
    Perform an HTTP GET to the constructed URL and store the response on the Behave context.
    
    Performs a GET request to "http://{hostname}:{port}/{endpoint}" using DEFAULT_TIMEOUT seconds and assigns the resulting requests.Response to `context.response`.
    """
    # initial value
    context.response = None

    # perform REST API call
    context.response = requests.get(
        f"http://{hostname}:{port}/{endpoint}", timeout=DEFAULT_TIMEOUT
    )


@step("The status code of the response is {status:d}")
def check_status_code(context: Context, status: int) -> None:
    """
    Assert that the last HTTP response has the given status code.
    
    If no request has been performed, the step fails. On mismatch, the failure message includes the response body (parsed as JSON when possible, otherwise raw text) to aid debugging.
    
    Parameters:
        status (int): Expected HTTP status code.
    """
    assert context.response is not None, "Request needs to be performed first"
    if context.response.status_code != status:
        # Include response body in error message for debugging
        try:
            error_body = context.response.json()
        except Exception:
            error_body = context.response.text
        assert False, (
            f"Status code is {context.response.status_code}, expected {status}. "
            f"Response: {error_body}"
        )


@then('Content type of response should be set to "{content_type}"')
def check_content_type(context: Context, content_type: str) -> None:
    """
    Validate that the most recent HTTP response's Content-Type header begins with the given value.
    
    Parameters:
        content_type (str): Expected Content-Type prefix (e.g., "application/json"); the response header must start with this value.
    """
    assert context.response is not None, "Request needs to be performed first"
    headers = context.response.headers
    assert "content-type" in headers, "Content type is not specified"
    actual = headers["content-type"]
    assert actual.startswith(content_type), f"Improper content type {actual}"


@then("The body of the response has the following schema")
def check_response_body_schema(context: Context) -> None:
    """
    Validate the JSON response body against a JSON schema provided in context.text.
    
    Asserts that a response exists and that context.text contains a schema, then parses the schema and the response body and validates them using validate_json.
    """
    assert context.response is not None, "Request needs to be performed first"
    assert context.text is not None, "Response does not contain any payload"
    schema = json.loads(context.text)
    body = context.response.json()

    validate_json(schema, body)


@then("The body of the response contains {substring}")
def check_response_body_contains(context: Context, substring: str) -> None:
    """Check that response body contains a substring."""
    assert context.response is not None, "Request needs to be performed first"
    assert (
        substring in context.response.text
    ), f"The response text '{context.response.text}' doesn't contain '{substring}'"


@then("The body of the response is the following")
def check_prediction_result(context: Context) -> None:
    """
    Assert that the HTTP response JSON exactly matches the expected JSON payload after placeholder replacement.
    
    Parses the expected JSON from context.text, applies placeholder replacement via replace_placeholders(context, ...), and compares it to context.response.json(). Raises an AssertionError if a response is missing, if no expected payload is provided, or if the actual and expected JSON objects differ; on mismatch the assertion message includes the actual and expected values.
    """
    assert context.response is not None, "Request needs to be performed first"
    assert context.text is not None, "Response does not contain any payload"

    # Replace {MODEL} and {PROVIDER} placeholders with actual values
    json_str = replace_placeholders(context, context.text or "{}")

    expected_body = json.loads(json_str)
    result = context.response.json()

    # compare both JSONs and print actual result in case of any difference
    assert result == expected_body, f"got:\n{result}\nwant:\n{expected_body}"


@then('The body of the response, ignoring the "{field}" field, is the following')
def check_prediction_result_ignoring_field(context: Context, field: str) -> None:
    """
    Assert that the JSON response body equals the expected JSON payload provided in the step text, after removing a specified field from both.
    
    Parameters:
        field (str): The key name to remove from both the expected JSON (from `context.text`) and the actual response JSON before comparison.
    """
    assert context.response is not None, "Request needs to be performed first"
    assert context.text is not None, "Response does not contain any payload"
    expected_body = json.loads(context.text).copy()
    result = context.response.json().copy()

    expected_body.pop(field, None)
    result.pop(field, None)

    # compare both JSONs and print actual result in case of any difference
    assert result == expected_body, f"got:\n{result}\nwant:\n{expected_body}"


@step("REST API service hostname is {hostname:w}")
def set_service_hostname(context: Context, hostname: str) -> None:
    """Set REST API hostname to be used in following steps."""
    context.hostname = hostname


@step("REST API service port is {port:d}")
def set_service_port(context: Context, port: int) -> None:
    """Set REST API port to be used in following steps."""
    context.port = port


@step("REST API service prefix is {prefix}")
def set_rest_api_prefix(context: Context, prefix: str) -> None:
    """
    Set the REST API path prefix used by subsequent HTTP step implementations.
    
    Parameters:
        prefix (str): The API prefix to apply to REST endpoints (stored on the test context as `api_prefix`, e.g. "api/v1" or "/api").
    """
    context.api_prefix = prefix


@when("I access endpoint {endpoint} using HTTP GET method")
def access_non_rest_api_endpoint_get(context: Context, endpoint: str) -> None:
    """
    Send an HTTP GET request to a non-REST endpoint on the configured host and port and store the response in the test context.
    
    The endpoint is normalized and combined with context.hostname and context.port to form the URL; the response is assigned to `context.response`. The request uses the module `DEFAULT_TIMEOUT`.
    """
    endpoint = normalize_endpoint(endpoint)
    base = f"http://{context.hostname}:{context.port}"
    path = f"{endpoint}".replace("//", "/")
    url = base + path
    # initial value
    context.response = None

    # perform REST API call
    context.response = requests.get(url, timeout=DEFAULT_TIMEOUT)
    assert context.response is not None, "Response is None"


@when("I access REST API endpoint {endpoint} using HTTP GET method")
def access_rest_api_endpoint_get(context: Context, endpoint: str) -> None:
    """
    Perform a GET request against the REST API endpoint using the hostname, port and api_prefix stored in the context and include any optional auth headers from context.
    
    The HTTP response is saved to context.response.
    
    Parameters:
        endpoint (str): REST endpoint path relative to the configured api_prefix (leading/trailing slashes are normalized).
    """
    endpoint = normalize_endpoint(endpoint)
    base = f"http://{context.hostname}:{context.port}"
    path = f"{context.api_prefix}/{endpoint}".replace("//", "/")
    url = base + path
    headers = context.auth_headers if hasattr(context, "auth_headers") else {}
    # initial value
    context.response = None

    # perform REST API call
    context.response = requests.get(url, headers=headers, timeout=DEFAULT_TIMEOUT)


@when("I access endpoint {endpoint} using HTTP POST method")
def access_non_rest_api_endpoint_post(context: Context, endpoint: str) -> None:
    """
    Send a POST request with a JSON payload taken from context.text to the configured non-REST endpoint.
    
    The payload is parsed from context.text (must be present). The request is sent to http://{context.hostname}:{context.port}{endpoint} and any headers in context.auth_headers are used if available. The HTTP response is stored on context.response.
    """
    endpoint = normalize_endpoint(endpoint)
    base = f"http://{context.hostname}:{context.port}"
    path = f"{endpoint}".replace("//", "/")
    url = base + path

    assert context.text is not None, "Payload needs to be specified"
    data = json.loads(context.text)
    headers = context.auth_headers if hasattr(context, "auth_headers") else {}
    # initial value
    context.response = None

    # perform REST API call
    context.response = requests.post(
        url, json=data, headers=headers, timeout=DEFAULT_TIMEOUT
    )


@when("I access REST API endpoint {endpoint} using HTTP POST method")
def access_rest_api_endpoint_post(context: Context, endpoint: str) -> None:
    """
    Send a POST request with a JSON body to the REST API endpoint.
    
    The JSON payload is read from `context.text` and parsed with `json.loads`; `context.text` must not be None.
    If `context` has `auth_headers`, they are used as the request headers. The received response is stored in
    `context.response`.
    
    Raises:
        AssertionError: If `context.text` is None.
    """
    endpoint = normalize_endpoint(endpoint)
    base = f"http://{context.hostname}:{context.port}"
    path = f"{context.api_prefix}/{endpoint}".replace("//", "/")
    url = base + path

    assert context.text is not None, "Payload needs to be specified"
    data = json.loads(context.text)
    headers = context.auth_headers if hasattr(context, "auth_headers") else {}
    # initial value
    context.response = None

    # perform REST API call
    context.response = requests.post(
        url, json=data, headers=headers, timeout=DEFAULT_TIMEOUT
    )


@when("I access REST API endpoint {endpoint} using HTTP PUT method")
def access_rest_api_endpoint_put(context: Context, endpoint: str) -> None:
    """
    Send a PUT request with a JSON payload taken from context.text to the REST API endpoint constructed from context.hostname, context.port and context.api_prefix, and store the HTTP response in context.response.
    
    Asserts that context.text is provided; parses it as JSON for the request body. If present, uses context.auth_headers for request headers.
    """
    endpoint = normalize_endpoint(endpoint)
    base = f"http://{context.hostname}:{context.port}"
    path = f"{context.api_prefix}/{endpoint}".replace("//", "/")
    url = base + path

    assert context.text is not None, "Payload needs to be specified"
    data = json.loads(context.text)
    headers = context.auth_headers if hasattr(context, "auth_headers") else {}
    # initial value
    context.response = None

    # perform REST API call
    context.response = requests.put(
        url, json=data, headers=headers, timeout=DEFAULT_TIMEOUT
    )


@then('The status message of the response is "{expected_message}"')
def check_status_of_response(context: Context, expected_message: str) -> None:
    """
    Validate that the JSON response contains a "status" field equal to the expected message.
    
    Checks that a response is present, that its body parses as JSON and contains a "status" key, and asserts the value equals expected_message.
    
    Parameters:
        expected_message (str): The expected value of the response's "status" field.
    
    Raises:
        AssertionError: If no response has been recorded, the body is not valid JSON, the "status" key is missing, or its value does not match expected_message.
    """
    assert context.response is not None, "Send request to service first"

    # try to parse response body as JSON
    body = context.response.json()
    assert body is not None, "Improper format of response body"

    assert "status" in body, "Response does not contain status message"
    actual_message = body["status"]

    assert (
        actual_message == expected_message
    ), f"Improper status message {actual_message}"


@then("I should see attribute named {attribute:w} in response")
def check_attribute_presence(context: Context, attribute: str) -> None:
    """
    Assert that the given key is present in the JSON body of the last HTTP response.
    
    Parameters:
        attribute (str): The JSON key expected to be present in the response body.
    
    Raises:
        AssertionError: If no request has been performed, if the response body is not valid JSON, or if `attribute` is not present in the JSON.
    """
    assert context.response is not None, "Request needs to be performed first"
    json = context.response.json()
    assert json is not None

    assert attribute in json, f"Attribute {attribute} is not returned by the service"


@then("Attribute {attribute:w} should be null")
def check_for_null_attribute(context: Context, attribute: str) -> None:
    """
    Verify that the specified attribute in the JSON response has a null value.
    
    Parameters:
        attribute (str): Name of the JSON attribute to check.
    
    Raises:
        AssertionError: If no prior response exists, the response body is not JSON, the attribute is missing, or the attribute's value is not null.
    """
    assert context.response is not None, "Request needs to be performed first"
    json = context.response.json()
    assert json is not None

    assert attribute in json, f"Attribute {attribute} is not returned by the service"
    value = json[attribute]
    assert (
        value is None
    ), f"Attribute {attribute} should be null, but it contains {value}"


@then("the body of the response has the following structure")
def check_response_partially(context: Context) -> None:
    """Validate that the response body matches the expected JSON structure.

    Compares the actual response JSON against the expected structure defined
    in `context.text`, ignoring extra keys or values not specified.
    """
    assert context.response is not None, "Request needs to be performed first"
    body = context.response.json()
    expected = json.loads(context.text or "{}")
    validate_json_partially(body, expected)