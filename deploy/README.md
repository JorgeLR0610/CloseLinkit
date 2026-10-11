# Production Deployment Guide (AWS EC2 + RDS)

This directory contains the production Docker Compose configurations and environment templates to deploy **CloseLinkit** across separate AWS infrastructure components:

- **Frontend Instance (EC2)**: Serves the React Single-Page Application (SPA) and acts as an unprivileged Nginx reverse proxy (`compose.web.yaml`).
- **Backend API Instance (EC2)**: Runs the compiled Go REST API in a minimal Distroless container (`compose.api.yaml`).
- **Database (AWS RDS PostgreSQL)**: Managed relational database.

Both EC2 instances reside within the same AWS Virtual Private Cloud (VPC) for secure, private inter-service communication.

---

## Architecture & Network Topology

```text
       Internet / Client Browser
                   │
                   ▼ (HTTP:80)
   ┌────────────────────────────────┐
   │    Frontend EC2 (Public IP)    │
   │  ┌──────────────────────────┐  │
   │  │   Nginx Reverse Proxy    │  │
   │  │       (Port 80:8080)     │  │
   │  │  - React SPA Assets      │  │
   │  │  - /api/, /{id}, /docs/  │  │
   │  └─────────────┬────────────┘  │
   └────────────────┼───────────────┘
                    │ Private VPC Subnet (HTTP:8080)
                    ▼
   ┌────────────────────────────────┐
   │    Backend EC2 (Private IP)    │
   │  ┌──────────────────────────┐  │
   │  │   Go REST API Container  │  │
   │  │      (Port 8080:8080)    │  │
   │  └─────────────┬────────────┘  │
   └────────────────┼───────────────┘
                    │ Private VPC Subnet (Postgres:5432)
                    ▼
   ┌────────────────────────────────┐
   │     AWS RDS (PostgreSQL 18+)   │
   └────────────────────────────────┘
```

---

## AWS Security Groups Setup

To implement least privilege security, configure three distinct Security Groups within the same VPC:

### 1. Frontend Security Group (`sg-closelinkit-frontend`)
| Type | Port Range | Source | Purpose |
| :--- | :--- | :--- | :--- |
| Inbound HTTP | `80` | `0.0.0.0/0` | Public web access |
| Inbound HTTPS | `443` | `0.0.0.0/0` | Secure web access (when TLS is configured) |
| Inbound SSH | `22` | `<your-admin-ip>/32` | Administrative access |
| Outbound | All | `0.0.0.0/0` | Standard egress |

### 2. Backend API Security Group (`sg-closelinkit-api`)
| Type | Port Range | Source | Purpose |
| :--- | :--- | :--- | :--- |
| Inbound Custom TCP | `8080` | `sg-closelinkit-frontend` | Allow traffic **only** from the Frontend EC2 instance |
| Inbound SSH | `22` | `<your-admin-ip>/32` | Administrative access |
| Outbound | All | `0.0.0.0/0` | Package updates & GHCR pulls |

> [!TIP]
> Do not expose port `8080` to `0.0.0.0/0`. Restricting inbound traffic to the Frontend Security Group (`sg-closelinkit-frontend`) prevents unauthorized public access to internal API ports.

### 3. Database Security Group (`sg-closelinkit-rds`)
| Type | Port Range | Source | Purpose |
| :--- | :--- | :--- | :--- |
| Inbound PostgreSQL | `5432` | `sg-closelinkit-api` | Allow database connections exclusively from Backend API |
| Inbound PostgreSQL | `5432` | `<your-admin-ip>/32` | Optional: allow schema migrations from admin workstation |
| Outbound | All | `0.0.0.0/0` | Standard egress |

---

## Step 0: Database Migrations (AWS RDS)

Before starting the API container, run Goose migrations against the RDS instance.

From your local machine (with access to RDS enabled in the RDS Security Group):

```bash
# Set DB_URL_GOOSE to your RDS connection string
export DB_URL_GOOSE="postgres://<db_user>:<db_password>@<rds_endpoint>:5432/<db_name>?sslmode=require"

# Check migration status
make migrate-status

# Apply all migrations (001 - 005)
make migrate-up
```

Verify that all tables (`urls`, `users`, `refresh_tokens`) are created in PostgreSQL before launching the backend service.

---

## Step 1: Deploy Backend API (EC2 Instance 1)

1. Connect to the Backend EC2 instance via SSH:
   ```bash
   ssh -i your-key.pem <user>@<backend-ec2-ip>
   ```

2. Install Docker and Docker Compose (if not already installed).

3. Create the project deployment directory:
   ```bash
   mkdir -p ~/closelinkit-deploy
   cd ~/closelinkit-deploy
   ```

