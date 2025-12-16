# Reasoning

This document serves for adding reasoning to the choices
made throughout the assignment implementation and for other observations.

## Router/web framework choice

I chose Chi. Originally, I wanted to use the middleware for validation. However, then I found out I would either need to read the body twice or store the validated result in a context, both of which I did not like. As I dont use nested routes, the only extra thing I got form using chi was the chi.URLParam utility.
