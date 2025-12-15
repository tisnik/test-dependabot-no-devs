"""Implementation of common test steps."""

import json
from behave import (
    step,
    when,
    then,
    given,
)  # pyright: ignore[reportAttributeAccessIssue]
from behave.runner import Context
import requests
from tests.e2e.utils.utils import replace_placeholders, restart_container, switch_config

# default timeout for HTTP operations
DEFAULT_TIMEOUT = 10


@step(
    "I use REST API conversation endpoint with conversation_id from above using HTTP GET method"
)
def access_conversation_endpoint_get(context: Context) -> None:
    """
    Request the conversation identified by the previously stored `conversation_id` and store the HTTP response on `context.response`.
    
    Asserts that `context.response_data["conversation_id"]` is present, constructs the conversation GET URL using `context.hostname`, `context.port`, and `context.api_prefix`, and performs the request using optional `context.auth_headers`.
    """
    assert (
        context.response_data["conversation_id"] is not None
    ), "conversation id not stored"
    endpoint = "conversations"
    base = f"http://{context.hostname}:{context.port}"
    path = f"{context.api_prefix}/{endpoint}/{context.response_data['conversation_id']}".replace(
        "//", "/"
    )
    url = base + path
    headers = context.auth_headers if hasattr(context, "auth_headers") else {}
    # initial value
    context.response = None

    # perform REST API call
    context.response = requests.get(url, headers=headers, timeout=DEFAULT_TIMEOUT)


@step(
    'I use REST API conversation endpoint with conversation_id "{conversation_id}" using HTTP GET method'
)
def access_conversation_endpoint_get_specific(
    context: Context, conversation_id: str
) -> None:
    """
    Fetch a conversation by its conversation_id from the API.
    
    Parameters:
        conversation_id (str): The identifier of the conversation to retrieve.
    """
    endpoint = "conversations"
    base = f"http://{context.hostname}:{context.port}"
    path = f"{context.api_prefix}/{endpoint}/{conversation_id}".replace("//", "/")
    url = base + path
    headers = context.auth_headers if hasattr(context, "auth_headers") else {}
    # initial value
    context.response = None

    # perform REST API call
    context.response = requests.get(url, headers=headers, timeout=DEFAULT_TIMEOUT)


@when(
    "I use REST API conversation endpoint with conversation_id from above using HTTP DELETE method"
)
def access_conversation_endpoint_delete(context: Context) -> None:
    """
    Delete the conversation identified by context.response_data['conversation_id'] using the service's REST API.
    
    Sends an HTTP DELETE to the conversations endpoint constructed from context.hostname, context.port and context.api_prefix, using context.auth_headers if present. The HTTP response object is stored on context.response for later assertions.
    
    Raises:
        AssertionError: if context.response_data['conversation_id'] is not present.
    """
    assert (
        context.response_data["conversation_id"] is not None
    ), "conversation id not stored"
    endpoint = "conversations"
    base = f"http://{context.hostname}:{context.port}"
    path = f"{context.api_prefix}/{endpoint}/{context.response_data['conversation_id']}".replace(
        "//", "/"
    )
    url = base + path
    headers = context.auth_headers if hasattr(context, "auth_headers") else {}
    # initial value
    context.response = None

    # perform REST API call
    context.response = requests.delete(url, headers=headers, timeout=DEFAULT_TIMEOUT)


@step(
    'I use REST API conversation endpoint with conversation_id "{conversation_id}" using HTTP DELETE method'
)
def access_conversation_endpoint_delete_specific(
    context: Context, conversation_id: str
) -> None:
    """
    Send an HTTP DELETE to the conversation endpoint for the given conversation ID and store the response on the context.
    
    Uses hostname, port, and api_prefix from the provided context to construct the URL, and includes context.auth_headers if present. The HTTP response is assigned to context.response.
    
    Parameters:
        conversation_id (str): The conversation identifier to delete.
    """
    endpoint = "conversations"
    base = f"http://{context.hostname}:{context.port}"
    path = f"{context.api_prefix}/{endpoint}/{conversation_id}".replace("//", "/")
    url = base + path
    headers = context.auth_headers if hasattr(context, "auth_headers") else {}
    # initial value
    context.response = None

    # perform REST API call
    context.response = requests.delete(url, headers=headers, timeout=DEFAULT_TIMEOUT)


