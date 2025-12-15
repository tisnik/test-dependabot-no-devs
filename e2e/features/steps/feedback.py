"""Implementation of common test steps for the feedback API."""

from behave import given, when, step  # pyright: ignore[reportAttributeAccessIssue]
from behave.runner import Context
import requests
import json
from tests.e2e.utils.utils import switch_config, restart_container
from tests.e2e.features.steps.common_http import access_rest_api_endpoint_get

# default timeout for HTTP operations
DEFAULT_TIMEOUT = 10


@step("The feedback is enabled")  # type: ignore
def enable_feedback(context: Context) -> None:
    """
    Enable the feedback endpoint.
    
    Sends a PUT request with payload {"status": True} to set feedback status to enabled and stores the HTTP response on the `context`.
    """
    assert context is not None
    payload = {"status": True}
    access_feedback_put_endpoint(context, payload)


@step("The feedback is disabled")  # type: ignore
def disable_feedback(context: Context) -> None:
    """Disable the feedback endpoint and assert success."""
    assert context is not None
    payload = {"status": False}
    access_feedback_put_endpoint(context, payload)


@when("I update feedback status with")  # type: ignore
def set_feedback(context: Context) -> None:
    """
    Update the feedback status using a JSON payload taken from the step context.
    
    Parses JSON from `context.text` and sends it to the feedback PUT endpoint.
    
    Raises:
        AssertionError: If `context.text` is not provided.
    """
    assert context.text is not None, "Payload needs to be specified"
    payload = json.loads(context.text or "{}")
    access_feedback_put_endpoint(context, payload)


def access_feedback_put_endpoint(context: Context, payload: dict) -> None:
    """
    Send a PUT request to update the feedback status.
    
    Sends the given dictionary as a JSON body to the feedback status endpoint and stores the HTTP response on context.response.
    
    Parameters:
        payload (dict): JSON-serializable payload to send in the PUT request (e.g., {"status": True}).
    """
    assert context is not None
    endpoint = "feedback/status"
    base = f"http://{context.hostname}:{context.port}"
    path = f"{context.api_prefix}/{endpoint}".replace("//", "/")
    url = base + path
    headers = context.auth_headers if hasattr(context, "auth_headers") else {}
    response = requests.put(url, headers=headers, json=payload)
    context.response = response


@when("I submit the following feedback for the conversation created before")  # type: ignore
def submit_feedback_valid_conversation(context: Context) -> None:
    """
    Submit feedback for the conversation previously created and stored on the Behave context.
    
    Raises:
        AssertionError: If the context has no `conversation_id` or it is None.
    """
    assert (
        hasattr(context, "conversation_id") and context.conversation_id is not None
    ), "Conversation for feedback submission is not created"
    access_feedback_post_endpoint(context, context.conversation_id)


@when('I submit the following feedback for nonexisting conversation "{conversation_id}"')  # type: ignore
def submit_feedback_nonexisting_conversation(
    context: Context, conversation_id: str
) -> None:
    """
    Submit feedback using the given conversation ID that is expected not to exist.
    
    Parameters:
        context (Context): Behave context carrying request configuration and where the response will be stored.
        conversation_id (str): Conversation ID to include in the feedback payload; intended to reference a non-existing conversation.
    """
    access_feedback_post_endpoint(context, conversation_id)


@when("I submit the following feedback without specifying conversation ID")  # type: ignore
def submit_feedback_without_conversation(context: Context) -> None:
    """
    Submit feedback without including a conversation identifier.
    
    Uses the step's text as the JSON payload and posts it to the feedback endpoint with no `conversation_id`; the HTTP response is stored on `context.response`.
    """
    access_feedback_post_endpoint(context, None)


def access_feedback_post_endpoint(
    context: Context, conversation_id: str | None
) -> None:
    """
    Submit the JSON payload from the step context to the server's feedback endpoint.
    
    Loads JSON from context.text (defaults to {}), injects the provided conversation_id into the payload when not None, sends the payload to the "feedback" endpoint using context.auth_headers if present, and stores the HTTP response on context.response.
    
    Parameters:
        conversation_id (str | None): Optional conversation identifier to include in the payload.
    """
    endpoint = "feedback"
    base = f"http://{context.hostname}:{context.port}"
    path = f"{context.api_prefix}/{endpoint}".replace("//", "/")
    url = base + path
    payload = json.loads(context.text or "{}")
    if conversation_id is not None:
        payload["conversation_id"] = conversation_id
    headers = context.auth_headers if hasattr(context, "auth_headers") else {}
    context.response = requests.post(url, headers=headers, json=payload)


@when("I retreive the current feedback status")  # type: ignore
def access_feedback_get_endpoint(context: Context) -> None:
    """Retrieve the current feedback status via GET request."""
    access_rest_api_endpoint_get(context, "feedback/status")


@given("A new conversation is initialized")  # type: ignore
def initialize_conversation(context: Context) -> None:
    """
    Initialize a new conversation and record its ID on the test context.
    
    Sends a request to the API to create a conversation, stores the returned `conversation_id` on
    `context.conversation_id`, appends it to `context.feedback_conversations`, and saves the HTTP
    response on `context.response`.
    
    Raises:
        AssertionError: If the API response status is not 200 or the response does not contain a
        `conversation_id`.
    """
    endpoint = "query"
    base = f"http://{context.hostname}:{context.port}"
    path = f"{context.api_prefix}/{endpoint}".replace("//", "/")
    url = base + path
    headers = context.auth_headers if hasattr(context, "auth_headers") else {}
    payload = {"query": "Say Hello.", "system_prompt": "You are a helpful assistant"}

    response = requests.post(url, headers=headers, json=payload)
    assert (
        response.status_code == 200
    ), f"Failed to create conversation: {response.text}"

    body = response.json()
    context.conversation_id = body["conversation_id"]
    assert context.conversation_id, "Conversation was not created."
    context.feedback_conversations.append(context.conversation_id)
    context.response = response


@given("An invalid feedback storage path is configured")  # type: ignore
def configure_invalid_feedback_storage_path(context: Context) -> None:
    """
    Apply the scenario's configuration to set an invalid feedback storage path and restart the test container.
    
    Parameters:
        context (Context): Behave context containing `scenario_config`; the config is applied via `switch_config` and the container named "lightspeed-stack" is restarted.
    """
    switch_config(context.scenario_config)
    restart_container("lightspeed-stack")