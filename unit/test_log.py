"""Unit tests for functions defined in src/log.py."""

from log import get_logger


def test_get_logger() -> None:
    """Check the function to retrieve logger."""
    logger_name = "foo"
    logger = get_logger(logger_name)
    assert logger is not None
    assert logger.name == logger_name

    # at least one handler need to be set
    assert len(logger.handlers) >= 1
