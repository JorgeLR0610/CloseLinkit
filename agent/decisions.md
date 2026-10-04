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

### Anonymous URL Expiration

Anonymous URLs get `expires_at = now + 7 days` on creation. Authenticated URLs have
`expires_at IS NULL`. `GetURL` filters expired rows, and a cleanup job
(`StartExpiredURLsCleanup`, every 12h) deletes them.

### URL Claiming

`POST /api/v1/urls/claim` runs a single atomic update:
`WHERE short_code = ANY(...) AND user_id IS NULL`, so URLs owned by someone else are
never modified. It sets `user_id` and clears `expires_at`. The client removes the
`history` key from localStorage only after a successful response. Claiming runs on
login, registration auto-login, and session restore.

### Single Route for Guests and Users

`/` shows localStorage links for guests and API links for authenticated users.

### URL Deletion Semantics

`DELETE /api/v1/urls/{shortCode}` performs an atomic delete with `WHERE short_code = $1 AND user_id = $2`.
If 0 rows are affected, the service returns `404 Not Found` rather than `403 Forbidden` to prevent
leaking whether a URL exists under another user's account.

### Timing Attack Mitigation in Login

When an email is not found in the database during login, the service executes a simulated Argon2id
password hash verification with a dummy hash before returning 401. This produces consistent execution
times and prevents email enumeration via timing discrepancies.

## Pending Decisions

* Whether custom short URLs should reuse the current short-code generation service.
* Whether dockerize the frontend to serve the static files with Nginx on EC2 or use S3 + CloudFront.