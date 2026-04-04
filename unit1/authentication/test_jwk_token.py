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
    Create a JWT header using RS256 with the first key's `kid`.
    
    Parameters:
        single_key_set (list[dict[str, Any]]): List of key dictionaries; the first element must include a `"kid"` which will be used for the header.
    
    Returns:
        dict[str, Any]: Header dictionary with keys `"alg": "RS256"`, `"typ": "JWT"`, and `"kid"` set to the first key's `kid`.
    """
    return {"alg": "RS256", "typ": "JWT", "kid": single_key_set[0]["kid"]}


@pytest.fixture
def token_payload() -> dict[str, Any]:
    """
    Create a sample JWT payload containing the test user's identity and standard timing claims.
    
    Returns:
        dict: Mapping with keys:
            - "user_id": test user identifier
            - "username": test user name
            - "exp": expiration time as a UNIX timestamp (seconds since epoch)
            - "iat": issued-at time as a UNIX timestamp (seconds since epoch)
    """
    return {
        "user_id": TEST_USER_ID,
        "username": TEST_USER_NAME,
        "exp": int(time.time()) + 3600,
        "iat": int(time.time()),
    }


def make_key() -> dict[str, Any]:
    """
    Generate an RSA JsonWebKey key pair for tests.
    
    Returns:
        dict[str, Any]: A dictionary containing:
            - "private_key": the generated private `JsonWebKey` instance.
            - "public_key": the corresponding public `JsonWebKey` instance.
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
    """
    Create a one-key JWKS-like set used to sign test JWTs.
    
    The returned list contains a single dict describing the generated key:
    - `private_key`: the private JsonWebKey used for signing.
    - `public_key`: the corresponding public JsonWebKey.
    - `kid`: the key identifier string.
    
    Returns:
        list[dict[str, Any]]: A list with one key descriptor dict as described above.
    """
    return [make_key()]


@pytest.fixture
def another_single_key_set() -> list[dict[str, Any]]:
    """
    Produce a single-entry JWK set containing a newly generated RSA key pair.
    
    Returns:
        list[dict[str, Any]]: A list with one dict containing:
            - private_key: the generated private `JsonWebKey`
            - public_key: the corresponding public `JsonWebKey`
            - kid: the key identifier (thumbprint)
    """
    return [make_key()]


@pytest.fixture
def valid_token(
    single_key_set: list[dict[str, Any]],
    token_header: dict[str, Any],
    token_payload: dict[str, Any],
) -> str:
    """
    Create an RS256 JWT signed with the first private key from a single-key JWK set.
    
    Parameters:
        single_key_set (list[dict[str, Any]]): List of key dicts where the first entry must contain a 'private_key' (a JsonWebKey) used to sign the token.
        token_header (dict[str, Any]): JWT header values to include in the token.
        token_payload (dict[str, Any]): JWT claims to include in the token.
    
    Returns:
        str: The compact serialized JWT.
    """
    jwt_instance = JsonWebToken(algorithms=["RS256"])
    return jwt_instance.encode(
        token_header, token_payload, single_key_set[0]["private_key"]
    ).decode()


@pytest.fixture(autouse=True)
def clear_jwk_cache() -> Generator:
    """Clear the global JWK cache before each test.

    This autouse fixture ensures the module-level `_jwk_cache` is emptied at
    setup and teardown to prevent cross-test interference.
    """
    _jwk_cache.clear()
    yield
    _jwk_cache.clear()


def make_signing_server(
    mocker: MockerFixture, key_set: list[dict[str, Any]], algorithms: list[str]
) -> Any:
    """
    Create and patch a mocked aiohttp.ClientSession that serves a JWKS payload derived from the provided key set.
    
    Parameters:
        mocker (pytest.MockerFixture): Pytest mocker used to patch aiohttp.ClientSession.
        key_set (list[dict[str, Any]]): List of signing key dicts; each item must include a
            `private_key` with an `as_dict(private=False)` method and a `kid` value.
        algorithms (list[str]): List of JWK `alg` values to assign to each corresponding key in `key_set`.
    
    Returns:
        Any: The mocked ClientSession class patched at `aiohttp.ClientSession`. The mock is configured so that:
            - Instantiating the class returns an async context manager.
            - Calling `await session.get(...)` returns an async context manager whose entered value is a response object.
            - `response.json()` returns `{"keys": [<public JWK dicts with 'kid' and 'alg'>]}`.
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
    Create and register a mocked JWKS HTTP server that serves a single RS256 public key.
    
    Parameters:
        single_key_set (list[dict[str, Any]]): A list containing one JWK dict (public key)
            to be returned by the mocked JWKS endpoint.
    """
    return make_signing_server(mocker, single_key_set, ["RS256"])


