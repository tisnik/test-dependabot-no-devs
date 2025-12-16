# Reasoning

This document serves for adding reasoning to the choices
made throughout the assignment implementation and for other observations.

## Router/web framework choice

For implementing the API, I decided to use the Chi router (`github.com/go-chi/chi/v5`). The main reasons for choosing Chi are:

### 1. Fully compatible with net/http

Chi does not hide the standard Go HTTP model behind abstractions. All handlers are standard `http.HandlerFunc`, which makes the code simple.

### 2. Clean and readable router groups

Chi provides elegant and minimalistic route grouping.

### 3. Lightweight but powerful

Chi remains extremely small in terms of dependencies and runtime overhead while still offering features that the standard library lacks.
It sits in a sweet spot between barebone `net/http` and heavier frameworks like Gin or Fiber.

### 4. Great documentation

Chi documentation is minimalistic and easy to follow. It requires no magic and maintains idiomatic Go design — exactly what I needed.

### 5. No need for legacy `basePath` hacks

OpenAPI 3.x removed `basePath`. Chi makes mounting path segments easy via `Route()` and `Mount()`, eliminating the need for custom API-root logic.

---

## Code Generation Tool: oapi-codegen

To generate request/response models and server interfaces from the OpenAPI specification, I used **oapi-codegen**.

### 1. Generates Go code aligned with Chi and net/http

Using `oapi-codegen --generate chi-server` produces:

* request/response models
* handler interfaces
* router wiring
* compliance with the OpenAPI spec


### 2. Clear separation of concerns

`main.go` stays clean — it only mounts the generated router. All logic resides in the handler implementations.

## OpenAPI 3.1.0 → 3.0.3 Downgrade

The original specification was written in OpenAPI 3.1.0, but **oapi-codegen currently supports only 3.0.x**, so a downgrade to 3.0.3 was necessary.

### Notes:

* No functional endpoints required changes.
* Only top-level 3.1 syntax (such as `$schema` and draft-2020-12 features) was removed.
* All definitions remain valid and equivalent.

---

## Use of Swagger Editor (editor-next.swagger.io)

It helped me with:

* validating OpenAPI syntax
* visually inspecting endpoints
* checking generated examples
* verifying response schemas
* previewing status codes and payloads
* ensuring handler logic matched the spec
* writting tests

---

## Summary

The combination of **Chi + oapi-codegen + Swagger Editor** allowed me to build a clean, maintainable, and specification-consistent API implementation.