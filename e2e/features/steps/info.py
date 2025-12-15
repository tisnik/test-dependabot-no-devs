"""Implementation of common test steps."""

import json
import re
from behave import then  # pyright: ignore[reportAttributeAccessIssue]
from behave.runner import Context


@then("The body of the response has proper name {service_name} and version {version}")
def check_name_version(context: Context, service_name: str, version: str) -> None:
    """
    Verify the response JSON contains the expected service name and version.
    
    Asserts that the parsed response JSON exists and that its "name" equals service_name and its "service_version" equals version; assertion failures include the actual values.
    
    Parameters:
        service_name (str): Expected value for the response "name" field.
        version (str): Expected value for the response "service_version" field.
    """
    response_json = context.response.json()
    assert response_json is not None, "Response is not valid JSON"

    assert response_json["name"] == service_name, f"name is {response_json["name"]}"
    assert (
        response_json["service_version"] == version
    ), f"version is {response_json["service_version"]}"


@then("The body of the response has llama-stack version {llama_version}")
def check_llama_version(context: Context, llama_version: str) -> None:
    """
    Verify that the response's llama-stack version matches the expected semantic version.
    
    Checks the response JSON's "llama_stack_version" for a semantic version substring (major.minor.patch) and asserts it equals the provided `llama_version`.
    
    Parameters:
        llama_version (str): Expected semantic version string in the form "X.Y.Z" (e.g., "1.2.3").
    
    Raises:
        AssertionError: If the response is not valid JSON, a semantic version cannot be extracted, or the extracted version does not equal `llama_version`.
    """
    response_json = context.response.json()
    assert response_json is not None, "Response is not valid JSON"

    version_pattern = r"\d+\.\d+\.\d+"
    llama_stack_version = response_json["llama_stack_version"]
    match = re.search(version_pattern, llama_stack_version)
    assert match is not None, f"Could not extract version from {llama_stack_version}"
    extracted_version = match.group(0)

    assert (
        extracted_version == llama_version
    ), f"llama-stack version is {extracted_version}, expected {llama_version}"


@then("The body of the response has proper model structure")
def check_model_structure(context: Context) -> None:
    """
    Validate that the response contains an LLM model matching the expected provider and model, and that the model's required fields have the correct values.
    
    This step expects the HTTP response JSON to include a non-empty "models" list and uses context.default_model and context.default_provider to identify the target model. It asserts the presence of a model with api_model_type "llm" and matching provider_id/provider_resource_id, and verifies the model's fields: "type" equals "model", "api_model_type" and "model_type" equal "llm", "provider_id" and "provider_resource_id" match the expected values, and "identifier" equals "{provider}/{model}".
    """
    response_json = context.response.json()
    assert response_json is not None, "Response is not valid JSON"

    assert "models" in response_json, "Response missing 'models' field"
    models = response_json["models"]
    assert len(models) > 0, "Response has empty list of models"

    # Get expected values from context (detected in before_all)
    expected_model = context.default_model
    expected_provider = context.default_provider

    # Search for the specific model that was detected in before_all
    llm_model = None
    for model in models:
        if (
            model.get("api_model_type") == "llm"
            and model.get("provider_id") == expected_provider
            and model.get("provider_resource_id") == expected_model
        ):
            llm_model = model
            break

    assert llm_model is not None, (
        f"Expected LLM model not found in response. "
        f"Looking for provider_id='{expected_provider}' and provider_resource_id='{expected_model}'"
    )

    # Validate structure and values
    assert (
        llm_model["type"] == "model"
    ), f"type should be 'model', but is {llm_model["type"]}"
    assert (
        llm_model["api_model_type"] == "llm"
    ), f"api_model_type should be 'llm', but is {llm_model["api_model_type"]}"
    assert (
        llm_model["model_type"] == "llm"
    ), f"model_type should be 'llm', but is {llm_model["model_type"]}"
    assert (
        llm_model["provider_id"] == expected_provider
    ), f"provider_id should be '{expected_provider}', but is '{llm_model["provider_id"]}'"
    assert (
        llm_model["provider_resource_id"] == expected_model
    ), f"provider_resource_id should be '{expected_model}', but is '{llm_model["provider_resource_id"]}'"
    assert (
        llm_model["identifier"] == f"{expected_provider}/{expected_model}"
    ), f"identifier should be '{expected_provider}/{expected_model}', but is '{llm_model["identifier"]}'"


