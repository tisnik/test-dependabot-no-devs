# pylint: disable=redefined-outer-name

"""Unit tests for functions defined in authentication/jwk_token.py"""

import time
from collections.abc import Generator
from typing import Any, cast

import pytest
from authlib.jose import JsonWebKey, JsonWebToken
from fastapi import HTTPException, Request
from pydantic import AnyHttpUrl
from pytest_mock import MockerFixture

from authentication.jwk_token import JwkTokenAuthDependency, _jwk_cache
from models.config import JwkConfiguration, JwtConfiguration

TEST_USER_ID = "test-user-123"
TEST_USER_NAME = "testuser"


@pytest.fixture
def token_header(single_key_set: list[dict[str, Any]]) -> dict[str, Any]:
    """
    Create a JWT header with algorithm "RS256" and `kid` taken from the first key in the provided key set.
    
    Parameters:
        single_key_set (list[dict]): List of signing key dictionaries; the first element must contain a `"kid"`.
    
    Returns:
        dict: JWT header with keys `"alg": "RS256"`, `"typ": "JWT"`, and `"kid"` set to `single_key_set[0]["kid"]`.
    """
    return {"alg": "RS256", "typ": "JWT", "kid": single_key_set[0]["kid"]}


@pytest.fixture
def token_payload() -> dict[str, Any]:
    """A sample token payload with the default user_id and username claims.

    Create a sample JWT payload containing the test user claims and timing claims.

    Returns:
        dict: A mapping with keys "user_id", "username", "exp", and "iat";
        "exp" and "iat" are UNIX timestamps (seconds since epoch).
    """
    return {
        "user_id": TEST_USER_ID,
        "username": TEST_USER_NAME,
        "exp": int(time.time()) + 3600,
        "iat": int(time.time()),
    }


def make_key() -> dict[str, Any]:
    """
    Create an RSA test key pair and return the private key, public key, and key identifier.
    
    Returns:
        dict: A dictionary with the following entries:
            - "private_key": the generated private JsonWebKey instance.
            - "public_key": the corresponding public JsonWebKey instance.
            - "kid": the key identifier (thumbprint) as a string.
    """
    key = JsonWebKey.generate_key("RSA", 2048, is_private=True)
    return {
        "private_key": key,
        "public_key": key.get_public_key(),
        "kid": key.thumbprint(),
    }


@pytest.fixture
def single_key_set() -> list[dict[str, Any]]:
    """Default single-key set for signing tokens."""
    return [make_key()]


@pytest.fixture
def another_single_key_set() -> list[dict[str, Any]]:
    """
    Create a single-key JWK set using a newly generated RSA key.
    
    Returns:
        list[dict[str, Any]]: A list containing one dict with keys:
            - `private_key`: the generated private JsonWebKey
            - `public_key`: the corresponding public JsonWebKey
            - `kid`: the key identifier (thumbprint)
    """
    return [make_key()]


@pytest.fixture
def valid_token(
    single_key_set: list[dict[str, Any]],
    token_header: dict[str, Any],
    token_payload: dict[str, Any],
) -> str:
    """
    Create an RS256-signed JWT using the first private key from a single-key JWK set.
    
    Parameters:
        single_key_set (list[dict[str, Any]]): JWK dicts where the first entry must contain a 'private_key' used to sign the token.
        token_header (dict[str, Any]): JWT header values to include in the token.
        token_payload (dict[str, Any]): JWT claims to include in the token.
    
    Returns:
        str: The compact serialized JWT signed with the provided private key.
    """
    jwt_instance = JsonWebToken(algorithms=["RS256"])
    return jwt_instance.encode(
        token_header, token_payload, single_key_set[0]["private_key"]
    ).decode()


@pytest.fixture(autouse=True)
def clear_jwk_cache() -> Generator:
    """
    Clears the global JWK cache before a test runs and again after the test completes.
    
    This autouse fixture ensures the module-level `_jwk_cache` is emptied at setup and teardown to prevent cross-test interference.
    """
    _jwk_cache.clear()
    yield
    _jwk_cache.clear()


