"""Unit tests for the /rags REST API endpoints."""

from pathlib import Path
from typing import Any

import pytest
from fastapi import HTTPException, Request, status
from llama_stack_client import APIConnectionError, BadRequestError
from pytest_mock import MockerFixture

from app.endpoints.rags import (
    _resolve_rag_id_to_vector_db_id,
    get_rag_endpoint_handler,
    rags_endpoint_handler,
)
from authentication.interface import AuthTuple
from configuration import AppConfig
from tests.unit.utils.auth_helpers import mock_authorization_resolvers


@pytest.mark.asyncio
async def test_rags_endpoint_configuration_not_loaded(
    mocker: MockerFixture,
) -> None:
    """Test that /rags endpoint raises HTTP 500 if configuration is not loaded."""
    mock_authorization_resolvers(mocker)
    mock_config = AppConfig()
    mock_config._configuration = None  # pylint: disable=protected-access
    mocker.patch("app.endpoints.rags.configuration", mock_config)
    request = Request(scope={"type": "http"})

    # Authorization tuple required by URL endpoint handler
    auth: AuthTuple = ("test_user_id", "test_user", True, "test_token")

    with pytest.raises(HTTPException) as e:
        await rags_endpoint_handler(request=request, auth=auth)
    assert e.value.status_code == status.HTTP_500_INTERNAL_SERVER_ERROR


@pytest.mark.asyncio
async def test_rags_endpoint_connection_error(
    mocker: MockerFixture, minimal_config: AppConfig
) -> None:
    """Test that /rags endpoint raises HTTP 503 if Llama Stack connection fails."""
    mocker.patch("app.endpoints.rags.configuration", minimal_config)
    mock_client = mocker.AsyncMock()
    mock_client.vector_stores.list.side_effect = APIConnectionError(request=None)  # type: ignore
    mocker.patch(
        "app.endpoints.rags.AsyncLlamaStackClientHolder"
    ).return_value.get_client.return_value = mock_client

    request = Request(scope={"type": "http"})

    # Authorization tuple required by URL endpoint handler
    auth: AuthTuple = ("test_user_id", "test_user", True, "test_token")

    with pytest.raises(HTTPException) as e:
        await rags_endpoint_handler(request=request, auth=auth)
    assert e.value.status_code == status.HTTP_503_SERVICE_UNAVAILABLE
    detail = e.value.detail
    assert isinstance(detail, dict)
    assert "response" in detail
    assert "Unable to connect to Llama Stack" in detail["response"]  # type: ignore[index]


@pytest.mark.asyncio
async def test_rags_endpoint_success(
    mocker: MockerFixture, minimal_config: AppConfig
) -> None:
    """Test that /rags endpoint returns list of RAG IDs."""
    mocker.patch("app.endpoints.rags.configuration", minimal_config)

    # pylint: disable=R0903
    class RagInfo:
        """RagInfo mock."""

        def __init__(self, rag_id: str) -> None:
            """
            Create a RagInfo instance with the specified identifier.
            
            Parameters:
                rag_id (str): The unique identifier for the RAG.
            """
            self.id = rag_id

    class RagList:
        """List of RAGs mock."""

        def __init__(self) -> None:
            """
            Initialize the RagList with three mock RagInfo entries for tests.
            
            The instance attribute `data` will contain three RagInfo objects with fixed IDs:
            "vs_00000000-cafe-babe-0000-000000000000", "vs_7b52a8cf-0fa3-489c-beab-27e061d102f3",
            and "vs_7b52a8cf-0fa3-489c-cafe-27e061d102f3".
            """
            self.data = [
                RagInfo("vs_00000000-cafe-babe-0000-000000000000"),
                RagInfo("vs_7b52a8cf-0fa3-489c-beab-27e061d102f3"),
                RagInfo("vs_7b52a8cf-0fa3-489c-cafe-27e061d102f3"),
            ]

    mock_client = mocker.AsyncMock()
    mock_client.vector_stores.list.return_value = RagList()
    mocker.patch(
        "app.endpoints.rags.AsyncLlamaStackClientHolder"
    ).return_value.get_client.return_value = mock_client

    request = Request(scope={"type": "http"})

    # Authorization tuple required by URL endpoint handler
    auth: AuthTuple = ("test_user_id", "test_user", True, "test_token")

    response = await rags_endpoint_handler(request=request, auth=auth)
    assert len(response.rags) == 3