4. Copy `deploy/compose.api.yaml` and create your `.env` file based on `.env.api.example`:
   ```bash
   nano .env
   ```

   **Production `.env` configuration example for the API:**
   ```env
   # Database connection string to AWS RDS
   DB_URL=postgres://closelinkit_admin:StrongPassword123!@closelinkit-db.cxxxxxx.us-east-1.rds.amazonaws.com:5432/closelinkit?sslmode=require

   # Allowed CORS Origins (Public address where users load the frontend)
   ALLOWED_ORIGINS=http://<frontend-ec2-public-ip>

   # Base URL for generated short links (Must be the frontend public address)
   BASE_URL=http://<frontend-ec2-public-ip>

   # Authentication Secret (Generate with: openssl rand -base64 32)
   JWT_SECRET=super_secret_jwt_random_string_here_32_bytes_min

   # Cookie Secure: set to false for plain HTTP testing via EC2 IP; set to true if using HTTPS/TLS
   COOKIE_SECURE=false

   # Trust proxy headers (X-Real-IP) forwarded by Nginx reverse proxy
   TRUST_PROXY=true
   ```

   > [!IMPORTANT]
   > `BASE_URL` **must** point to the frontend's public address (e.g., `http://<frontend-ec2-public-ip>`). The backend uses this to generate short links (`BASE_URL/<shortCode>`), which users visit through Nginx.

5. Start the backend container:
   ```bash
   docker compose -f compose.api.yaml up -d
   ```

6. Inspect logs to verify successful database connection:
   ```bash
   docker compose -f compose.api.yaml logs -f
   ```
   You should see:
   ```text
   level=INFO msg="server running" port=:8080
   ```

---

## Step 2: Deploy Frontend & Nginx (EC2 Instance 2)

1. Connect to the Frontend EC2 instance via SSH:
   ```bash
   ssh -i your-key.pem <user>@<frontend-ec2-ip>
   ```

2. Create the deployment directory:
   ```bash
   mkdir -p ~/closelinkit-deploy
   cd ~/closelinkit-deploy
   ```

3. Copy `deploy/compose.web.yaml` and create `.env` based on `.env.web.example`:
   ```bash
   nano .env
   ```

   **Production `.env` configuration example for the Frontend:**
   ```env
   # Backend API target: Private IP of the API EC2 instance WITH port 8080
   API_UPSTREAM=<backend-ec2-private-ip>:8080
   ```

   > [!IMPORTANT]
   > Ensure you append `:8080` to `API_UPSTREAM`. If omitted, Nginx will default to port 80 on the backend and connections will fail.

4. Start the frontend container:
   ```bash
   docker compose -f compose.web.yaml up -d
   ```

5. Inspect container logs:
   ```bash
   docker compose -f compose.web.yaml logs -f
   ```
   You should see Nginx startup messages confirming configuration generated with your upstream.

---

## Step 3: Verification & End-to-End Testing

1. **Frontend UI**: Open `http://<frontend-ec2-public-ip>` in your browser. The single-page application should render immediately.
2. **Shorten a URL**: Submit a URL in the UI. The generated short URL should appear as `http://<frontend-ec2-public-ip>/<shortCode>`.
3. **Redirect Resolution**: Click the short URL or paste it into a new browser tab. Nginx should proxy the 7-character route to the API, and the API should respond with an HTTP 302 redirect to the destination URL.
4. **Authentication**: Register or log in to a user account. Check browser DevTools -> Application -> Cookies to confirm the `refresh_token` cookie is present.
5. **Interactive Swagger Docs**: Navigate to `http://<frontend-ec2-public-ip>/docs` to inspect the Swagger UI and OpenAPI documentation loaded via the reverse proxy.

---

## Troubleshooting

- **502 Bad Gateway on `/api/` or short codes:**
  - Check that the Backend EC2 instance is running and healthy on port 8080 (`docker compose logs`).
  - Verify that `API_UPSTREAM` in `.env` includes `:8080` (e.g. `<private-ip>:8080`).
  - Check the Security Group of the Backend EC2 instance: verify that port 8080 allows inbound traffic from the Frontend Security Group.
  - From the Frontend EC2 shell, test connectivity:
    ```bash
    curl -I http://<backend-ec2-private-ip>:8080/docs/
    ```

- **Authentication / Refresh Token issues (401 on refresh):**
  - Verify `COOKIE_SECURE`: If testing via plain HTTP (`http://<ip>`), modern browsers discard cookies marked `Secure`. Ensure `COOKIE_SECURE=false` in the backend `.env`.
  - Verify `ALLOWED_ORIGINS`: If cross-origin requests are made, ensure `ALLOWED_ORIGINS` in the backend `.env` contains `http://<frontend-ec2-public-ip>`.

- **Database connection error on API startup:**
  - Verify that the RDS Security Group allows port 5432 from the Backend API Security Group.
  - Ensure the database user, password, and database name in `DB_URL` match your RDS configuration.
