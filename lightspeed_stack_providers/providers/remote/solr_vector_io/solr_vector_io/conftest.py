"""Pytest configuration file for Solr vector IO tests."""

from collections.abc import Generator

import pytest
import requests


@pytest.fixture(scope="session", autouse=True)
def check_solr_running() -> Generator:
    """
    Verify Solr is reachable before running tests.
    
    Performs a one-time check against the configured Solr collection select endpoint (uses tests.SOLR_URL and tests.COLLECTION_NAME). If Solr responds with HTTP 200 the fixture yields and allows tests to proceed; on non-200 responses, connection errors, timeouts (5 second request timeout), or other exceptions the fixture aborts the entire pytest run via pytest.exit with a non-zero return code.
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
