# Reasoning

This document serves for adding reasoning to the choices
made throughout the assignment implementation and for other observations.

## Router/web framework choice

Initially, I wrote everything by myself, but then I realized how useful the oapi-codegen tool is, so I generated the interfaces and structs and implemented it that way. I had to modify openapi.yaml version to older one so it is compatible with oapi-codegen

I chose the Chi framework because it’s lightweight, minimalist, and compatible with the Go standard library (and oapi-codegen). Its documentation is easy to follow—with emojis. Chi also offers a convenient implementation of subrouters (which I initially used, but then I decided to switch to generated code with oapi-codegen for other reasons). It is also simple to organize routes into meaningful groups—for example, for separating public and private endpoints (which I also did not use in the end, but I find it useful in many cases).
