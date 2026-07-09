# ADR 0006: Use API Versioning

## Status
Accepted

## Context
As the `CloseLinkit` service grows, its REST API endpoints, request bodies, and response structures will inevitably evolve.
For example, v0.1 provides simple anonymous shortening and basic redirects. Future versions plan to introduce JWT authentication, user accounts, customizable redirection options, password protection, and advanced analytics.
Without versioning, modifying existing JSON structures or endpoint behaviors would introduce breaking changes, disrupting existing clients (such as mobile apps, browser extensions, or frontends).
We need a versioning mechanism that:
1. Prevents breaking changes for existing consumers.
2. Is clear, transparent, and easy to consume.
3. Maps cleanly to our directory and code structure.

## Decision
We will use URL path-based versioning for the REST API, starting with the `/api/v1/` prefix.
* For example, the shortening endpoint is exposed as `POST /api/v1/shorten`.
* The codebase mimics this structure by grouping controllers/handlers in directory-based packages, such as `internal/api/v1/`.

## Consequences
* **Positive:**
  * **Backward Compatibility:** Future changes (such as `/api/v2/`) can be launched side-by-side without affecting clients integrated with `/api/v1/`.
  * **Code Organization:** Clean package boundaries (e.g., `internal/api/v1` and eventually `internal/api/v2`) ensure handlers for different API versions do not pollute each other.
  * **Developer Experience:** Path-based versioning is highly visible, standard, and self-documenting for API consumers.
  * **Routing Simplicity:** Fits naturally with Go's `http.ServeMux` path matching.
* **Negative:**
  * **Code Duplication Risk:** If business logic or types are shared but slightly modified between versions, duplication can occur unless carefully factored out into shared service layers.
  * **Maintenance Overhead:** The team must maintain and support multiple version endpoints simultaneously until deprecated versions are formally retired.
