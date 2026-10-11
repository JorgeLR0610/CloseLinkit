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

### Frontend Production Containerization & Reverse Proxy

The production frontend is packaged into a Docker container (`frontend/Dockerfile`) using unprivileged
Nginx (`nginxinc/nginx-unprivileged:1.30.5-alpine3.24`, ADR 0012). It serves the compiled React SPA assets
and acts as a reverse proxy forwarding `/api/`, 7-character short codes (`~ "^/[a-zA-Z0-9]{7}$"`), and
Swagger documentation (`/docs/`, `/openapi.yaml`) to the Go backend upstream (`API_UPSTREAM`). This allows
same-origin cookie handling without CORS complexity and unifies containerized deployment on EC2.

### AWS Multi-Tier EC2 + RDS Deployment Topology

Production deployment isolates tiers across distinct AWS infrastructure resources within the same VPC:
- **Frontend EC2 Instance**: Runs `deploy/compose.web.yaml` (unprivileged Nginx + React SPA) mapping port `80:8080`.
- **Backend API EC2 Instance**: Runs `deploy/compose.api.yaml` (Go Distroless API container) exposing port `8080:8080`. Security Group rules restrict inbound traffic on port 8080 to the Frontend Security Group only.
- **AWS RDS PostgreSQL**: Managed database instance accessible on port 5432 only by the Backend API Security Group. Migrations are executed with Goose (`make migrate-up`) from an authorized workstation prior to launching the API container.
- **Inter-service Communication (`API_UPSTREAM`)**: Because both EC2 instances reside within the same VPC, `API_UPSTREAM` specifies the private IP of the API instance and must include `:8080` (e.g., `<api_private_ip>:8080`).
- **Public Short Links (`BASE_URL`)**: `BASE_URL` on the API must be configured as the public frontend address (e.g., `http://<frontend_public_ip>`), ensuring generated short URLs (`BASE_URL/<shortCode>`) resolve through the Nginx proxy.
- **Session Cookie Security (`COOKIE_SECURE`)**: Configured as `false` when testing over raw HTTP using EC2 IP addresses (to prevent modern browsers from discarding the HttpOnly refresh token cookie), and set to `true` when custom domains with HTTPS/TLS termination are introduced.

### Client IP Resolution and Proxy Trusting (`TRUST_PROXY`)

Client IP extraction is centralized in `server/internal/middleware/clientIP.go`:
- **Nginx Forwarding**: `frontend/default.conf.template` sets `proxy_set_header X-Real-IP $remote_addr;`.
- **Proxy Trusting Flag**: Controlled via the boolean environment variable `TRUST_PROXY=true|false` (evaluated once in `main.go`).
- **IP Validation & Security**: When `TRUST_PROXY=true`, the `X-Real-IP` header is parsed with `net.ParseIP`. Only valid IP addresses are accepted; malformed headers automatically fall back to the connection's `RemoteAddr`.
- **Untrusted / Development Environments**: When `TRUST_PROXY=false` (default in dev), `X-Real-IP` is ignored, preventing IP spoofing from clients sending arbitrary headers directly to the API.
- **Context Propagation**: Resolved IP is stored in `r.Context()` with `clientIPKey{}`. Consumers (`RateLimiting`, `RequestLogging`) access it via `GetClientIP(r)`, which safely falls back to `RemoteAddr` when the middleware was not invoked (e.g. in isolated unit tests).

## Pending Decisions

* Whether custom short URLs should reuse the current short-code generation service.