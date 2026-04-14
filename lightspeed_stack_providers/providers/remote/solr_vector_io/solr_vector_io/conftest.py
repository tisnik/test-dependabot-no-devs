"""Pytest configuration file for Solr vector IO tests."""

from collections.abc import Generator

import pytest
import requests


@pytest.fixture(scope="session", autouse=True)
def check_solr_running() -> Generator:
    """
    Verify that the configured Solr collection endpoint responds before any tests run.
    
    Performs a connectivity check to the Solr collection select endpoint and aborts the entire test session with pytest.exit if Solr is unreachable or returns a non-200 status. Exits occur for connection refusal, request timeout, or any other unexpected error. Yields once when the check succeeds to allow the test session to proceed.
    """
    # Import SOLR_URL from test file to avoid duplication
    from tests import COLLECTION_NAME, SOLR_URL

    solr_test_url = SOLR_URL + "/" + COLLECTION_NAME + "/select"
    print(solr_test_url)

    try:
        response = requests.get(solr_test_url, timeout=5)
        if response.status_code == 200:
            print(f"✓ Solr is running at {solr_test_url}")
        else:
            print(f"\n✗ FAILED: Solr returned status code {response.status_code}")
            print(f"  Expected: 200, Got: {response.status_code}")
            pytest.exit(
                f"Solr is not responding correctly at {solr_test_url} "
                f"(status code: {response.status_code}). "
                f"Please start Solr before running tests.",
                returncode=1,
            )
    except requests.exceptions.ConnectionError:
        print(f"\n✗ FAILED: Could not connect to Solr at {SOLR_URL}")
        print("  Error: Connection refused")
        print("\nPossible solutions:")
        print("  1. Start Solr if it's not running")
        print("  2. Verify Solr is running on the correct host/port")
        print("  3. Check firewall settings")
        pytest.exit(
            f"Cannot connect to Solr at {SOLR_URL}. "
            f"Please start Solr before running tests.",
            returncode=1,
        )
    except requests.exceptions.Timeout:
        print(f"\n✗ FAILED: Timeout connecting to Solr at {SOLR_URL}")
        print("  Request timed out after 5 seconds")
        pytest.exit(
            f"Timeout connecting to Solr at {SOLR_URL}. "
            f"Solr may be starting up or experiencing issues.",
            returncode=1,
        )
    except Exception as e:
        print(f"\n✗ FAILED: Unexpected error checking Solr at {SOLR_URL}")
        print(f"  Error: {type(e).__name__}: {e}")
        pytest.exit(f"Unexpected error checking Solr: {e}", returncode=1)

    yield
