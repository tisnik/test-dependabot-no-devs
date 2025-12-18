# Reasoning

This document serves for adding reasoning to the choices
made throughout the assignment implementation and for other observations.

## Router/web framework choice

For this project, I chose to use oapi-codegen to generate server code from the OpenAPI specification in openapi.yaml file in combination with the Chi framework.
I liked this approach because Chi allows clean and flexible routing, making it easy to organize handlers in a readable way. Chi also integrates smoothly with oapi-codegen, which automates much of the boilerplate code for request validation, parameter parsing, and response serialization.
Using oapi-codegen required me to extend the openapi.yaml with x-oapi-codegen-extra-tags to properly handle validation. This improved consistency between the API specification and the server implementation.
Another advantage of this approach is that the generated code enforces compliance with the OpenAPI specification, reducing the risk of inconsistencies or missing fields in request/response handling.
