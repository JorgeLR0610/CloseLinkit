# Current State

**Version:** v0.3.0 · **Branch:** dockerize-frontend
**Stack:** Go + PostgreSQL (Goose, sqlc) + React/TS/Vite + Nginx. See `docs/architecture.md`.

## Done
Auth (register/login/refresh/logout), user URL listing and deletion, anonymous URL expiry
with cleanup jobs, cookie-based refresh, unified `/` view, URL claiming, stats authentication,
minimal Distroless API container (ADR 0011), production frontend container with unprivileged
Nginx reverse proxy (ADR 0012), multi-image GHCR publishing workflow, AWS EC2/RDS deployment
specifications, and ClientIP middleware with TRUST_PROXY support.
History: `docs/CHANGELOG.md`. Endpoints: `server/docs/openapi.yaml`.

## Active task
None (ClientIP middleware testing, TRUST_PROXY configuration, and documentation completed)

## Next steps
1. Edit URLs / per-link analytics.
2. Custom aliases (open question in `agent/decisions.md`).
3. Infra: Domain configuration with TLS termination and Kubernetes manifests.

## Watch out
- Cleanup goroutines (refresh tokens hourly, expired URLs every 12h) are started in
  `server/cmd/CloseLinkit/main.go`.