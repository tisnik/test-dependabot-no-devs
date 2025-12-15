"""LLM query and response steps."""

import json
import requests
from behave import then, step  # pyright: ignore[reportAttributeAccessIssue]
from behave.runner import Context
from tests.e2e.utils.utils import replace_placeholders


DEFAULT_LLM_TIMEOUT = 60


@step("I wait for the response to be completed")
def wait_for_complete_response(context: Context) -> None:
    """
    Block until the streaming Server-Sent Events (SSE) response is complete and valid.
    
    Parses the streaming response text, stores the parsed result in context.response_data, calls raise_for_status() on context.response to surface HTTP errors, and asserts that the parsed response indicates completion.
    
    Parameters:
        context (behave.runner.Context): Behave context containing the HTTP response in `context.response`.
    """
    context.response_data = _parse_streaming_response(context.response.text)
    print(context.response_data)
    context.response.raise_for_status()
    assert context.response_data["finished"] is True


@step('I use "{endpoint}" to ask question')
def ask_question(context: Context, endpoint: str) -> None:
    """
    Send a question payload to the LLM REST API endpoint and store the HTTP response on the Behave context.
    
    The request body is produced from the current scenario text with model/provider placeholders replaced; the resulting JSON is POSTed to the constructed service URL. The HTTP response object is assigned to `context.response`.
    
    Parameters:
        endpoint (str): API endpoint path (appended to the configured api_prefix) to which the question is posted.
    """
    base = f"http://{context.hostname}:{context.port}"
    path = f"{context.api_prefix}/{endpoint}".replace("//", "/")
    url = base + path

    # Replace {MODEL} and {PROVIDER} placeholders with actual values
    json_str = replace_placeholders(context, context.text or "{}")

    data = json.loads(json_str)
    print(f"Request data: {data}")
    context.response = requests.post(url, json=data, timeout=DEFAULT_LLM_TIMEOUT)


@step('I use "{endpoint}" to ask question with authorization header')
def ask_question_authorized(context: Context, endpoint: str) -> None:
    """
    Send a question payload to the specified API endpoint using authorization headers.
    
    Builds a JSON request body from the current scenario text (with placeholders replaced), performs an HTTP POST to the constructed URL using `context.auth_headers`, and stores the HTTP response on `context.response`.
    
    Parameters:
        endpoint (str): Path segment of the API endpoint to call (appended to the configured API prefix and host).
    
    """
    base = f"http://{context.hostname}:{context.port}"
    path = f"{context.api_prefix}/{endpoint}".replace("//", "/")
    url = base + path

    # Replace {MODEL} and {PROVIDER} placeholders with actual values
    json_str = replace_placeholders(context, context.text or "{}")

    data = json.loads(json_str)
    print(f"Request data: {data}")
    context.response = requests.post(
        url, json=data, headers=context.auth_headers, timeout=DEFAULT_LLM_TIMEOUT
    )


@step("I store conversation details")
def store_conversation_details(context: Context) -> None:
    """
    Parse the HTTP response body and store it on the Behave context.
    
    Parses JSON from context.response.text and assigns the resulting object to context.response_data.
    """
    context.response_data = json.loads(context.response.text)


@step('I use "{endpoint}" to ask question with same conversation_id')
def ask_question_in_same_conversation(context: Context, endpoint: str) -> None:
    """
    Send a question to the specified API endpoint reusing the existing conversation ID.
    
    Builds the request URL from the test context, replaces placeholders in the scenario text to form the JSON payload, injects context.response_data["conversation_id"] into the payload, includes context.auth_headers if present, performs an HTTP POST, and stores the resulting HTTP response on context.response.
    
    Parameters:
        context (Context): Behave test context containing hostname, port, api_prefix, text, response_data, and optionally auth_headers.
        endpoint (str): Path segment of the API endpoint to call (appended to context.api_prefix).
    """
    base = f"http://{context.hostname}:{context.port}"
    path = f"{context.api_prefix}/{endpoint}".replace("//", "/")
    url = base + path

    # Replace {MODEL} and {PROVIDER} placeholders with actual values
    json_str = replace_placeholders(context, context.text or "{}")

    data = json.loads(json_str)
    headers = context.auth_headers if hasattr(context, "auth_headers") else {}
    data["conversation_id"] = context.response_data["conversation_id"]

    print(f"Request data: {data}")
    context.response = requests.post(
        url, json=data, headers=headers, timeout=DEFAULT_LLM_TIMEOUT
    )


