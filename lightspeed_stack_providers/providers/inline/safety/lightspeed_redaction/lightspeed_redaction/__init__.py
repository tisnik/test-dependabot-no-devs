from typing import Any

from .config import RedactionShieldConfig


async def get_provider_impl(
    config: RedactionShieldConfig,
    deps: dict[str, Any],
):
    """
    Constructs and initializes a RedactionShield provider implementation from the given configuration and dependencies.
    
    Parameters:
        config (RedactionShieldConfig): Configuration for the RedactionShield provider.
        deps (dict[str, Any]): Runtime dependencies required by the implementation.
    
    Returns:
        RedactionShieldImpl: The initialized provider implementation.
    
    Raises:
        AssertionError: If `config` is not a `RedactionShieldConfig` (error message includes the actual type).
    """
    from .redaction import (
        RedactionShieldImpl,
    )

    assert isinstance(
        config, RedactionShieldConfig
    ), f"Unexpected config type: {type(config)}"
    impl = RedactionShieldImpl(config, deps)
    await impl.initialize()
    return impl
