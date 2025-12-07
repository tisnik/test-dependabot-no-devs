"""REST API routers."""

from fastapi import FastAPI

from app.endpoints import (
    info,
    models,
    shields,
    providers,
    root,
    query,
    health,
    config,
    feedback,
    streaming_query,
    authorized,
    conversations,
    conversations_v2,
    metrics,
    tools,
    # V2 endpoints for Response API support
    query_v2,
)


def include_routers(app: FastAPI) -> None:
    """
    Register and mount project routers on the given FastAPI application.
    
    Registers endpoint routers and assigns URL prefixes: most endpoints are mounted under "/v1", v2 routers under "/v2", and core endpoints (health, authorized, metrics) are mounted without a version prefix.
    
    Parameters:
        app (FastAPI): The FastAPI application to which routers will be attached.
    """
    app.include_router(root.router)

    app.include_router(info.router, prefix="/v1")
    app.include_router(models.router, prefix="/v1")
    app.include_router(tools.router, prefix="/v1")
    app.include_router(shields.router, prefix="/v1")
    app.include_router(providers.router, prefix="/v1")
    app.include_router(query.router, prefix="/v1")
    app.include_router(streaming_query.router, prefix="/v1")
    app.include_router(config.router, prefix="/v1")
    app.include_router(feedback.router, prefix="/v1")
    app.include_router(conversations.router, prefix="/v1")
    app.include_router(conversations_v2.router, prefix="/v2")

    # V2 endpoints - Response API support
    app.include_router(query_v2.router, prefix="/v2")

    # road-core does not version these endpoints
    app.include_router(health.router)
    app.include_router(authorized.router)
    app.include_router(metrics.router)