@then("The response should have proper LLM response format")
def check_llm_response_format(context: Context) -> None:
    """
    Validate that the stored HTTP response contains required LLM fields.
    
    Asserts that context.response is present and that its JSON body includes the keys "conversation_id" and "response".
    """
    assert context.response is not None
    response_json = context.response.json()
    assert "conversation_id" in response_json
    assert "response" in response_json


@then("The response should not be truncated")
def check_llm_response_not_truncated(context: Context) -> None:
    """Check that the response from LLM is not truncated."""
    assert context.response is not None
    response_json = context.response.json()
    assert response_json["truncated"] is False


@then("The response should contain following fragments")
def check_fragments_in_response(context: Context) -> None:
    """Check that all specified fragments are present in the LLM response.

    First checks that the HTTP response exists and contains a
    "response" field. For each fragment listed in the scenario's
    table under "Fragments in LLM response", asserts that it
    appears as a substring in the LLM's response. Raises an
    assertion error if any fragment is missing or if the fragments
    table is not provided.
    """
    assert context.response is not None
    response_json = context.response.json()
    response = response_json["response"]

    assert context.table is not None, "Fragments are not specified in table"

    for fragment in context.table:
        expected = fragment["Fragments in LLM response"]
        assert (
            expected in response
        ), f"Fragment '{expected}' not found in LLM response: '{response}'"


@then("The streamed response should contain following fragments")
def check_streamed_fragments_in_response(context: Context) -> None:
    """
    Verify that each fragment listed in the scenario table appears in the streamed LLM response.
    
    Requires context.response_data["response_complete"] to be non-None and uses context.response_data["response"] as the streamed content. Expects a table in the scenario with a column named "Fragments in LLM response"; raises an AssertionError if the table is missing or any fragment is not found.
    """
    assert context.response_data["response_complete"] is not None
    response = context.response_data["response"]

    assert context.table is not None, "Fragments are not specified in table"

    for fragment in context.table:
        expected = fragment["Fragments in LLM response"]
        assert (
            expected in response
        ), f"Fragment '{expected}' not found in LLM response: '{response}'"


@then("The streamed response is equal to the full response")
def compare_streamed_responses(context: Context) -> None:
    """
    Verify that the streamed partial response matches the final complete response and is not empty.
    
    Asserts that context.response_data contains non-null "response" and "response_complete", that the streamed response is not an empty string, and that both strings are equal.
    """
    assert context.response_data["response"] is not None
    assert context.response_data["response_complete"] is not None

    response = context.response_data["response"]
    complete_response = context.response_data["response_complete"]

    assert response != "", "response is empty"
    assert (
        response == complete_response
    ), f"{response} and {complete_response} do not match"


def _parse_streaming_response(response_text: str) -> dict:
    """
    Parse a Server-Sent Events (SSE) streaming response and reconstruct the conversation output.
    
    Processes lines prefixed with "data: " containing JSON SSE payloads and extracts the conversation ID, accumulated streaming tokens, the final completed response when provided, and whether the stream finished.
    
    Parameters:
        response_text (str): Raw SSE response text containing newline-separated "data: " JSON payloads.
    
    Returns:
        dict: A dictionary with keys:
            - "conversation_id": The conversation identifier if a "start" event was present, otherwise None.
            - "response": The accumulated streamed response reconstructed from "token" events (concatenated).
            - "response_complete": The final complete response from a "turn_complete" event, or an empty string if absent.
            - "finished": `True` if an "end" event was received, `False` otherwise.
    """
    lines = response_text.strip().split("\n")
    conversation_id = None
    full_response = ""
    full_response_split = []
    finished = False
    first_token = True

    for line in lines:
        if line.startswith("data: "):
            try:
                data = json.loads(line[6:])  # Remove 'data: ' prefix
                event = data.get("event")

                if event == "start":
                    conversation_id = data["data"]["conversation_id"]
                elif event == "token":
                    # Skip the first token (shield status message)
                    if first_token:
                        first_token = False
                        continue
                    full_response_split.append(data["data"]["token"])
                elif event == "turn_complete":
                    full_response = data["data"]["token"]
                elif event == "end":
                    finished = True
            except json.JSONDecodeError:
                continue  # Skip malformed lines

    return {
        "conversation_id": conversation_id,
        "response": "".join(full_response_split),
        "response_complete": full_response,
        "finished": finished,
    }