@pytest.mark.asyncio
async def test_rag_info_endpoint_configuration_not_loaded(
    mocker: MockerFixture,
) -> None:
    """Test that /rags/{rag_id} endpoint raises HTTP 500 if configuration is not loaded."""
    mock_authorization_resolvers(mocker)
    mock_config = AppConfig()
    mock_config._configuration = None  # pylint: disable=protected-access
    mocker.patch("app.endpoints.rags.configuration", mock_config)
    request = Request(scope={"type": "http"})

    # Authorization tuple required by URL endpoint handler
    auth: AuthTuple = ("test_user_id", "test_user", True, "test_token")

    with pytest.raises(HTTPException) as e:
        await get_rag_endpoint_handler(request=request, auth=auth, rag_id="xyzzy")
    assert e.value.status_code == status.HTTP_500_INTERNAL_SERVER_ERROR


@pytest.mark.asyncio
async def test_rag_info_endpoint_rag_not_found(
    mocker: MockerFixture, minimal_config: AppConfig
) -> None:
    """Test that /rags/{rag_id} endpoint returns HTTP 404 when the requested RAG is not found."""
    mocker.patch("app.endpoints.rags.configuration", minimal_config)
    mock_client = mocker.AsyncMock()
    mock_client.vector_stores.retrieve = mocker.AsyncMock(
        side_effect=BadRequestError(
            message="RAG not found",
            response=mocker.Mock(request=None),
            body=None,
        )
    )  # type: ignore
    mocker.patch(
        "app.endpoints.rags.AsyncLlamaStackClientHolder"
    ).return_value.get_client.return_value = mock_client

    request = Request(scope={"type": "http"})

    # Authorization tuple required by URL endpoint handler
    auth: AuthTuple = ("test_user_id", "test_user", True, "test_token")

    with pytest.raises(HTTPException) as e:
        await get_rag_endpoint_handler(request=request, auth=auth, rag_id="xyzzy")
    assert e.value.status_code == status.HTTP_404_NOT_FOUND
    detail = e.value.detail
    assert isinstance(detail, dict)
    assert "response" in detail
    assert "Rag not found" in detail["response"]  # type: ignore[index]


@pytest.mark.asyncio
async def test_rag_info_endpoint_connection_error(
    mocker: MockerFixture, minimal_config: AppConfig
) -> None:
    """Test that /rags/{rag_id} endpoint raises HTTP 503 if Llama Stack connection fails."""
    mocker.patch("app.endpoints.rags.configuration", minimal_config)
    mock_client = mocker.AsyncMock()
    mock_client.vector_stores.retrieve.side_effect = APIConnectionError(
        request=None  # type: ignore
    )
    mocker.patch(
        "app.endpoints.rags.AsyncLlamaStackClientHolder"
    ).return_value.get_client.return_value = mock_client

    request = Request(scope={"type": "http"})

    # Authorization tuple required by URL endpoint handler
    auth: AuthTuple = ("test_user_id", "test_user", True, "test_token")

    with pytest.raises(HTTPException) as e:
        await get_rag_endpoint_handler(request=request, auth=auth, rag_id="xyzzy")
    assert e.value.status_code == status.HTTP_503_SERVICE_UNAVAILABLE
    detail = e.value.detail
    assert isinstance(detail, dict)
    assert "response" in detail
    assert "Unable to connect to Llama Stack" in detail["response"]  # type: ignore[index]