def make_signing_server(
    mocker: MockerFixture, key_set: list[dict[str, Any]], algorithms: list[str]
) -> Any:
    """
    Create and patch a mocked aiohttp.ClientSession that serves a JWKS response built from the provided key set.
    
    Parameters:
        mocker (pytest.MockerFixture): Pytest mocker used to patch aiohttp.ClientSession.
        key_set (list[dict[str, Any]]): List of signing key dicts; each item must include a
            `private_key` exposing `as_dict(private=False)` and a `kid` value.
        algorithms (list[str]): List of `alg` values to assign to each corresponding key
            in `key_set`.
    
    Returns:
        Any: The patched `aiohttp.ClientSession` class. The mock is configured so that:
            - Instantiating the session returns an async-capable mock instance.
            - Calling `session.get(...)` returns an async context manager whose entered value
              is a response mock.
            - `response.json()` returns `{"keys": keys}` where each key is the public JWK
              derived from `private_key.as_dict(private=False)` extended with `kid` and `alg`.
            - `response.raise_for_status()` is a no-op.
    """
    mock_session_class = mocker.patch("aiohttp.ClientSession")
    mock_response = mocker.AsyncMock()

    # Create JWK dict from private key as public key
    keys = [
        {
            **key["private_key"].as_dict(private=False),
            "kid": key["kid"],
            "alg": alg,
        }
        for alg, key in zip(algorithms, key_set)
    ]
    mock_response.json.return_value = {
        "keys": keys,
    }
    mock_response.raise_for_status = mocker.MagicMock(return_value=None)

    # Create mock session instance that acts as async context manager
    mock_session_instance = mocker.AsyncMock()
    mock_session_instance.__aenter__ = mocker.AsyncMock(
        return_value=mock_session_instance
    )
    mock_session_instance.__aexit__ = mocker.AsyncMock(return_value=None)

    # Mock the get method to return a context manager
    mock_get_context = mocker.AsyncMock()
    mock_get_context.__aenter__ = mocker.AsyncMock(return_value=mock_response)
    mock_get_context.__aexit__ = mocker.AsyncMock(return_value=None)

    mock_session_instance.get = mocker.MagicMock(return_value=mock_get_context)
    mock_session_class.return_value = mock_session_instance

    return mock_session_class


@pytest.fixture
def mocked_signing_keys_server(
    mocker: MockerFixture, single_key_set: list[dict[str, Any]]
) -> None:
    """
    Register a mocked JWKS HTTP server that serves the provided single RS256 JWK.
    
    Parameters:
        mocker (pytest_mock.MockerFixture): Pytest-mock fixture used to patch aiohttp.ClientSession and related network calls.
        single_key_set (list[dict[str, Any]]): List containing one JWK dict (public key representation) that the mocked JWKS endpoint will return.
    """
    return make_signing_server(mocker, single_key_set, ["RS256"])


@pytest.fixture
def default_jwk_configuration() -> JwkConfiguration:
    """
    Create a JwkConfiguration configured for tests.
    
    Returns:
        JwkConfiguration: configuration pointing at a mocked JWKS URL and using "user_id" as the user ID claim and "username" as the username claim.
    """
    return JwkConfiguration(
        url=AnyHttpUrl("https://this#isgonnabemocked.com/jwks.json"),
        jwt_configuration=JwtConfiguration(
            # Should default to:
            # user_id_claim="user_id", username_claim="username"
        ),  # pyright: ignore[reportCallIssue]
    )


def dummy_request(token: str) -> Request:
    """Generate a dummy request with a given token.

    Create a FastAPI Request with an Authorization Bearer header containing the provided token.

    Parameters:
    ----------
        token (str): Token string to place after the "Bearer " prefix in the Authorization header.

    Returns:
    -------
        request (Request): FastAPI Request object with the Authorization header
        set to "Bearer <token>".
    """
    return Request(
        scope={
            "type": "http",
            "query_string": b"",
            "headers": [(b"authorization", f"Bearer {token}".encode())],
        },
    )


@pytest.fixture
def no_token_request() -> Request:
    """
    Create a FastAPI Request representing an HTTP request without an Authorization header.
    
    Returns:
        request (Request): A Request with an HTTP scope whose headers list contains no Authorization header.
    """
    return Request(
        scope={
            "type": "http",
            "query_string": b"",
            "headers": [],
        },
    )


