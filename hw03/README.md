# ReelGoofy: REST API & Testing

ReelGoofy is movie recommendation service.
It ingests movie reviews created by users and generates recommendations
based on the review data.

This project serves as a template for the homework assignment.
To learn more about the homework assignments in general, visit
the [homework](https://github.com/course-go/homework) repository.

## Assignment

Throughout this homework assignment you will implement a REST API
service with a simple recommendation algorithm.

![ReelGoofy diagram](assets/reelgoofy.svg)

The above diagram provides a visualization of the possible service component
architecture and its possible relationship to other services. This only serves
as an example and does not dictate the actual structure of your implementation.
However, if you are unsure how to structure your project, it is suggested
to follow the diagram.

To implement the API you can choose whatever router or web framework you want.
Here are just some of the possible options:

- [net/http](https://pkg.go.dev/net/http)
- [chi](https://github.com/go-chi/chi)
- [gin](https://github.com/gin-gonic/gin)
- [echo](https://github.com/labstack/echo)
- [fiber](https://github.com/gofiber/fiber)

Whatever your choice will be, document it in the `REASONING.md` file.
Describe why you chose the library, what did you like about it, what
additional features it offers compared to the standard library etc.

### Specification

The API is specified using
the [OpenAPI](https://spec.openapis.org/oas/latest.html) standard in version 3.1.0.
You can work with the OpenAPI just in plain text or use some visualization tools
like the new version of [Swagger editor](https://editor-next.swagger.io).
You can also play around with tools that generate code from the specification
or that verify your solution complies with the specification.
If you decide to do so, please note your observations to the `REASONING.md` file.
The specification itself can be found in the `docs` directory.

#### Recommendations

The API will require you to implement a recommendation algorithm.
There is no need to come up with a complex solution. A simple one will do.

The ingested data gives you a lot of flexibility on how to approach
this, as it provides many arguments. Feel free to use, aggregate, or ignore
them as you wish. The data should, for now, be just stored in memory.
We will look at persistence in the following homework.

### Testing

To verify your solution does what it is suppose to, you need to test it.
The suggested approach is to write Go API tests
[in parallel](https://martinfowler.com/bliki/TestDrivenDevelopment.html)
(see [Bonus](#bonus)) to coding the actual implementation.
If you do not want to implement automated tests you can just manually
test the application using tools like `curl`
(you can generate the commands from the specification in the editor) or
GUI applications like
[Postman](https://www.postman.com/)
or [Insomnia](https://insomnia.rest/)
(you can import the specification and they will generate request
collection for you).

### Bonus

You can also gain up to **5 bonus points** for
implementing a test suite testing the API in Go.

- **2 points** for covering ingest endpoints
- **3 points** for covering the recommendation endpoints

A minimal set of tests just to cover the basic functionality will do.

If you decide to implement the bonus, please do so in an additional commit/s.

## Requirements

The implemented API complies with the OpenAPI specification.
The router/web framework choice is documented in the specified file.
All code is written in a idiomatic way, it is reasonably structured,
contains no data races and complies with common Go conventions.

## Motivation

The main goal of this homework is to practice
implementing a REST API in Go.
Additionally, it showcases the OpenAPI specification commonly
used in practice and demonstrates its usage.

## Packages

For generating the IDs you can use
the [UUID](https://pkg.go.dev/github.com/google/uuid) package for Go.