@pytest.mark.asyncio
async def test_rag_info_endpoint_success(
    mocker: MockerFixture, minimal_config: AppConfig
) -> None:
    """Test that /rags/{rag_id} endpoint returns information about selected RAG."""
    mocker.patch("app.endpoints.rags.configuration", minimal_config)

    # pylint: disable=R0902
    # pylint: disable=R0903
    class RagInfo:
        """RagInfo mock."""

        def __init__(self) -> None:
            """
            Initialize a sample RagInfo-like object populated with fixed test values.
            
            Attributes:
                id (str): "xyzzy"
                name (str): "rag_name"
                created_at (int): 123456
                last_active_at (int): 1234567
                expires_at (int): 12345678
                object (str): "faiss"
                status (str): "completed"
                usage_bytes (int): 100
            """
            self.id = "xyzzy"
            self.name = "rag_name"
            self.created_at = 123456
            self.last_active_at = 1234567
            self.expires_at = 12345678
            self.object = "faiss"
            self.status = "completed"
            self.usage_bytes = 100

    mock_client = mocker.AsyncMock()
    mock_client.vector_stores.retrieve.return_value = RagInfo()
    mocker.patch(
        "app.endpoints.rags.AsyncLlamaStackClientHolder"
    ).return_value.get_client.return_value = mock_client

    request = Request(scope={"type": "http"})

    # Authorization tuple required by URL endpoint handler
    auth: AuthTuple = ("test_user_id", "test_user", True, "test_token")

    response = await get_rag_endpoint_handler(
        request=request, auth=auth, rag_id="xyzzy"
    )
    assert response.id == "xyzzy"
    assert response.name == "rag_name"
    assert response.created_at == 123456
    assert response.last_active_at == 1234567
    assert response.expires_at == 12345678
    assert response.object == "faiss"
    assert response.status == "completed"
    assert response.usage_bytes == 100


def _make_byok_config(tmp_path: Any) -> AppConfig:
    """
    Create an AppConfig pre-populated with two BYOK RAG entries and a SQLite file at tmp_path/test.db for testing.
    
    Parameters:
        tmp_path (Any): Path-like location where the test SQLite file `test.db` will be created.
    
    Returns:
        AppConfig: Configuration populated with Llama Stack connection info and two `byok_rag` entries mapping user-facing `rag_id`s ("ocp-4.18-docs", "company-kb") to their `vector_db_id`s ("vs_abc123", "vs_def456") and pointing both to the created DB path.
    """
    db_file = Path(tmp_path) / "test.db"
    db_file.touch()
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
            "byok_rag": [
                {
                    "rag_id": "ocp-4.18-docs",
                    "rag_type": "inline::faiss",
                    "embedding_model": "all-MiniLM-L6-v2",
                    "embedding_dimension": 384,
                    "vector_db_id": "vs_abc123",
                    "db_path": str(db_file),
                },
                {
                    "rag_id": "company-kb",
                    "rag_type": "inline::faiss",
                    "embedding_model": "all-MiniLM-L6-v2",
                    "embedding_dimension": 384,
                    "vector_db_id": "vs_def456",
                    "db_path": str(db_file),
                },
            ],
        }
    )
    return cfg


@pytest.mark.asyncio
async def test_rags_endpoint_returns_rag_ids_from_config(
    mocker: MockerFixture, tmp_path: Path
) -> None:
    """Test that /rags endpoint maps llama-stack IDs to user-facing rag_ids."""
    byok_config = _make_byok_config(str(tmp_path))
    mocker.patch("app.endpoints.rags.configuration", byok_config)

    # pylint: disable=R0903
    class RagInfo:
        """RagInfo mock."""

        def __init__(self, rag_id: str) -> None:
            """
            Initialize the instance with the given RAG identifier.
            
            Parameters:
                rag_id (str): RAG identifier to assign to the instance's `id` attribute.
            """
            self.id = rag_id

    # pylint: disable=R0903
    class RagList:
        """List of RAGs mock."""

        def __init__(self) -> None:
            """
            Create a predefined list of RagInfo entries used by tests to exercise BYOK mapping and passthrough behavior.
            
            Initializes self.data to three RagInfo objects whose `id` values are:
            - "vs_abc123" (mapped in tests to "ocp-4.18-docs"),
            - "vs_def456" (mapped in tests to "company-kb"),
            - "vs_unmapped" (not present in the BYOK config and expected to pass through unchanged).
            """
            self.data = [
                RagInfo("vs_abc123"),  # mapped to ocp-4.18-docs
                RagInfo("vs_def456"),  # mapped to company-kb
                RagInfo("vs_unmapped"),  # not in config, passed through
            ]

    mock_client = mocker.AsyncMock()
    mock_client.vector_stores.list.return_value = RagList()
    mocker.patch(
        "app.endpoints.rags.AsyncLlamaStackClientHolder"
    ).return_value.get_client.return_value = mock_client

    request = Request(scope={"type": "http"})
    auth: AuthTuple = ("test_user_id", "test_user", True, "test_token")

    response = await rags_endpoint_handler(request=request, auth=auth)
    assert response.rags == ["ocp-4.18-docs", "company-kb", "vs_unmapped"]


