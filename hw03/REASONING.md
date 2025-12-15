# Reasoning

This document serves for adding reasoning to the choices
made throughout the assignment implementation and for other observations.

## Router/web framework choice

I decided to implement the API using only the standard library net/http, without using external frameworks like Gin, Echo, or Fiber.

### Here are the main reasons for my choice:

#### Simplicity and no dependencies: 
I wanted to keep the project clean. Since this is a homework assignment, adding external libraries felt unnecessary. It is easier to build and run the project when it has zero dependencies.

#### New Go 1.22 features: 
With the recent updates to Go (version 1.22), the standard http.ServeMux is much more powerful. It now natively supports method-based routing (like POST /reviews) and path parameters (like {id}). Previously, I would have needed a router like Chi for this, but now the standard library handles it perfectly.

#### Learning experience: 
I believe it is better to understand the basics first. Frameworks often hide the underlying logic behind "magic" functions. By using net/http, I have full control over the request and response handling, which helps me understand how building a REST API in Go actually works.