@when(
    'I use REST API conversation endpoint with conversation_id from above and topic_summary "{topic_summary}" using HTTP PUT method'
)
def access_conversation_endpoint_put(context: Context, topic_summary: str) -> None:
    """
    Update a conversation's topic_summary via the conversation PUT endpoint.
    
    Requires that context.response_data contains a valid "conversation_id". The function sends a PUT request to the conversation resource and stores the HTTP response on context.response. If `topic_summary` equals "<EMPTY>", it is converted to an empty string before sending.
    
    Parameters:
        topic_summary (str): The new topic summary to set for the conversation; use "<EMPTY>" to send an empty string.
    """
    assert hasattr(context, "response_data"), "response_data not found in context"
    assert context.response_data.get("conversation_id"), "conversation id not stored"

    endpoint = "conversations"
    base = f"http://{context.hostname}:{context.port}"
    path = f"{context.api_prefix}/{endpoint}/{context.response_data['conversation_id']}".replace(
        "//", "/"
    )
    url = base + path
    headers = context.auth_headers if hasattr(context, "auth_headers") else {}
    context.response = None

    if topic_summary == "<EMPTY>":
        topic_summary = ""

    payload = {"topic_summary": topic_summary}

    context.response = requests.put(
        url, json=payload, headers=headers, timeout=DEFAULT_TIMEOUT
    )


@step(
    'I use REST API conversation endpoint with conversation_id "{conversation_id}" and topic_summary "{topic_summary}" using HTTP PUT method'
)
def access_conversation_endpoint_put_specific(
    context: Context, conversation_id: str, topic_summary: str
) -> None:
    """
    Update a specific conversation's topic summary via the conversation REST endpoint.
    
    Parameters:
        conversation_id (str): Identifier of the conversation to update.
        topic_summary (str): Topic summary to apply to the conversation.
    """
    endpoint = "conversations"
    base = f"http://{context.hostname}:{context.port}"
    path = f"{context.api_prefix}/{endpoint}/{conversation_id}".replace("//", "/")
    url = base + path
    headers = context.auth_headers if hasattr(context, "auth_headers") else {}
    context.response = None

    payload = {"topic_summary": topic_summary}

    context.response = requests.put(
        url, json=payload, headers=headers, timeout=DEFAULT_TIMEOUT
    )


@when(
    "I use REST API conversation endpoint with conversation_id from above and empty topic_summary using HTTP PUT method"
)
def access_conversation_endpoint_put_empty(context: Context) -> None:
    """
    Send a PUT request that updates the conversation's `topic_summary` to an empty string to exercise validation.
    
    The conversation to update is taken from `context.response_data['conversation_id']`; the HTTP response is stored in `context.response`.
    """
    assert hasattr(context, "response_data"), "response_data not found in context"
    assert context.response_data.get("conversation_id"), "conversation id not stored"

    endpoint = "conversations"
    base = f"http://{context.hostname}:{context.port}"
    path = f"{context.api_prefix}/{endpoint}/{context.response_data['conversation_id']}".replace(
        "//", "/"
    )
    url = base + path
    headers = context.auth_headers if hasattr(context, "auth_headers") else {}
    context.response = None

    payload = {"topic_summary": ""}

    context.response = requests.put(
        url, json=payload, headers=headers, timeout=DEFAULT_TIMEOUT
    )


@then("The conversation with conversation_id from above is returned")
def check_returned_conversation_id(context: Context) -> None:
    """
    Finds and stores the conversation object from the response that matches the expected conversation_id.
    
    Sets context.found_conversation to the matching conversation dictionary. Raises an AssertionError if no conversation with context.response_data["conversation_id"] is present in response_json["conversations"].
    """
    response_json = context.response.json()
    found_conversation = None
    for conversation in response_json["conversations"]:
        if conversation["conversation_id"] == context.response_data["conversation_id"]:
            found_conversation = conversation
            break

    context.found_conversation = found_conversation

    assert found_conversation is not None, "conversation not found"


