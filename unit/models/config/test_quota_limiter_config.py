"""Unit tests for QuotaLimiterConfig model."""

import pytest

from models.config import QuotaLimiterConfiguration


def test_quota_limiter_configuration() -> None:
    """Test the default configuration."""
    cfg = QuotaLimiterConfiguration(
        type="cluster_limiter",
        name="cluster_monthly_limits",
        initial_quota=0,
        quota_increase=10,
        period="3 seconds",
    )
    assert cfg is not None
    assert cfg.type == "cluster_limiter"
    assert cfg.name == "cluster_monthly_limits"
    assert cfg.initial_quota == 0
    assert cfg.quota_increase == 10
    assert cfg.period == "3 seconds"


def test_quota_limiter_configuration_improper_value_1() -> None:
    """Test the default configuration."""
    with pytest.raises(ValueError, match="Input should be greater than or equal to 0"):
        _ = QuotaLimiterConfiguration(
            type="cluster_limiter",
            name="cluster_monthly_limits",
            initial_quota=-1,
            quota_increase=10,
            period="3 seconds",
        )


def test_quota_limiter_configuration_improper_value_2() -> None:
    """Test the default configuration."""
    with pytest.raises(ValueError, match="Input should be greater than or equal to 0"):
        _ = QuotaLimiterConfiguration(
            type="cluster_limiter",
            name="cluster_monthly_limits",
            initial_quota=1,
            quota_increase=-10,
            period="3 seconds",
        )


def test_quota_limiter_configuration_improper_value_3() -> None:
    """Test the default configuration."""
    with pytest.raises(
        ValueError, match="Input should be 'user_limiter' or 'cluster_limiter'"
    ):
        _ = QuotaLimiterConfiguration(
            type="unknown_limiter",
            name="cluster_monthly_limits",
            initial_quota=1,
            quota_increase=10,
            period="3 seconds",
        )
