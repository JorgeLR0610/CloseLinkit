# ADR 0010: Access Token in Memory, Refresh Token in HttpOnly Cookie

## Status
Accepted

## Context
The frontend needs sessions that survive page reloads without exposing tokens to
JavaScript-readable storage (XSS risk).

## Decision
- Access tokens (JWT) are held only in memory on the client.
- Refresh tokens are opaque random values, stored hashed (SHA-256) in PostgreSQL,
  sent as an HttpOnly, Secure, SameSite cookie scoped to `Path=/api/v1/auth`, and
  rotated on every refresh.
- The session is restored on app load through `POST /api/v1/auth/refresh`.
- The refresh endpoint also accepts the token in the request body for non-browser clients.
- Passwords are hashed with Argon2id.

## Consequences
* **Positive:** tokens are not readable by injected scripts; revocation and rotation are possible.
* **Negative:** CORS must allow credentials with explicit origins; every page load makes one refresh request; expired tokens need periodic cleanup.