@then("The conversation has topic_summary and last_message_timestamp")
def check_conversation_metadata_not_empty(context: Context) -> None:
    """
    Verify the found conversation contains required metadata fields.
    
    Checks that a conversation is present on the context and that it includes a numeric `last_message_timestamp` value greater than zero and a `topic_summary` field that is not None.
    """
    found_conversation = context.found_conversation

    assert found_conversation is not None, "conversation not found in context"

    assert (
        "last_message_timestamp" in found_conversation
    ), "last_message_timestamp field missing"
    timestamp = found_conversation["last_message_timestamp"]
    assert isinstance(
        timestamp, (int, float)
    ), f"last_message_timestamp should be a number, got {type(timestamp)}"
    assert timestamp > 0, f"last_message_timestamp should be positive, got {timestamp}"

    assert "topic_summary" in found_conversation, "topic_summary field missing"
    topic_summary = found_conversation["topic_summary"]
    assert topic_summary is not None, "topic_summary should not be None"


@then('The conversation topic_summary is "{expected_summary}"')
def check_conversation_topic_summary(context: Context, expected_summary: str) -> None:
    """
    Assert that the located conversation's topic_summary equals the expected value.
    
    Checks that a conversation has been found on the context and that its "topic_summary"
    field exactly matches `expected_summary`. Raises an AssertionError if the conversation
    is missing, the field is absent, or the values differ.
    
    Parameters:
        expected_summary (str): The expected topic summary to compare against.
    """
    found_conversation = context.found_conversation

    assert found_conversation is not None, "conversation not found in context"
    assert "topic_summary" in found_conversation, "topic_summary field missing"

    actual_summary = found_conversation["topic_summary"]
    assert (
        actual_summary == expected_summary
    ), f"Expected topic_summary '{expected_summary}', but got '{actual_summary}'"


@then("The conversation details are following")
def check_returned_conversation_content(context: Context) -> None:
    """
    Validate that the located conversation's `last_used_model`, `last_used_provider`, and `message_count` match the expected values provided in the step text.
    
    The expected values are read from JSON in `context.text` after placeholder replacement. Raises `AssertionError` if any field does not match.
    """
    json_str = replace_placeholders(context, context.text or "{}")

    expected_data = json.loads(json_str)
    found_conversation = context.found_conversation

    assert (
        found_conversation["last_used_model"] == expected_data["last_used_model"]
    ), f"last_used_model mismatch, was {found_conversation["last_used_model"]}"
    assert (
        found_conversation["last_used_provider"] == expected_data["last_used_provider"]
    ), f"last_used_provider mismatch, was {found_conversation["last_used_provider"]}"
    assert (
        found_conversation["message_count"] == expected_data["message_count"]
    ), f"message count mismatch, was {found_conversation["message_count"]}"


@then("The returned conversation details have expected conversation_id")
def check_found_conversation_id(context: Context) -> None:
    """
    Verify that the response contains the expected conversation_id.
    
    Raises:
        AssertionError: If the conversation_id in the response does not match the expected value in context.response_data.
    """
    response_json = context.response.json()

    assert (
        response_json["conversation_id"] == context.response_data["conversation_id"]
    ), "found wrong conversation"


@then("The body of the response has following messages")
def check_found_conversation_content(context: Context) -> None:
    """
    Validate that the first two messages in the conversation match expected content and types.
    
    Parses expected values from the step text (JSON) and asserts:
    - the first message's `content` equals `content` and its `type` equals `type`;
    - the second message's `content` contains `content_response` and its `type` equals `type_response`.
    
    Raises:
        AssertionError: if any of the assertions fail.
    """
    expected_data = json.loads(context.text)
    response_json = context.response.json()
    chat_messages = response_json["chat_history"][0]["messages"]

    assert chat_messages[0]["content"] == expected_data["content"]
    assert chat_messages[0]["type"] == expected_data["type"]
    assert (
        expected_data["content_response"] in chat_messages[1]["content"]
    ), f"expected substring not in response, has {chat_messages[1]["content"]}"
    assert chat_messages[1]["type"] == expected_data["type_response"]