@pytest.mark.asyncio
async def test_rag_info_endpoint_accepts_rag_id_from_config(
    mocker: MockerFixture, tmp_path: Path
) -> None:
    """Test that /rags/{rag_id} accepts a user-facing rag_id and resolves it."""
    byok_config = _make_byok_config(str(tmp_path))
    mocker.patch("app.endpoints.rags.configuration", byok_config)

    # pylint: disable=R0902,R0903
    class RagInfo:
        """RagInfo mock."""

        def __init__(self) -> None:
            """
            Create a test RAG/vector-store object populated with fixed sample fields for use in unit tests.
            
            Attributes:
                id (str): Underlying vector-store identifier (e.g., "vs_abc123").
                name (str): Human-readable name of the RAG.
                created_at (int): Creation timestamp.
                last_active_at (int): Last active timestamp.
                expires_at (int): Expiration timestamp.
                object (str): Object type, e.g., "vector_store".
                status (str): Current status, e.g., "completed".
                usage_bytes (int): Storage usage in bytes.
            """
            self.id = "vs_abc123"
            self.name = "OCP 4.18 Docs"
            self.created_at = 100
            self.last_active_at = 200
            self.expires_at = 300
            self.object = "vector_store"
            self.status = "completed"
            self.usage_bytes = 500

    mock_client = mocker.AsyncMock()
    mock_client.vector_stores.retrieve.return_value = RagInfo()
    mocker.patch(
        "app.endpoints.rags.AsyncLlamaStackClientHolder"
    ).return_value.get_client.return_value = mock_client

    request = Request(scope={"type": "http"})
    auth: AuthTuple = ("test_user_id", "test_user", True, "test_token")

    # Pass the user-facing rag_id, not the vector_store_id
    response = await get_rag_endpoint_handler(
        request=request, auth=auth, rag_id="ocp-4.18-docs"
    )

    # The endpoint should resolve ocp-4.18-docs -> vs_abc123 for the lookup
    mock_client.vector_stores.retrieve.assert_called_once_with("vs_abc123")
    # The response should show the user-facing ID
    assert response.id == "ocp-4.18-docs"


def test_resolve_rag_id_to_vector_db_id_with_mapping(tmp_path: Path) -> None:
    """Test that _resolve_rag_id_to_vector_db_id maps rag_id to vector_db_id."""
    byok_config = _make_byok_config(str(tmp_path))
    byok_rags = byok_config.configuration.byok_rag
    assert _resolve_rag_id_to_vector_db_id("ocp-4.18-docs", byok_rags) == "vs_abc123"
    assert _resolve_rag_id_to_vector_db_id("company-kb", byok_rags) == "vs_def456"


def test_resolve_rag_id_to_vector_db_id_passthrough(tmp_path: Path) -> None:
    """Test that unmapped IDs are passed through unchanged."""
    byok_config = _make_byok_config(str(tmp_path))
    byok_rags = byok_config.configuration.byok_rag
    assert _resolve_rag_id_to_vector_db_id("vs_unknown", byok_rags) == "vs_unknown"