@pytest.fixture
def not_bearer_token_request() -> Request:
    """
    Create a FastAPI Request with an Authorization header using a non-Bearer scheme.
    
    Returns:
        Request: A request whose `Authorization` header value is "NotBearer anything".
    """
    return Request(
        scope={
            "type": "http",
            "query_string": b"",
            "headers": [(b"authorization", b"NotBearer anything")],
        },
    )


def set_auth_header(request: Request, token: str) -> None:
    """Helper function to set the Authorization header in a request.

    Replace the Request's Authorization header with the given token.

    This mutates request.scope["headers"] to remove any existing Authorization
    header and append a new one using the provided token value. The token
    parameter should be the full header value (for example, "Bearer <token>").

    Parameters:
    ----------
        request (Request): FastAPI/Starlette Request whose headers will be modified.
        token (str): Full Authorization header value to set (e.g., "Bearer <token>").
    """
    new_headers = [
        (k, v) for k, v in request.scope["headers"] if k.lower() != b"authorization"
    ]
    new_headers.append((b"authorization", token.encode()))
    request.scope["headers"] = new_headers


def ensure_test_user_id_and_name(auth_tuple: tuple, expected_token: str) -> None:
    """
    Validate that an authentication tuple matches the expected test user values and token.
    
    Parameters:
        auth_tuple (tuple): A 4-tuple (user_id, username, skip_userid_check, token) to validate.
        expected_token (str): The expected token value that must equal the tuple's fourth element.
    """
    user_id, username, skip_userid_check, token = auth_tuple
    assert user_id == TEST_USER_ID
    assert username == TEST_USER_NAME
    assert skip_userid_check is False
    assert token == expected_token


async def test_valid(
    default_jwk_configuration: JwkConfiguration,
    mocked_signing_keys_server: Any,
    valid_token: str,
) -> None:
    """Test with a valid token."""
    _ = mocked_signing_keys_server

    dependency = JwkTokenAuthDependency(default_jwk_configuration)
    auth_tuple = await dependency(dummy_request(valid_token))

    # Assert the expected values
    ensure_test_user_id_and_name(auth_tuple, valid_token)


@pytest.fixture
def expired_token(
    single_key_set: list[dict[str, Any]],
    token_header: dict[str, Any],
    token_payload: dict[str, Any],
) -> str:
    """
    Create a JWT that is correctly signed but has its expiration time set in the past.
    
    Parameters:
        single_key_set (list[dict]): Key dicts where the first element's `private_key` is used to sign the token.
        token_header (dict): JWT header values to include in the token.
        token_payload (dict): JWT payload values; this function overwrites the `exp` claim to a past timestamp.
    
    Returns:
        str: Signed JWT string with the `exp` claim set to a time in the past.
    """
    jwt_instance = JsonWebToken(algorithms=["RS256"])
    token_payload["exp"] = int(time.time()) - 3600  # Set expiration in the past
    return jwt_instance.encode(
        token_header, token_payload, single_key_set[0]["private_key"]
    ).decode()


async def test_expired(
    default_jwk_configuration: JwkConfiguration,
    mocked_signing_keys_server: Any,
    expired_token: str,
) -> None:
    """Test with an expired token.

    Verifies that JwkTokenAuthDependency rejects an expired JWT.

    Asserts that calling the dependency with an expired token raises an
    HTTPException with status code 401 and a message containing "Token has
    expired".
    """
    _ = mocked_signing_keys_server

    dependency = JwkTokenAuthDependency(default_jwk_configuration)

    # Assert that an HTTPException is raised when the token is expired
    with pytest.raises(HTTPException) as exc_info:
        await dependency(dummy_request(expired_token))

    assert "Token has expired" in str(exc_info.value)
    assert exc_info.value.status_code == 401