@pytest.fixture
def default_jwk_configuration() -> JwkConfiguration:
    """
    Create a JwkConfiguration for tests with a mocked JWKS URL and default JWT claim mappings.
    
    The returned configuration points `url` to a deterministic mocked JWKS endpoint and sets `jwt_configuration` so the user identifier and username are read from the `user_id` and `username` claims respectively.
    
    Returns:
        JwkConfiguration: Configuration using a mocked JWKS URL and default claim names `user_id` and `username`.
    """
    return JwkConfiguration(
        url=AnyHttpUrl("https://this#isgonnabemocked.com/jwks.json"),
        jwt_configuration=JwtConfiguration(
            # Should default to:
            # user_id_claim="user_id", username_claim="username"
        ),  # pyright: ignore[reportCallIssue]
    )


def dummy_request(token: str) -> Request:
    """
    Create a FastAPI Request whose Authorization header is set to a Bearer token.
    
    Parameters:
        token (str): Token string to append after the "Bearer " prefix in the Authorization header.
    
    Returns:
        Request: FastAPI Request object with the `authorization` header set to "Bearer <token>".
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
    Create a FastAPI Request whose scope contains no Authorization header.
    
    Returns:
        request (Request): A Request with scope {"type": "http", "query_string": b\"\", "headers": []}.
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
    Constructs a FastAPI Request with an Authorization header using a non-Bearer scheme.
    
    Returns:
        Request: A Request object whose Authorization header is set to "NotBearer anything".
    """
    return Request(
        scope={
            "type": "http",
            "query_string": b"",
            "headers": [(b"authorization", b"NotBearer anything")],
        },
    )


def set_auth_header(request: Request, token: str) -> None:
    """
    Remove any existing Authorization header from the request and set it to the provided header value.
    
    Parameters:
        request (Request): FastAPI/Starlette request whose headers will be modified.
        token (str): Full Authorization header value to set (e.g., "Bearer <token>").
    """
    new_headers = [
        (k, v) for k, v in request.scope["headers"] if k.lower() != b"authorization"
    ]
    new_headers.append((b"authorization", token.encode()))
    request.scope["headers"] = new_headers


