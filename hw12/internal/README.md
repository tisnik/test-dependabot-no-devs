# Internal Packages

This directory contains the internal packages of the ReelGoofy service.

- `app`: Contains the main application logic for starting and running the service. It's responsible for wiring all the components together (storage, services, handlers).

- `handler`: Contains the HTTP handlers for the API endpoints. These handlers are responsible for parsing requests, calling the appropriate services, and writing responses.

- `model`: Defines the core data structures (domain models) used throughout the application, such as `Review` and `Recommendation`.

- `service`: Implements the business logic of the application. This includes the recommendation algorithm and any other logic that is not directly related to handling HTTP requests or storing data.

- `storage`: Provides an abstraction layer for data persistence. It defines interfaces for data access and includes concrete implementations.