@pytest.fixture
def invalid_token(
    another_single_key_set: list[dict[str, Any]],
    token_header: dict[str, Any],
    token_payload: dict[str, Any],
) -> str:
    """
    Create a JWT signed with a different private key than the verifier's keys to produce an invalid-signature token for tests.
    
    Parameters:
        another_single_key_set: Key set whose first entry's private key will be used to sign the token; this key should not match the verifier's keys.
        token_header: JWT header values to encode into the token.
        token_payload: JWT claims to encode into the token.
    
    Returns:
        The compact serialized JWT string signed with the provided private key.
    """
    jwt_instance = JsonWebToken(algorithms=["RS256"])
    return jwt_instance.encode(
        token_header, token_payload, another_single_key_set[0]["private_key"]
    ).decode()


async def test_invalid(
    default_jwk_configuration: JwkConfiguration,
    mocked_signing_keys_server: Any,
    invalid_token: str,
) -> None:
    """Test with an invalid token."""
    _ = mocked_signing_keys_server

    dependency = JwkTokenAuthDependency(default_jwk_configuration)

    with pytest.raises(HTTPException) as exc_info:
        await dependency(dummy_request(invalid_token))

    assert "Invalid token" in str(exc_info.value)
    assert exc_info.value.status_code == 401


async def test_no_auth_header(
    default_jwk_configuration: JwkConfiguration,
    mocked_signing_keys_server: Any,
    no_token_request: Request,
) -> None:
    """Test with no Authorization header returns 401 Unauthorized."""
    _ = mocked_signing_keys_server

    dependency = JwkTokenAuthDependency(default_jwk_configuration)

    with pytest.raises(HTTPException) as exc_info:
        await dependency(no_token_request)

    assert exc_info.value.status_code == 401
    detail = cast(dict, exc_info.value.detail)
    assert detail["cause"] == "No Authorization header found"
    assert detail["response"] == "Missing or invalid credentials provided by client"


async def test_no_bearer(
    default_jwk_configuration: JwkConfiguration,
    mocked_signing_keys_server: Any,
    not_bearer_token_request: Request,
) -> None:
    """Test with Authorization header that does not start with Bearer."""
    _ = mocked_signing_keys_server

    dependency = JwkTokenAuthDependency(default_jwk_configuration)

    with pytest.raises(HTTPException) as exc_info:
        await dependency(not_bearer_token_request)

    assert exc_info.value.status_code == 401
    detail = cast(dict[str, str], exc_info.value.detail)
    assert detail["response"] == ("Missing or invalid credentials provided by client")
    assert detail["cause"] == "No token found in Authorization header"


@pytest.fixture
def no_user_id_token(
    single_key_set: list[dict[str, Any]],
    token_payload: dict[str, Any],
    token_header: dict[str, Any],
) -> str:
    """
    Create a signed JWT that omits the `user_id` claim.
    
    Modifies the provided `token_payload` in-place to remove the `user_id` key and returns the JWT encoded and signed with the first private key from `single_key_set`.
    
    Parameters:
        token_payload (dict): Payload to encode; will be mutated to remove `user_id`.
    
    Returns:
        jwt (str): Encoded JWT string that does not contain the `user_id` claim.
    """
    jwt_instance = JsonWebToken(algorithms=["RS256"])
    # Modify the token payload to include different claims
    del token_payload["user_id"]

    return jwt_instance.encode(
        token_header, token_payload, single_key_set[0]["private_key"]
    ).decode()


async def test_no_user_id(
    default_jwk_configuration: JwkConfiguration,
    mocked_signing_keys_server: Any,
    no_user_id_token: str,
) -> None:
    """Test with a token that has no user_id claim."""
    _ = mocked_signing_keys_server

    dependency = JwkTokenAuthDependency(default_jwk_configuration)

    with pytest.raises(HTTPException) as exc_info:
        await dependency(dummy_request(no_user_id_token))

    assert exc_info.value.status_code == 401
    assert "user_id" in str(exc_info.value.detail) and "missing" in str(
        exc_info.value.detail
    )


