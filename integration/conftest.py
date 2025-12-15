"""Shared fixtures for integration tests."""

from pathlib import Path
from typing import Generator

import pytest
from fastapi import Request, Response

from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker, Session
from sqlalchemy.engine import Engine

from authentication.noop import NoopAuthDependency
from authentication.interface import AuthTuple

from configuration import configuration
from models.database.base import Base


@pytest.fixture(autouse=True)
def reset_configuration_state() -> Generator:
    """
    Reset the module-level configuration singleton before each test.
    
    This autouse pytest fixture clears the loaded configuration state so tests run with a fresh configuration context and do not leak state between each other. It yields once to allow the test to run and performs no further actions.
    """
    # pylint: disable=protected-access
    configuration._configuration = None
    yield


@pytest.fixture(name="test_config", scope="function")
def test_config_fixture() -> Generator:
    """
    Load and expose the project's real configuration file for integration tests.
    
    Yields:
        The `configuration` module with the loaded settings.
    """
    config_path = (
        Path(__file__).parent.parent / "configuration" / "lightspeed-stack.yaml"
    )
    assert config_path.exists(), f"Config file not found: {config_path}"

    # Load configuration
    configuration.load_configuration(str(config_path))

    yield configuration
    # Note: Cleanup is handled by the autouse reset_configuration_state fixture


@pytest.fixture(name="current_config", scope="function")
def current_config_fixture() -> Generator:
    """
    Load and provide the project's current configuration for integration tests.
    
    Yields the configuration object loaded from the project's lightspeed-stack.yaml at the repository root. Cleanup/reset of global configuration state is performed by the autouse reset_configuration_state fixture.
    
    Returns:
        configuration: The loaded configuration object.
    """
    config_path = Path(__file__).parent.parent.parent / "lightspeed-stack.yaml"
    assert config_path.exists(), f"Config file not found: {config_path}"

    # Load configuration
    configuration.load_configuration(str(config_path))

    yield configuration
    # Note: Cleanup is handled by the autouse reset_configuration_state fixture


@pytest.fixture(name="test_db_engine", scope="function")
def test_db_engine_fixture() -> Generator:
    """
    Create a fresh in-memory SQLite database engine and initialize its schema for a test.
    
    The fixture creates all tables before yielding and drops them and disposes the engine after the test completes.
    
    Returns:
        engine (Engine): A SQLAlchemy Engine connected to a new in-memory SQLite database.
    """
    # Create in-memory SQLite database
    engine = create_engine(
        "sqlite:///:memory:",
        echo=False,  # Set to True to see SQL queries
        connect_args={"check_same_thread": False},  # Allow multi-threaded access
    )

    # Create all tables
    Base.metadata.create_all(engine)

    yield engine

    # Cleanup
    Base.metadata.drop_all(engine)
    engine.dispose()


@pytest.fixture(name="test_db_session", scope="function")
def test_db_session_fixture(test_db_engine: Engine) -> Generator[Session, None, None]:
    """
    Provide a SQLAlchemy Session bound to the provided test Engine for use in tests.
    
    Returns:
        session (Session): A database session bound to the test engine; the fixture closes the session after the test.
    """
    session_local = sessionmaker(autocommit=False, autoflush=False, bind=test_db_engine)
    session = session_local()

    yield session

    session.close()


@pytest.fixture(name="test_request")
def test_request_fixture() -> Request:
    """
    Create a FastAPI Request with a minimal HTTP scope suitable for tests.
    
    Returns:
        request (fastapi.Request): A Request object whose scope has `"type": "http"`, an empty `query_string`, and no headers.
    """
    return Request(
        scope={
            "type": "http",
            "query_string": b"",
            "headers": [],
        }
    )


@pytest.fixture(name="test_response")
def test_response_fixture() -> Response:
    """
    Create a FastAPI Response object for tests.
    
    Returns:
        Response: Response with empty content, status 200, and media_type "application/json".
    """
    return Response(content="", status_code=200, media_type="application/json")


@pytest.fixture(name="test_auth")
async def test_auth_fixture(test_request: Request) -> AuthTuple:
    """
    Invoke the NoopAuthDependency with the provided request to obtain authentication.
    
    Returns:
        AuthTuple: Authentication information produced by NoopAuthDependency.
    """
    noop_auth = NoopAuthDependency()
    return await noop_auth(test_request)