# ADR 0012: Dockerize Frontend with Nginx Reverse Proxy

## Status

Accepted

## Context

CloseLinkit provides a React single-page application (SPA) built with TypeScript and Vite. During local development, the frontend runs with Vite's development server (`Dockerfile.dev`) to support Hot Module Replacement (HMR).

For production deployment on an AWS EC2 instance, the static frontend assets must be served efficiently to clients, and API requests must reach the Go backend. Two primary architectures were considered:

1. **Object Storage and CDN (Amazon S3 + CloudFront):**
   Host static assets on an S3 bucket and distribute them via CloudFront. While performant and scalable, it introduces cross-origin considerations (CORS) between the frontend domain and the API domain. Since authentication relies on an `HttpOnly`, `SameSite`, `Secure` refresh cookie scoped to `/api/v1/auth` (ADR 0010), cross-origin requests require explicit cookie domain configuration and complex CORS handling. It also introduces additional AWS infrastructure components to manage during this project stage.

2. **Containerized Nginx on EC2:**
   Build a production Docker container that serves the compiled Vite assets using Nginx and acts as a reverse proxy forwarding `/api/`, short URLs (`/{shortCode}`), and Swagger documentation (`/docs/`) to the Go backend (`API_UPSTREAM`).

## Decision

Dockerize the frontend for production using a multi-stage Docker build and `nginxinc/nginx-unprivileged:1.30.5-alpine3.24` as the production runtime.

The Dockerfile (`frontend/Dockerfile`) will:

1. Use a Node 24 Alpine build stage (`node:24.18-alpine3.23`) to install dependencies via `npm ci` and compile production assets via `npm run build`.
2. Copy the resulting static bundle (`/app/dist`) into `/usr/share/nginx/html` in the unprivileged Nginx image.
3. Provide an Nginx configuration template (`frontend/default.conf.template`) mounted into `/etc/nginx/templates/default.conf.template`. The Nginx entrypoint automatically runs `envsubst` on startup, injecting the upstream address configured in `API_UPSTREAM`.
4. Route requests as follows:
   - `/api/`: proxied to `http://${API_UPSTREAM};` with standard forwarding headers (`Host`, `X-Forwarded-For`, `X-Forwarded-Proto`).
   - `~ "^/[a-zA-Z0-9]{7}$"`: proxied to `http://${API_UPSTREAM};` to allow 7-character short URL redirection on the Go backend.
   - `/docs/` and `= /openapi.yaml`: proxied to `http://${API_UPSTREAM};` to serve Swagger UI and the OpenAPI specification.
   - `/`: served from `/usr/share/nginx/html` with `try_files $uri /index.html;` to enable client-side React Router navigation.
5. Expose port `8080` (standard non-root port for `nginx-unprivileged`) and run as the unprivileged `nginx` user (UID 101).
6. Build and publish the image to GitHub Container Registry (GHCR) as `ghcr.io/<owner>/closelinkit-frontend` alongside `closelinkit-api` via `.github/workflows/docker-publish.yaml`.

## Consequences

* **Positive:**
  * **Same-Origin Simplicity:** The frontend and API share the same public origin. Browsers transmit HttpOnly session cookies automatically without CORS preflight overhead or cross-site cookie restrictions.
  * **Unified Deployment:** Both the backend and frontend are containerized and published to GHCR, enabling simple orchestration with Docker Compose on an EC2 instance.
  * **Least Privilege:** The container runs as an unprivileged user on port 8080, adhering to the principle of least privilege.
  * **Dynamic Configuration:** Upstream API target (`API_UPSTREAM`) is configurable via environment variables without rebuilding the Docker image.
  * **Clean Separation of Concerns:** Development retains Vite HMR via `Dockerfile.dev`, while production builds use an optimized static web server.

* **Negative:**
  * **Nginx Configuration Maintenance:** Routing rules (such as short code regex matching and Swagger endpoints) must be maintained in `default.conf.template`.
  * **Container Compute Overhead:** Serving static files from a container on EC2 consumes server resources compared to an S3/CloudFront serverless architecture. If traffic scales significantly in the future, transitioning to CloudFront can be revisited.
