"""Implementation of common test steps."""

from behave import given  # pyright: ignore[reportAttributeAccessIssue]
from behave.runner import Context
import os


@given("The service is started locally")
def service_is_started_locally(context: Context) -> None:
    """
    Populate the Behave context with local service endpoint values read from environment variables.
    
    Sets the following context attributes:
    - hostname: from E2E_LSC_HOSTNAME, default "localhost"
    - port: from E2E_LSC_PORT, default "8080"
    - hostname_llama: from E2E_LLAMA_HOSTNAME, default "localhost"
    - port_llama: from E2E_LLAMA_PORT, default "8321"
    
    Parameters:
        context (Context): Behave context object to receive the endpoint attributes.
    """
    assert context is not None
    context.hostname = os.getenv("E2E_LSC_HOSTNAME", "localhost")
    context.port = os.getenv("E2E_LSC_PORT", "8080")
    context.hostname_llama = os.getenv("E2E_LLAMA_HOSTNAME", "localhost")
    context.port_llama = os.getenv("E2E_LLAMA_PORT", "8321")


@given("The system is in default state")
def system_in_default_state(context: Context) -> None:
    """
    Ensure the Behave test context is present for steps that assume the system is in its default state.
    
    Parameters:
        context (Context): Behave Context instance used to store and share test state.
    
    Raises:
        AssertionError: If `context` is None.
    """
    assert context is not None