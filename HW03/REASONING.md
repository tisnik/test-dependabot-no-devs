# Reasoning

This document serves for adding reasoning to the choices
made throughout the assignment implementation and for other observations.

## Router/web framework choice

I chose Gin (github.com/gin-gonic/gin) as the web framework for this project.

Why Gin?

- Industry standard - most popular Go web framework I've seen
- High performance
- Production ready - used by companies
- Developer productivity - clean API with less boilerplate
- Built-in JSON handling
- Easy parameter extraction - simple c.Param()  c.Query() methods
- Perfect for REST APIs - natural fit for OpenAPI specification
- easy to read documentation and community support

Alternatives considered:
- net/http: too low-level, requires more boilerplate
- Chi: minimal but lacks built-in JSON handling
- Fiber: not compatible with standard library

Gin provides good balance of performance, features, and seems to be ease to use.

