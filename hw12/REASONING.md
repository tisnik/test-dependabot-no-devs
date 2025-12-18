# Reasoning

## Router/web framework choice
For this project, I have chosen to use the `chi` router.

`chi` is a lightweight, idiomatic and composable router for building Go HTTP services. It's a great choice for this project because:

*   **Standard Library Compatibility**: It's built on top of the `net/http` package, which means it's compatible with the standard library and many other `net/http`-based libraries.
*   **Middleware**: It has a great middleware ecosystem, which can be used for logging, authentication, and other cross-cutting concerns.
*   **URL Parameters**: It provides a simple and efficient way to handle URL parameters, which is needed for the recommendation endpoint.
*   **Lightweight**: It's very lightweight and doesn't have a lot of boilerplate, which makes it easy to get started with.
