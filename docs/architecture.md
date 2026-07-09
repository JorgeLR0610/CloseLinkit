# Architecture

## Objective

CloseLinkit is a RESTful URL shortening service.

The current version provides an HTTP API for creating, retrieving and resolving shortened URLs.

Future iterations will introduce a React frontend, user authentication and deployment automation.

## Components

### Backend (Go)

Exposes the REST API, implements the business logic and interacts with the database.

### PostgreSQL

Persists application data.

### Docker Compose

Provides the local development environment.

## Layer Responsibilities

### REST API

Responsible for handling HTTP requests and responses.

### Service Layer

Contains the application's business logic.

### Repository Layer

Provides data persistence and database access.

### PostgreSQL

Stores application data.

## Request Flow

```text
Client
↓
Handler
↓
Service
↓
Repository
↓
PostgreSQL
↓
Service
↓
Handler
↓
Client
```

## Future Architecture
Future versions are expected to include:

- React frontend
- JWT authentication
- Docker image builds
- CI/CD pipeline
- AWS deployment
- Kubernetes orchestration