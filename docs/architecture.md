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
| Database | PostgreSQL |
| SQL Code Generation | sqlc |
| Migrations Tool | goose |
| Containerization | Docker Compose |

---

## System Components

### Backend (Go)

Implements the REST API, business logic, and communication with the database.

### Frontend (React)

Provides the graphical user interface (GUI) and communicates with the backend through HTTP requests.

### PostgreSQL

Persists application data.

### Docker Compose

Provides the local development environment by orchestrating the application services.

---

## Backend Layers

### Handler Layer

Receives HTTP requests, validates input, invokes the service layer, and builds HTTP responses.

### Service Layer

Implements the application's business logic and coordinates domain operations.

### Repository Layer

Provides database access through SQLC-generated queries.

### Databae (PostgreSQL)

Persists application data.

---

## Request Flow

```text
                Browser
                   │
                   ▼
             React Frontend
                   │
              HTTP / JSON
                   │
                   ▼
             Handler Layer
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

### Planned Features

- Analytics dashboard
- Custom short URLs
- User authentication (JWT)
- User accounts
- Link management

## Infrastructure

- Docker image publishing
- CI/CD pipeline
- AWS deployment
- Kubernetes manifests