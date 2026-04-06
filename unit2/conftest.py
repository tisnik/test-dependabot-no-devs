"""Shared pytest fixtures for unit tests."""

from __future__ import annotations

from collections.abc import Generator

import pytest
from pytest_mock import AsyncMockType, MockerFixture

from configuration import AppConfig

type AgentFixtures = Generator[
    tuple[
        AsyncMockType,
        AsyncMockType,
    ],
    None,
    None,
]


@pytest.fixture(name="prepare_agent_mocks", scope="function")
def prepare_agent_mocks_fixture(
    mocker: MockerFixture,
) -> AgentFixtures:
    """
    Provide configured AsyncMock objects for an LLM client and agent for tests.
    
    Returns:
        tuple: (mock_client, mock_agent) — two AsyncMock objects representing the client and the agent. The mock agent has `agent_id` and `_agent_id` set to a test value and its `create_turn.return_value.steps` initialized to an empty list to prevent agent-initialization errors.
    """
    mock_client = mocker.AsyncMock()
    mock_agent = mocker.AsyncMock()

    # Set up agent_id property to avoid "Agent ID not initialized" error
    mock_agent._agent_id = "test_agent_id"  # pylint: disable=protected-access
    mock_agent.agent_id = "test_agent_id"

    # Set up create_turn mock structure for query endpoints that need it
    mock_agent.create_turn.return_value.steps = []

    yield mock_client, mock_agent


@pytest.fixture(name="minimal_config")
def minimal_config_fixture() -> AppConfig:
    """
    Create a minimal AppConfig containing only the required fields for tests.
    
    The returned configuration is initialized with a small set of keys used by tests: name, service (host/port), llama_stack (api_key, url, use_as_library_client), user_data_collection, authentication (module), and authorization (access_rules).
    
    Returns:
        AppConfig: An AppConfig instance initialized with the minimal required fields.
    """
    cfg = AppConfig()
    cfg.init_from_dict(
        {
            "name": "test",
            "service": {"host": "localhost", "port": 8080},
            "llama_stack": {
                "api_key": "test-key",
                "url": "http://test.com:1234",
                "use_as_library_client": False,
            },
            "user_data_collection": {},
            "authentication": {"module": "noop"},
            "authorization": {"access_rules": []},
        }
    )
    return cfg
