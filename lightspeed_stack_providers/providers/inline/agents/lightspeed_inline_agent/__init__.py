from typing import Any

from llama_stack.core.datatypes import AccessRule
from llama_stack_api import Api

from .config import LightspeedAgentsImplConfig


async def get_provider_impl(
    config: LightspeedAgentsImplConfig,
    deps: dict[Api, Any],
    policy: list[AccessRule],
):
    """
    Create and initialize a LightspeedAgentsImpl.
    
    Sets litellm.drop_params = True so unsupported model parameters are removed globally.
    Parameters:
        config (LightspeedAgentsImplConfig): Configuration for the LightspeedAgentsImpl.
        deps (dict[Api, Any]): Mapping of Api enum keys to service implementations. Must include keys
            Api.inference, Api.vector_io, Api.tool_runtime, Api.tool_groups, Api.conversations,
            Api.prompts, Api.files, and Api.connectors. The Api.safety entry may be omitted.
        policy (list[AccessRule]): Access rules to apply to the created implementation.
    
    Returns:
        LightspeedAgentsImpl: The initialized LightspeedAgentsImpl instance ready for use.
    """
    import litellm

    litellm.drop_params = True

    from .agents import LightspeedAgentsImpl

    impl = LightspeedAgentsImpl(
        config,
        deps[Api.inference],
        deps[Api.vector_io],
        deps.get(Api.safety),
        deps[Api.tool_runtime],
        deps[Api.tool_groups],
        deps[Api.conversations],
        deps[Api.prompts],
        deps[Api.files],
        deps[Api.connectors],
        policy,
    )
    await impl.initialize()
    return impl