@pytest.fixture
def no_username_token(
    single_key_set: list[dict[str, Any]],
    token_payload: dict[str, Any],
    token_header: dict[str, Any],
) -> str:
    """
    Create a JWT signed with the provided private key that omits the `username` claim.
    
    Returns:
        compact_jwt (str): A signed compact JWT string whose payload does not include the `username` claim.
    """
    jwt_instance = JsonWebToken(algorithms=["RS256"])
    # Modify the token payload to include different claims
    del token_payload["username"]

    return jwt_instance.encode(
        token_header, token_payload, single_key_set[0]["private_key"]
    ).decode()


async def test_no_username(
    default_jwk_configuration: JwkConfiguration,
    mocked_signing_keys_server: Any,
    no_username_token: str,
) -> None:
    """Test with a token that has no username claim."""
    _ = mocked_signing_keys_server

    dependency = JwkTokenAuthDependency(default_jwk_configuration)

    with pytest.raises(HTTPException) as exc_info:
        await dependency(dummy_request(no_username_token))

    assert exc_info.value.status_code == 401
    assert "username" in str(exc_info.value.detail) and "missing" in str(
        exc_info.value.detail
    )


@pytest.fixture
def custom_claims_token(
    single_key_set: list[dict[str, Any]],
    token_payload: dict[str, Any],
    token_header: dict[str, Any],
) -> str:
    """
    Create an RS256-signed JWT that uses custom claim names for the user id and username.
    
    Replaces `user_id` and `username` in the provided payload with `id_of_the_user` and `name_of_the_user`, then signs the token using the first key in `single_key_set`.
    
    Parameters:
        single_key_set (list[dict[str, Any]]): Signing key dicts; the first entry's `private_key` is used to sign the token.
        token_payload (dict[str, Any]): Base JWT claims; `user_id` and `username` will be replaced by the custom claim names.
        token_header (dict[str, Any]): JWT header to include in the encoded token.
    
    Returns:
        str: The encoded JWT as a compact serialized string.
    """
    jwt_instance = JsonWebToken(algorithms=["RS256"])

    del token_payload["user_id"]
    del token_payload["username"]

    # Add custom claims
    token_payload["id_of_the_user"] = TEST_USER_ID
    token_payload["name_of_the_user"] = TEST_USER_NAME

    return jwt_instance.encode(
        token_header, token_payload, single_key_set[0]["private_key"]
    ).decode()


@pytest.fixture
def custom_claims_configuration(
    default_jwk_configuration: JwkConfiguration,
) -> JwkConfiguration:
    """
    Create a configuration that maps custom JWT claim names for the user id and username.
    
    Parameters:
        default_jwk_configuration (JwkConfiguration): Base configuration to copy and modify.
    
    Returns:
        JwkConfiguration: A copy of the input configuration with `jwt_configuration.user_id_claim`
        set to "id_of_the_user" and `jwt_configuration.username_claim` set to "name_of_the_user".
    """
    # Create a copy of the default configuration
    custom_config = default_jwk_configuration.model_copy()

    # Set custom claims
    custom_config.jwt_configuration.user_id_claim = "id_of_the_user"
    custom_config.jwt_configuration.username_claim = "name_of_the_user"

    return custom_config


async def test_custom_claims(
    custom_claims_configuration: JwkConfiguration,
    mocked_signing_keys_server: Any,
    custom_claims_token: str,
) -> None:
    """
    Verifies that the dependency extracts user id and username from a token using custom claim names.
    
    Asserts the returned auth tuple matches the expected user id, username, skip flag, and original token.
    """
    _ = mocked_signing_keys_server

    dependency = JwkTokenAuthDependency(custom_claims_configuration)

    auth_tuple = await dependency(dummy_request(custom_claims_token))

    # Assert the expected values
    ensure_test_user_id_and_name(auth_tuple, custom_claims_token)


@pytest.fixture
def token_header_256_1(multi_key_set: list[dict[str, Any]]) -> dict[str, Any]:
    """
    Create a JWT header that specifies algorithm RS256 and sets the `kid` to the first key's `kid` from multi_key_set.
    
    Parameters:
        multi_key_set (list[dict[str, Any]]): List of JWK dictionaries; the header's `kid` is taken from multi_key_set[0]["kid"].
    
    Returns:
        dict[str, Any]: JWT header containing "alg", "typ", and "kid".
    """
    return {"alg": "RS256", "typ": "JWT", "kid": multi_key_set[0]["kid"]}


