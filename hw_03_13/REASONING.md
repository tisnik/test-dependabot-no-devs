# Reasoning

Wow, I really enjoyed working on this project :D I hope I managed to handle 
it successfully and didn’t overengineer things too much. That said, I’m not
entirely satisfied with some of the features—especially in terms of performance
and usability—but given the lack of a proper database layer, there wasn’t much
else I could do.

I submitted it by the original deadline (27.11.) and tried to complete the 
bonus part by adding tests.

## Router / Web Framework Choice
I chose Chi as the framework because I like its simplicity and its
compatibility with the standard library. It seems to provide sufficient
capabilities for this project. For request validation, I used basic checks
provided by https://github.com/go-playground/validator and for testing I have
created a seeder which is using Faker (https://github.com/jaswdr/faker/) for
real-looking data.

## Running the application
Application can be run via Makefile using command `make run`.

If you want to specify port, request timeout, or run a seeder to create
initial data, use following command in the directory with your binary file:

`./reelgoofy [COMMAND] [OPTIONS]`
```
Commands:
seed-reviews        Seed the reviews dataset

Options:
--port, -p PORT     Port to bind the server to
--timeout, -t SEC   Request timeout in seconds
```

Examples:

`./reelgoofy --port 3000 --timeout 30`

`./reelgoofy seed-reviews --port 9000`

## Recommendation Logic
In a real system with thousands of records, recommending content based on 
multiple attributes could be challenging. I implemented a simplified
cosine similarity algorithm to address this. Cosine similarity measures the
similarity between two vectors. When a new review is stored, a vector is 
computed from some of its attributes and saved alongside the review.

When recommending content based on a `userId`, the "user's profile" is first
computed. This profile is calculated as an average of the user's top-rated 
reviews, with emphasis on their scores, resulting in a vector that can then
be compared with existing reviews.

If no limit and offset is specified, API returns 3 recommendations as default.

## Project Structure
I'm not entirely sure if I chose the optimal structure for a Go project,
but I followed a pattern similar to what I use at work, which seemed practical
for navigating the codebase. The `internal` directory contains two main 
subfolders: `core` and `containers`.

- `core` holds logic shared across the system, such as API response formatting, 
request validation, etc.
- `containers` contains individual logical modules (like `recommendations` 
and `reviews`), each encapsulating a separate part of the application logic.

```
reelgoofy
│   ...
│
└───cmd/reelgoofy/
│   │   main.go
│   
└───internal/
│   │
│   └───containers/
│   │   │   recommendations/
│   │   │   reviews/
│   │
│   └───core/
│       │   request/
│       │   response/
│       │   server/
│       │   utils/
│  
└───routes/api/v1/
    │   routes.go

```

## Testing

For this task, I developed two sets of tests designed to cover all
application endpoints as well as part of the business logic. 
The CRUD operation tests check not only the correct statuses but also 
error cases. The recommendation tests focus on verifying the correctness
of the responses. Additionally, the tests also validate the logic used for
comparing the similarity of reviews alongside the API responses.

For testing purposes, I created a seeder that uses 
Faker (https://github.com/jaswdr/faker/) to generate realistically-looking data.