@then("The body of the response has proper shield structure")
def check_shield_structure(context: Context) -> None:
    """
    Validate the first shield entry in the response has the required structure and expected values.
    
    Checks that the response JSON contains a non-empty "shields" list, locates the first entry with "type" equal to "shield", and verifies that its "type" is "shield", "provider_id" is "llama-guard", "provider_resource_id" equals context.default_model, and "identifier" is "llama-guard-shield".
    
    Parameters:
        context (behave.runner.Context): Test context containing `response` (with .json()), and `default_model`.
    """
    response_json = context.response.json()
    assert response_json is not None, "Response is not valid JSON"

    assert "shields" in response_json, "Response missing 'shields' field"
    shields = response_json["shields"]
    assert len(shields) > 0, "Response has empty list of shields"

    # Find first shield
    found_shield = None
    for shield in shields:
        if shield.get("type") == "shield":
            found_shield = shield
            break

    assert found_shield is not None, "No shield found in response"

    expected_model = context.default_model

    # Validate structure and values
    assert found_shield["type"] == "shield", "type should be 'shield'"
    assert (
        found_shield["provider_id"] == "llama-guard"
    ), "provider_id should be 'llama-guard'"
    assert (
        found_shield["provider_resource_id"] == expected_model
    ), f"provider_resource_id should be '{expected_model}', but is '{found_shield["provider_resource_id"]}'"
    assert (
        found_shield["identifier"] == "llama-guard-shield"
    ), f"identifier should be 'llama-guard-shield', but is '{found_shield["identifier"]}'"


@then("The response contains {count:d} tools listed for provider {provider_name}")
def check_tool_count(context: Context, count: int, provider_name: str) -> None:
    """
    Verify the response contains exactly the expected number of tools for the specified provider.
    
    Parameters:
    	context (Context): Behave context whose `response.json()` contains the payload with a "tools" list.
    	count (int): Expected number of tools for the provider.
    	provider_name (str): The provider_id to filter tools by.
    """
    response_json = context.response.json()
    assert response_json is not None, "Response is not valid JSON"

    assert "tools" in response_json, "Response missing 'tools' field"
    tools = response_json["tools"]
    assert len(tools) > 0, "Response has empty list of tools"

    provider_tools = []

    for tool in tools:
        if tool["provider_id"] == provider_name:
            provider_tools.append(tool)

    assert len(provider_tools) == count


@then("The body of the response has proper structure for provider {provider_name}")
def check_tool_structure(context: Context, provider_name: str) -> None:
    """
    Validate that the first tool for the given provider matches the expected tool JSON provided in the step context.
    
    Loads expected JSON from context.text and compares it to the first tool in response.json()["tools"] whose "provider_id" equals provider_name. The comparison verifies the fields: identifier, description, provider_id, toolgroup_id, server_source, and type.
    
    Parameters:
        provider_name (str): Provider identifier used to select which tool to validate.
    """
    response_json = context.response.json()
    assert response_json is not None, "Response is not valid JSON"

    expected_json = json.loads(context.text)

    assert "tools" in response_json, "Response missing 'tools' field"
    tools = response_json["tools"]
    assert len(tools) > 0, "Response has empty list of tools"

    provider_tool = None

    for tool in tools:
        if tool["provider_id"] == provider_name:
            provider_tool = tool
            break

    assert provider_tool is not None, "No tool found in response"

    # Validate structure and values
    assert (
        provider_tool["identifier"] == expected_json["identifier"]
    ), f"identifier should be {expected_json["identifier"]}, but was {provider_tool["identifier"]}"
    assert (
        provider_tool["description"] == expected_json["description"]
    ), f"description should be {expected_json["description"]}"
    assert (
        provider_tool["provider_id"] == expected_json["provider_id"]
    ), f"provider_id should be {expected_json["provider_id"]}"
    assert (
        provider_tool["toolgroup_id"] == expected_json["toolgroup_id"]
    ), f"toolgroup_id should be {expected_json["toolgroup_id"]}"
    assert (
        provider_tool["server_source"] == expected_json["server_source"]
    ), f"server_source should be {expected_json["server_source"]}"
    assert (
        provider_tool["type"] == expected_json["type"]
    ), f"type should be {expected_json["type"]}"