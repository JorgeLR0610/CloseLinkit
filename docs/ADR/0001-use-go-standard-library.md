# ADR 0001: Use Go Standard Library (net/http)

## Status
Accepted

## Context
For the `CloseLinkit` URL shortener service, we need an HTTP server and router to handle REST API requests (such as creating shortened URLs, retrieving stats, and resolving short codes for redirection).
There are several third-party web frameworks (e.g., Gin, Echo, Fiber) and routing packages (e.g., Chi, Gorilla Mux) available in the Go ecosystem.
However, introducing third-party routing frameworks increases dependency complexity, can affect compilation times, and risks API compatibility changes over time.
With modern Go versions (Go 1.22+), the standard library's `net/http` package features an enhanced `http.ServeMux` that natively supports HTTP method matching (e.g., `GET`, `POST`) and wildcards/path parameters.

## Decision
We will use the Go standard library's `net/http` package and its native `http.ServeMux` router for all HTTP request handling, routing, and middleware integration. We will not use external routing frameworks or web engines.

## Consequences
* **Positive:**
  * **Minimal Dependencies:** Zero third-party packages for routing, minimizing security vulnerabilities, dependency management overhead, and compatibility issues.
  * **Stability and Compatibility:** Leverages the robust backward-compatibility guarantees of the Go standard library.
  * **Standard Interfaces:** Handlers will use standard `http.Handler` and `http.HandlerFunc` signatures, making them universally compatible with standard Go libraries and middleware.
  * **Simplicity:** Codebase remains straightforward, using native language constructs.
* **Negative:**
  * **Manual Boilerplate:** JSON serialization/deserialization, error response formatting, and validation helper functions must be implemented manually or with minimal custom utility functions.
  * **Middleware Ecosystem:** Common utility middleware (like CORS, logger, recovery, rate-limiting) must be written from scratch or integrated using lightweight, standard-compliant external packages if necessary, rather than using built-in framework features.