@then("The conversation with details and conversation_id from above is not found")
def check_deleted_conversation(context: Context) -> None:
    """Check whether the deleted conversation is gone."""
    assert context.response is not None


@then("The conversation history contains {count:d} messages")
def check_conversation_message_count(context: Context, count: int) -> None:
    """
    Ensure the conversation's `chat_history` contains exactly the given number of entries.
    
    Parameters:
        count (int): Expected number of entries in the response's `chat_history`.
    
    Raises:
        AssertionError: If `chat_history` is missing or its length does not equal `count`.
    """
    response_json = context.response.json()

    assert "chat_history" in response_json, "chat_history not found in response"
    actual_count = len(response_json["chat_history"])

    assert actual_count == count, (
        f"Expected {count} messages in conversation history, "
        f"but found {actual_count}"
    )


@then("The conversation history has correct metadata")
def check_conversation_metadata(context: Context) -> None:
    """
    Validate that the conversation's chat_history contains required metadata and correctly structured messages.
    
    Asserts that `chat_history` exists and is not empty. For each turn, asserts presence of `provider`, `model`, `messages`, `started_at`, and `completed_at`, that `provider` and `model` are not empty, that `messages` contains exactly two entries, and that the first message is a user message and the second is an assistant message with a `content` field.
    """
    response_json = context.response.json()

    assert "chat_history" in response_json, "chat_history not found in response"
    chat_history = response_json["chat_history"]

    assert len(chat_history) > 0, "chat_history is empty"

    for idx, turn in enumerate(chat_history):
        assert "provider" in turn, f"Turn {idx} missing 'provider'"
        assert "model" in turn, f"Turn {idx} missing 'model'"
        assert "messages" in turn, f"Turn {idx} missing 'messages'"
        assert "started_at" in turn, f"Turn {idx} missing 'started_at'"
        assert "completed_at" in turn, f"Turn {idx} missing 'completed_at'"

        assert turn["provider"], f"Turn {idx} has empty provider"
        assert turn["model"], f"Turn {idx} has empty model"

        messages = turn["messages"]
        assert (
            len(messages) == 2
        ), f"Turn {idx} should have 2 messages (user + assistant)"

        user_msg = messages[0]
        assert user_msg["type"] == "user", f"Turn {idx} first message should be user"
        assert "content" in user_msg, f"Turn {idx} user message missing content"

        assistant_msg = messages[1]
        assert (
            assistant_msg["type"] == "assistant"
        ), f"Turn {idx} second message should be assistant"
        assert (
            "content" in assistant_msg
        ), f"Turn {idx} assistant message missing content"


@then("The conversation uses model {model} and provider {provider}")
def check_conversation_model_provider(
    context: Context, model: str, provider: str
) -> None:
    """
    Verify every turn in the response's `chat_history` uses the expected model and provider.
    
    Replaces placeholders in the provided `model` and `provider` using the test context, then checks each turn's `model` and `provider` fields match the resolved expectations. Raises AssertionError if `chat_history` is missing or empty, or if any turn's model/provider differs from the expected values.
    
    Parameters:
        model (str): Expected model name (placeholders will be replaced from `context`).
        provider (str): Expected provider name (placeholders will be replaced from `context`).
    """
    response_json = context.response.json()

    assert "chat_history" in response_json, "chat_history not found in response"
    chat_history = response_json["chat_history"]

    assert len(chat_history) > 0, "chat_history is empty"

    expected_model = replace_placeholders(context, model)
    expected_provider = replace_placeholders(context, provider)

    for idx, turn in enumerate(chat_history):
        actual_model = turn.get("model")
        actual_provider = turn.get("provider")

        assert (
            actual_model == expected_model
        ), f"Turn {idx} expected model '{expected_model}', got '{actual_model}'"
        assert (
            actual_provider == expected_provider
        ), f"Turn {idx} expected provider '{expected_provider}', got '{actual_provider}'"


@given("An invalid conversation cache path is configured")  # type: ignore
def configure_invalid_conversation_cache_path(context: Context) -> None:
    """Set an invalid conversation cache path and restart the container."""
    switch_config(context.scenario_config)
    restart_container("lightspeed-stack")