# Changelog

## v0.3.1 (released)
- Production frontend container using unprivileged Nginx (`nginxinc/nginx-unprivileged:1.30.5-alpine3.24`, ADR 0012).
- Dynamic Nginx reverse proxy template routing `/api/`, 7-character short URLs (`~ "^/[a-zA-Z0-9]{7}$"`), and Swagger docs (`/docs/`, `/openapi.yaml`) to `${API_UPSTREAM}`.
- Multi-image GitHub Actions workflow publishing both `closelinkit-api` and `closelinkit-frontend` to GHCR on release tags with path-filtered PR checks.
- Production deployment configurations in `deploy/` for AWS EC2 (Frontend + Backend in same VPC) and AWS RDS PostgreSQL, accompanied by a step-by-step deployment guide (`deploy/README.md`).
- Added `ClientIP` middleware (`server/internal/middleware/clientIP.go`) with `TRUST_PROXY` support and `X-Real-IP` Nginx header forwarding for accurate client IP resolution in rate limiting and request logging.