# Development Decisions

This file contains recent implementation decisions that are important
for continuity but do not necessarily warrant a formal ADR yet.

## Current Decisions

### URL Validation

The backend only accepts `http` and `https` URLs.

Thus, hosts such as:

* loopback addresses
* private IP addresses
* link-local addresses

are rejected.

This behavior is part of the current application behavior and should
not be changed unintentionally.

### Short Code Collision Handling

Short code generation is probabilistic.

The service attempts to insert the generated short code into PostgreSQL
and retries when the database reports the expected unique constraint
violation.

The current maximum retry count is 5.

## Pending Decisions

* Whether custom short URLs should reuse the current short-code generation service.
