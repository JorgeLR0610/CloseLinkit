# Architecture

## Overview

CloseLinkit is a URL shortening service composed of a Go backend, a React frontend and a PostgreSQL database.

The system exposes a REST API and a web client that allows users to create, retrieve and resolve shortened URLs.

---

## Technology Stack

| Component | Technology |
|----------|------------|
| Backend | Go (net/http) |
| Frontend | React + TypeScript + Vite |
| Web Server & Reverse Proxy | Nginx (nginxinc/nginx-unprivileged) |
| Database | PostgreSQL |
| SQL Code Generation | sqlc |
| Migrations Tool | goose |
| Security & Auth | Argon2id, JWT (golang-jwt/jwt/v5) |
| Containerization | Docker Compose, Distroless Static (ADR 0011), Nginx Alpine (ADR 0012) |

---

## System Components

### Backend (Go)

Implements the REST API, business logic, security/auth primitives, and communication with the database.

### Frontend (React + Nginx)

Provides the graphical user interface (GUI).
- **Development**: Runs via Vite development server (`Dockerfile.dev`) with Hot Module Replacement (HMR).
- **Production**: Packaged into a multi-stage Docker container (`frontend/Dockerfile`) using `nginxinc/nginx-unprivileged:1.30.5-alpine3.24` (ADR 0012). Nginx serves the compiled SPA assets and acts as a reverse proxy, forwarding `/api/`, short URLs (`~ "^/[a-zA-Z0-9]{7}$"`), and documentation (`/docs/`, `/openapi.yaml`) to the backend upstream (`API_UPSTREAM`).

### PostgreSQL

Persists application data.

### Docker Compose

Provides the local development environment by orchestrating the application services.

---

## Backend Layers

### Handler Layer

Receives HTTP requests, validates input, invokes the service layer, and builds HTTP responses.

### Middleware Layer

Provides cross-cutting HTTP request handling:
- **`ClientIP`**: Resolves the client IP once per request and stores it in context (`clientIPKey`). When `TRUST_PROXY=true`, it parses and validates the `X-Real-IP` header set by the Nginx reverse proxy (`net.ParseIP`); otherwise it falls back to the connection's `RemoteAddr` to prevent IP spoofing.
- **`RequestLogging`**: Structured JSON logging of request status, duration, and client IP via `GetClientIP`.
- **`RateLimiting`**: Token bucket rate limiting partitioned by client IP via `GetClientIP`.
- **`RequireAuth` & `OptionalAuth`**: JWT validation and context injection.
- **`CORSMiddleware`**: Cross-origin policy enforcement.
- **`Recover` & `RequestID`**: Panic safety and end-to-end request tracing UUIDs.

### Service Layer

Implements the application's business logic, user authentication, URL claiming, and coordinates domain operations.

### Security Layer

Provides Argon2id password hashing and constant-time verification against timing attacks, as well as JWT access and refresh token generation and cryptographic validation (`internal/security/`).

### Repository Layer

Provides database access through SQLC-generated queries.

### Database (PostgreSQL)

Persists application data.

---

## Authentication

CloseLinkit uses short-lived JWT access tokens and rotating refresh tokens.
See ADR 0010 for the rationale.

### Flow

```text
Login / Register
  Client ── POST /api/v1/auth/login ──▶ API
  Client ◀── access_token (JSON body) + refresh_token (HttpOnly cookie) ── API

Authenticated request
  Client ── Authorization: Bearer <access_token> ──▶ API

Session restore / token expiry
  Client ── POST /api/v1/auth/refresh (cookie sent automatically) ──▶ API
  Client ◀── new access_token + user + rotated refresh cookie ── API

Logout
  Client ── POST /api/v1/auth/logout ──▶ API (revokes token, clears cookie)
```

### Token storage
- The access token is kept only in memory on the client (never in localStorage).
- The refresh token is an opaque random value; only its SHA-256 hash is stored in
  PostgreSQL. The cookie is HttpOnly, Secure, SameSite, scoped to `Path=/api/v1/auth`.
- Each refresh revokes the used token and issues a new one (rotation).
- Expired refresh tokens are deleted by an hourly background job.

### Route protection
| Route | Middleware | Behavior |
|-------|-----------|----------|
| `POST /api/v1/shorten` | `OptionalAuth` | Anonymous: URL expires in 7 days. Authenticated: permanent, owned by the user. |
| `GET /api/v1/urls` | `RequireAuth` | Lists the authenticated user's URLs. |
| `POST /api/v1/urls/claim` | `RequireAuth` | Transfers anonymous URLs to the user. |
| `DELETE /api/v1/urls/{shortCode}` | `RequireAuth` | Deletes a shortened URL owned by the authenticated user. |
| `GET /api/v1/{shortCode}/stats` | `RequireAuth` | Retrieves access counts and statistics for a short code. |

### Claiming anonymous URLs
Guests keep their links in localStorage (`history`). After login, registration, or
session restore, the client sends the short codes to `POST /api/v1/urls/claim`. The
backend assigns ownership only to URLs with no owner and clears their expiration.

---

## Request Flow

```text
                     Browser / Client
                            │
                            ▼
              Nginx Container (Reverse Proxy)
              ┌─────────────┴─────────────┐
              │                           │
         [SPA / Assets]          [/api/, /{shortCode}, /docs/]
              │                           │
              ▼                           ▼
      Static Files / HTML          Go Handler Layer
                                          │
                                          ▼
                                    Service Layer
                                          │
                                          ▼
                                   Repository Layer
                                          │
                                          ▼
                                      PostgreSQL
```

The response follows the same path in reverse.

---

## Roadmap

### Implemented
- User authentication (JWT + refresh tokens), user accounts, per-user URL listing
- Anonymous URL expiration (7 days) and claiming
- User URL deletion (`DELETE /api/v1/urls/{shortCode}`) and protected analytics (`GET /api/v1/{shortCode}/stats`)
- Hardened Distroless runtime image for API (ADR 0011)
- Unprivileged Nginx runtime image for frontend with reverse proxy (ADR 0012)

### Planned
- Link editing (updating target URL)
- Custom short URLs (user-specified aliases)

## Infrastructure

### Implemented
- CI: backend (tests, golangci-lint, govulncheck), frontend (tests, lint, format check),
  secrets scan.
- Minimal static distroless runtime image (`gcr.io/distroless/static-debian13:nonroot`, ADR 0011).
- Unprivileged Nginx runtime image for frontend (`nginxinc/nginx-unprivileged:1.30.5-alpine3.24`, ADR 0012).
- API and Frontend Docker images published to GHCR on `v*.*.*` tags (`docker-publish.yaml`)
  with path-filtered change detection on pull requests.
- Multi-tier production deployment for AWS EC2 + RDS (`deploy/`):
  - Frontend EC2 instance running unprivileged Nginx reverse proxy and React SPA (`deploy/compose.web.yaml`).
  - Backend API EC2 instance running minimal Go Distroless container (`deploy/compose.api.yaml`).
  - Managed AWS RDS PostgreSQL instance for relational persistence.
  - Private VPC inter-service communication and least-privilege Security Groups (`deploy/README.md`).

### Planned
- Custom domain setup with TLS termination (ACM / ALB / CloudFront)
- Kubernetes manifests