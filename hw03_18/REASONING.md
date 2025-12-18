# Reasoning

This document serves for adding reasoning to the choices
made throughout the assignment implementation and for other observations.

## Router/web framework choice

I chose the Go standard library `net/http` for the HTTP stack because it’s stable, fast, and covers the core features this service needs without extra dependencies:
- HTTP client and server implementations
- `http.Handler` and `http.ResponseWriter` for clear, testable handlers.
- `net/http/httptest` to unit-test handlers and the router with in-memory requests/responses.

To keep routing ergonomic while staying lightweight, I added `chi` as a supplementary router. In this codebase I specifically use:
- Route grouping/versioning: `r.Route("/api/v1", ...)` to mount all endpoints under the API base path.
- Method-specific routes: `Post`, `Delete`, and `Get` for REST-style handlers.
- Path parameters: `{reviewId}`, `{contentId}`, `{userId}` resolved via `chi.URLParam`.

All in all, I was satisfied with my choice.