@pytest.fixture
def token_header_256_2(multi_key_set: list[dict[str, Any]]) -> dict[str, Any]:
    """A sample token header for RS256 using multi_key_set.

    Create a JWT header for RS256 that references the second key in a multi-key set.

    Parameters:
    ----------
        multi_key_set (list[dict[str, Any]]): List of JWK-like dicts where each
        dict contains a `"kid"` entry; the second entry (index 1) is used.

    Returns:
    -------
        dict[str, Any]: JWT header with keys `"alg": "RS256"`, `"typ": "JWT"`,
        and `"kid"` taken from `multi_key_set[1]["kid"]`.
    """
    return {"alg": "RS256", "typ": "JWT", "kid": multi_key_set[1]["kid"]}


@pytest.fixture
def token_header_384(multi_key_set: list[dict[str, Any]]) -> dict[str, Any]:
    """
    Builds a JWT header for RS384 using the third key's `kid`.
    
    Parameters:
        multi_key_set (list[dict[str, Any]]): A list of JWK-like dictionaries; must contain at least three entries. The `kid` from index 2 is used.
    
    Returns:
        dict[str, Any]: JWT header with keys `"alg": "RS384"`, `"typ": "JWT"`, and `"kid"` set to the third key's `kid`.
    """
    return {"alg": "RS384", "typ": "JWT", "kid": multi_key_set[2]["kid"]}


@pytest.fixture
def token_header_256_no_kid() -> dict[str, Any]:
    """
    Constructs a JWT header indicating RS256 without a key ID.
    
    Returns:
        header (dict[str, Any]): JWT header with "alg" set to "RS256", "typ" set to "JWT", and no "kid" field.
    """
    return {"alg": "RS256", "typ": "JWT"}


@pytest.fixture
def token_header_384_no_kid() -> dict[str, Any]:
    """
    Create a JWT header that specifies the RS384 algorithm and omits the `kid` field.
    
    Returns:
        header (dict): JWT header containing `"alg": "RS384"` and `"typ": "JWT"`.
    """
    return {"alg": "RS384", "typ": "JWT"}


@pytest.fixture
def multi_key_set() -> list[dict[str, Any]]:
    """
    Create three distinct RSA signing key dictionaries for multi-key tests.
    
    Each dictionary contains 'private_key', 'public_key', and 'kid' for use by the test suite.
    
    Returns:
        list[dict[str, Any]]: A list of three signing key dictionaries.
    """
    return [make_key(), make_key(), make_key()]


@pytest.fixture
def valid_tokens(
    multi_key_set: list[dict[str, Any]],
    token_header_256_1: dict[str, Any],
    token_header_256_2: dict[str, Any],
    token_payload: dict[str, Any],
    token_header_384: dict[str, Any],
) -> tuple[str, str, str]:
    """
    Create three JWTs signed by the three keys in `multi_key_set` using the provided headers and payload.
    
    Each returned token is signed with the corresponding key and algorithm:
    - token1: `multi_key_set[0]` with `token_header_256_1` (RS256)
    - token2: `multi_key_set[1]` with `token_header_256_2` (RS256)
    - token3: `multi_key_set[2]` with `token_header_384` (RS384)
    
    Returns:
        tuple[str, str, str]: A tuple of JWT compact-serialization strings `(token1, token2, token3)`.
    """
    key_for_256_1 = multi_key_set[0]
    key_for_256_2 = multi_key_set[1]
    key_for_384 = multi_key_set[2]

    jwt_instance1 = JsonWebToken(algorithms=["RS256"])
    token1 = jwt_instance1.encode(
        token_header_256_1, token_payload, key_for_256_1["private_key"]
    ).decode()

    jwt_instance2 = JsonWebToken(algorithms=["RS256"])
    token2 = jwt_instance2.encode(
        token_header_256_2, token_payload, key_for_256_2["private_key"]
    ).decode()

    jwt_instance3 = JsonWebToken(algorithms=["RS384"])
    token3 = jwt_instance3.encode(
        token_header_384, token_payload, key_for_384["private_key"]
    ).decode()

    return token1, token2, token3