def ensure_test_user_id_and_name(auth_tuple: tuple, expected_token: str) -> None:
    """Utility to ensure that the values in the auth tuple match the test values.

    Assert that an authentication tuple contains the expected test user values.

    Parameters:
    ----------
        auth_tuple (tuple): A 4-tuple in the form (user_id, username, skip_userid_check, token).
        expected_token (str): The token value expected to be present as the fourth element.

    Raises:
    ------
        AssertionError: If any element of auth_tuple does not match the expected test values
                        (user id equals TEST_USER_ID, username equals TEST_USER_NAME,
                        skip_userid_check is False, and token equals expected_token).
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
    """
    Verify that JwkTokenAuthDependency accepts a correctly signed JWT and returns the expected authentication tuple.
    
    Uses a mocked JWKS endpoint and the provided valid_token fixture; asserts the returned (user_id, username, skip_userid_check, token) tuple matches the token's claims.
    """
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
    Create a correctly signed JWT whose `exp` claim is set to a past timestamp.
    
    Parameters:
        single_key_set (list[dict]): List of key dictionaries; the first element's
            `private_key` is used to sign the token.
        token_header (dict): JWT header values to include in the token.
        token_payload (dict): JWT payload values; the `exp` claim will be overwritten
            to a timestamp in the past.
    
    Returns:
        str: Signed JWT string with an `exp` claim set to a past timestamp.
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
    Create a JWT signed with a private key that does not match the verifier's keys for use in invalid-signature tests.
    
    Parameters:
        another_single_key_set: A key set whose first entry's `private_key` is used to sign the token; it must differ from the verifier's key set.
    
    Returns:
        The serialized JWT as a compact string.
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
    Create a JWT signed with RS256 that omits the `user_id` claim.
    
    Mutates `token_payload` by removing the `user_id` key, then returns the compact JWT string signed using the first private key in `single_key_set` with the provided header.
    
    Parameters:
    	token_payload (dict): Payload dict to encode; this dict will be modified in-place (the `user_id` key is removed).
    
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
    """Token without a username claim.

    Create a JWT signed with the provided private key that omits the `username` claim.

    Returns:
        A compact JWT string (signed) that does not contain the `username` claim.
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
    Create a JWT signed with the first key that uses custom claim names for user id and username.
    
    Removes `user_id` and `username` from the provided payload, injects `id_of_the_user` and `name_of_the_user` with test values, and encodes the JWT using the first entry's private key and the given header.
    
    Parameters:
        single_key_set (list[dict[str, Any]]): Signing key dicts; the first entry's `private_key` is used.
        token_payload (dict[str, Any]): Base claims payload to modify before signing.
        token_header (dict[str, Any]): JWT header to include in the token.
    
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
    Create a JwkConfiguration that maps JWT user ID and username to custom claim names.
    
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
    Verifies that the dependency extracts user id and username from tokens using custom claim names configured in JwkConfiguration.
    
    Asserts the returned authentication tuple contains the expected user id, username, skip_userid_check == False, and the original token.
    """
    _ = mocked_signing_keys_server

    dependency = JwkTokenAuthDependency(custom_claims_configuration)

    auth_tuple = await dependency(dummy_request(custom_claims_token))

    # Assert the expected values
    ensure_test_user_id_and_name(auth_tuple, custom_claims_token)


@pytest.fixture
def token_header_256_1(multi_key_set: list[dict[str, Any]]) -> dict[str, Any]:
    """A sample token header for RS256 using multi_key_set.

    Parameters:
        multi_key_set (list[dict[str, Any]]): List of JWK dictionaries; the
        header's `kid` is taken from multi_key_set[0]["kid"].

    Returns:
        dict[str, Any]: JWT header containing "alg", "typ", and "kid".
    """
    return {"alg": "RS256", "typ": "JWT", "kid": multi_key_set[0]["kid"]}


@pytest.fixture
def token_header_256_2(multi_key_set: list[dict[str, Any]]) -> dict[str, Any]:
    """
    Create a JWT header for RS256 that references the second key in a multi-key set.
    
    Parameters:
        multi_key_set (list[dict]): List of JWK-like dictionaries; the function uses the `"kid"` value from the second entry (index 1).
    
    Returns:
        dict: JWT header with `"alg": "RS256"`, `"typ": "JWT"`, and `"kid"` set to `multi_key_set[1]["kid"]`.
    """
    return {"alg": "RS256", "typ": "JWT", "kid": multi_key_set[1]["kid"]}


@pytest.fixture
def token_header_384(multi_key_set: list[dict[str, Any]]) -> dict[str, Any]:
    """
    Builds a JWT header for RS384 that uses the third key's `kid` from a multi-key set.
    
    Parameters:
        multi_key_set (list[dict[str, Any]]): JWK-like dicts; must contain at least three entries. The `kid` from index 2 is used.
    
    Returns:
        dict[str, Any]: JWT header with keys `"alg": "RS384"`, `"typ": "JWT"`, and `"kid"` set to the third key's `kid`.
    """
    return {"alg": "RS384", "typ": "JWT", "kid": multi_key_set[2]["kid"]}


@pytest.fixture
def token_header_256_no_kid() -> dict[str, Any]:
    """
    Builds a JWT header specifying RS256 and omitting a key ID.
    
    Returns:
        header (dict[str, Any]): JWT header with "alg" set to "RS256", "typ" set to "JWT", and no "kid" field.
    """
    return {"alg": "RS256", "typ": "JWT"}


@pytest.fixture
def token_header_384_no_kid() -> dict[str, Any]:
    """
    Create a JWT header for the RS384 algorithm that omits the `kid` field.
    
    Returns:
        header (dict): A JWT header containing `"alg": "RS384"` and `"typ": "JWT"`, without a `kid` entry.
    """
    return {"alg": "RS384", "typ": "JWT"}


@pytest.fixture
def multi_key_set() -> list[dict[str, Any]]:
    """
    Return three distinct RSA signing key dictionaries for multi-key tests.
    
    Each dictionary contains the generated key pair and identifier used by the tests:
    - `private_key`: private JsonWebKey used to sign tokens
    - `public_key`: public JsonWebKey for verification
    - `kid`: key identifier string
    
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
    Create three JWTs signed by the three keys in the provided multi-key set.
    
    Each returned token is encoded with the given payload and the corresponding header:
    the first two use RS256 with headers `token_header_256_1` and `token_header_256_2`,
    the third uses RS384 with `token_header_384`.
    
    Returns:
        tuple[str, str, str]: (token1, token2, token3) — JWT strings signed with
        multi_key_set[0] (RS256, `token_header_256_1`), multi_key_set[1] (RS256,
        `token_header_256_2`), and multi_key_set[2] (RS384, `token_header_384`).
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
    Generate three JWTs signed by the provided multi-key set with headers that omit the `kid`.
    
    Returns:
        tuple[str, str, str]: JWT strings in order:
            - first: RS256 signed with the first key,
            - second: RS256 signed with the second key,
            - third: RS384 signed with the third key.
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
    """
    Builds a mocked JWKS HTTP server that serves the provided multi-key JWK set.
    
    Creates and returns a patched aiohttp client session mock configured to return a JWKS response whose "keys" value is derived from the supplied `multi_key_set` and which advertises algorithms ["RS256", "RS256", "RS384"].
    
    Parameters:
        mocker (pytest_mock.MockerFixture): Fixture used to patch aiohttp.ClientSession.
        multi_key_set (list[dict[str, Any]]): List of JWK dictionaries to be served by the mock JWKS endpoint.
    
    Returns:
        Any: A mock of the aiohttp ClientSession class whose `get` method yields an async context manager whose response exposes `json()` returning `{"keys": <provided keys>}` and whose `raise_for_status()` is a no-op.
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
    """
    Validate JwkTokenAuthDependency handling of tokens that omit the `kid` header when multiple JWKS keys are available.
    
    Checks that:
    - a token whose algorithm uniquely identifies a single key is accepted,
    - a token that omits `kid` but has multiple candidate keys for its algorithm is rejected with a 401,
    - a token for an algorithm with only one matching key is accepted.
    """
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
