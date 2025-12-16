## Router/web framework choice

I chose to use the standard **`net/http`** package from Go.

I decided to use it because the project is simple, so we don't need a big framework. Also, since Go 1.22, the standard router is powerful enough. It can handle methods like `POST` and paths like `/reviews/{id}`. We don't need extra libraries for this anymore. Finally, using the standard library means we don't have to manage many external packages.

## Logging

I used the **`log/slog`** package.

I chose it because it formats logs as JSON automatically. This makes it easier to read logs in production. It is also built into Go, so we don't need to install external tools.

## Project Structure

I followed the standard Go project layout, similar to the `course-go/todos` example.

-   **`internal/api`**: Handles web requests (reading input, sending output).
-   **`internal/service`**: Contains the main logic (calculating recommendations).
-   **`internal/repository`**: Saves the data in memory.
-   **`internal/domain`**: Defines what the data looks like.