@pytest.fixture
def valid_tokens_no_kid(
    multi_key_set: list[dict[str, Any]],
    token_header_256_no_kid: dict[str, Any],
    token_payload: dict[str, Any],
    token_header_384_no_kid: dict[str, Any],
) -> tuple[str, str, str]:
    """
    Generate three JWTs signed by the provided multi-key set using headers that omit the `kid`.
    
    Returns:
        tuple[str, str, str]: Three JWT strings in order:
            (1) RS256 signed with the first key,
            (2) RS256 signed with the second key,
            (3) RS384 signed with the third key.
    """
    key_for_256_1 = multi_key_set[0]
    key_for_256_2 = multi_key_set[1]
    key_for_384 = multi_key_set[2]

    jwt_instance1 = JsonWebToken(algorithms=["RS256"])
    token1 = jwt_instance1.encode(
        token_header_256_no_kid, token_payload, key_for_256_1["private_key"]
    ).decode()

    jwt_instance2 = JsonWebToken(algorithms=["RS256"])
    token2 = jwt_instance2.encode(
        token_header_256_no_kid, token_payload, key_for_256_2["private_key"]
    ).decode()

    jwt_instance3 = JsonWebToken(algorithms=["RS384"])
    token3 = jwt_instance3.encode(
        token_header_384_no_kid, token_payload, key_for_384["private_key"]
    ).decode()

    return token1, token2, token3


@pytest.fixture
def multi_key_signing_server(
    mocker: MockerFixture, multi_key_set: list[dict[str, Any]]
) -> Any:
    """Multi-key signing server.

    Builds a mocked JWKS HTTP server that serves a multi-key key set.

    Creates and returns a mock aiohttp signing-keys server wired to the
    provided `multi_key_set` and configured to advertise algorithms ["RS256",
    "RS256", "RS384"].

    Parameters:
    ----------
        mocker: pytest-mock MockerFixture used to patch aiohttp client behavior.
        multi_key_set (list[dict[str, Any]]): List of JWK dictionaries to be
        served by the mock JWKS endpoint.

    Returns:
    -------
        A mock object that simulates an aiohttp client/session which, when
        queried, yields a response containing the configured JWKs.
    """
    return make_signing_server(mocker, multi_key_set, ["RS256", "RS256", "RS384"])


async def test_multi_key_valid(
    default_jwk_configuration: JwkConfiguration,
    multi_key_signing_server: Any,
    valid_tokens: tuple[str, str, str],
) -> None:
    """Test with valid tokens from a multi-key set."""
    _ = multi_key_signing_server

    token1, token2, token3 = valid_tokens

    dependency = JwkTokenAuthDependency(default_jwk_configuration)
    auth_tuple = await dependency(dummy_request(token1))
    ensure_test_user_id_and_name(auth_tuple, token1)

    auth_tuple = await dependency(dummy_request(token2))
    ensure_test_user_id_and_name(auth_tuple, token2)

    auth_tuple = await dependency(dummy_request(token3))
    ensure_test_user_id_and_name(auth_tuple, token3)


async def test_multi_key_no_kid(
    default_jwk_configuration: JwkConfiguration,
    multi_key_signing_server: Any,
    valid_tokens_no_kid: tuple[str, str, str],
) -> None:
    """Test with valid tokens from a multi-key set without a kid."""
    _ = multi_key_signing_server

    token1, token2, token3 = valid_tokens_no_kid

    dependency = JwkTokenAuthDependency(default_jwk_configuration)

    auth_tuple = await dependency(dummy_request(token1))
    ensure_test_user_id_and_name(auth_tuple, token1)

    # Token 2 should fail, as it has no kid and multiple keys for its algorithm are present
    # and the one that signed it is not the first key

    with pytest.raises(HTTPException) as exc_info:
        await dependency(dummy_request(token2))
    assert exc_info.value.status_code == 401

    # Token 3 will succeed, as it has a different algorithm (RS384) and there's only one key
    # for that algorithm in the multi-key set

    auth_tuple = await dependency(dummy_request(token3))
    ensure_test_user_id_and_name(auth_tuple, token3)
