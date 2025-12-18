# Reasoning

This document serves for adding reasoning to the choices
made throughout the assignment implementation and for other observations.

## My Router/web framework choice : Gin

I was deciding between the net/http package and the Gin framework, since they are the most commonly used options for tasks like this. After studying the features of both, I chose Gin. Choosing Gin seemed like better option than net/http because it handles many things automatically and feels more intuitive. As a beginner, I found it much easier to use.
With net/http, you need to manually handle routing, parameters, JSON parsing, error handling and response formatting. Gin provides all of this built in, so you write less code and make fewer mistakes. It also has great simple declarative routing, convenient handler management, a recovery mechanism, router groups, and JSON model binding.

I also used Swagger to visualize and test my API. To run it properly, I had to deploy Swagger Editor locally in Docker, as it could not load local files directly from the browser.