# ADR 0007: Use Monorepo

## Status
Accepted

## Context
The `CloseLinkit` project is designed as a multi-component system. Under version 0.1, the application contains a Go backend server and database migrations.
Future phases will introduce:
1. A frontend user interface built with React and TypeScript.
2. Infrastructure-as-code configuration, Kubernetes manifests, or Docker configuration files.
3. Auxiliary scripts or tools for administration and testing.

Managing these components across multiple, separate Git repositories increases administrative overhead (e.g., managing permissions, issues, and pull requests across multiple places). It also complicates feature development, since adding a single feature might require coordinated commits across three different repositories. This can easily lead to synchronization problems and build breakages.

## Decision
We will use a monorepo structure. All components, configuration files, schemas, frontends, backends, and documentation for the entire project will reside in a single, unified Git repository.
* The Go backend is located in the `/server` directory.
* Project-wide documentation is located in `/docs`.
* Any future frontend application (e.g., React client) will be placed in a dedicated top-level directory (e.g., `/web` or `/frontend`) in the same repository.
* Shared environment files, Docker Compose configs, and tooling scripts live at the repository root.

## Consequences
* **Positive:**
  * **Atomic Commits:** A single feature or fix spanning the database schema, backend, frontend, and deployment configurations can be committed and reviewed in a single, atomic Pull Request.
  * **Unified Developer Onboarding:** A new developer can clone a single repository and run a single command (e.g., `docker compose up`) to run the entire stack locally.
  * **Shared Documentation:** Documentation (such as architecture designs and these ADRs) lives alongside the code, ensuring it remains visible and up-to-date.
  * **Simplified Dependency Management:** Code sharing and common environment configs are easier to manage when files are co-located.
* **Negative:**
  * **Repository Size:** The repository size will grow faster than a single-purpose repository.
  * **CI/CD Configuration:** Build and deployment pipelines must be configured carefully (using path filters) to ensure that changes in one subdirectory do not trigger unnecessary builds or deployments of unrelated components.
