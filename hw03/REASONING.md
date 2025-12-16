# Reasoning

This document serves for adding reasoning to the choices
made throughout the assignment implementation and for other observations.

## Router/web framework choice

My goal for this assignment was to gain universally applicable knowledge of HTTP and API implementation in Go, 
therefore I avoided selecting a heavier framework.
I chose chi router as the middle ground between pure `net/http` and more complex frameworks.
Chi allowed me to improve the readability and simplicity of code, while still allowing me to work with Go's HTTP
package directly.

## Domain model decomposition

Although the OpenAPI specification defines the /reviews ingestion input
as a single object, I chose to decompose the Review DTO object into two distinct entities internally.
My interpretation of the movie recommendation service is that each movie exists only once, uniquely identified 
by its `contentID`, and may have multiple user reviews.
The `Content` entity represents all movie metadata, which may be optional in the API ingest.
The `Review` entity contains only the user-specific interaction data.
I chose this approach to eliminate data redundancy, with the future addition of persistence
in mind.