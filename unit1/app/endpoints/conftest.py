"""Shared pytest fixtures for endpoint unit tests."""

from collections.abc import Callable
from typing import Any

import pytest
from pytest_mock import MockerFixture


@pytest.fixture(name="mock_request_factory")
def mock_request_factory_fixture(mocker: MockerFixture) -> Callable[..., Any]:
    """
    Create a factory that builds mock FastAPI Request objects with an optional RH Identity.
    
    The returned callable accepts an optional `rh_identity`; the created mock always has `headers` set to `{"User-Agent": "CLA/0.5.0"}`. If `rh_identity` is provided the mock's `state.rh_identity_data` is set to that value; if not provided the mock's `state` is a spec-less mock with no attributes to simulate absence of `rh_identity_data`.
    
    Parameters:
        mocker: Pytest `MockerFixture` used to create mock objects.
    
    Returns:
        A callable `create(rh_identity: Any = None) -> Any` that returns the constructed mock Request.
    """

    def _create(rh_identity: Any = None) -> Any:
        """
        Create a mocked FastAPI Request-like object with a User-Agent header and optional `state.rh_identity_data`.
        
        Parameters:
            rh_identity (Any): If provided, assigned to `mock_request.state.rh_identity_data`; if omitted, `mock_request.state` is a mock with no attributes to simulate absence of `rh_identity_data`.
        
        Returns:
            Any: A mock object representing a Request with `headers` set to {"User-Agent": "CLA/0.5.0"} and `state` configured as described.
        """
        mock_request = mocker.Mock()
        mock_request.headers = {"User-Agent": "CLA/0.5.0"}

        if rh_identity is not None:
            mock_request.state = mocker.Mock()
            mock_request.state.rh_identity_data = rh_identity
        else:
            # Use spec=[] to create a Mock with no attributes, simulating absent rh_identity_data
            mock_request.state = mocker.Mock(spec=[])

        return mock_request

    return _create


@pytest.fixture(name="mock_background_tasks")
def mock_background_tasks_fixture(mocker: MockerFixture) -> Any:
    """Create a mock BackgroundTasks object.

    Returns:
        A Mock object representing FastAPI BackgroundTasks.
    """
    return mocker.Mock()
