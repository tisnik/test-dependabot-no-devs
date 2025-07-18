"""REST API routers."""

from fastapi import FastAPI

from app.endpoints import (
    info,
    models,
    root,
    query,
    health,
    config,
    feedback,
    streaming_query,
    authorized,
    conversations,
)


def include_routers(app: FastAPI) -> None:
    """
    Registers all API routers onto the provided FastAPI application instance.
    
    Routers for various endpoints are included, with most grouped under the `/v1` prefix for versioning. The health and authorization endpoints are included without a version prefix.
    """
    app.include_router(root.router)
    app.include_router(info.router, prefix="/v1")
    app.include_router(models.router, prefix="/v1")
    app.include_router(query.router, prefix="/v1")
    app.include_router(streaming_query.router, prefix="/v1")
    app.include_router(config.router, prefix="/v1")
    app.include_router(feedback.router, prefix="/v1")
    app.include_router(conversations.router, prefix="/v1")

    # road-core does not version these endpoints
    app.include_router(health.router)
    app.include_router(authorized.router)
