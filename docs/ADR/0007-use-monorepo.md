# ADR 0007: Use Monorepo

## Status
Accepted

## Context
The `CloseLinkit` project is designed as a multi-component system. Under version 0.1, the application contains a Go backend server, a React frontend client, database migrations and Docker configurations.
Future phases may introduce:
1. Auxiliary scripts or tools for administration and testing.
2. CI/CD automation
3. Infrastructure-as-code configuration
4. Kubernetes manifests

Managing these components across multiple, separate Git repositories increases administrative overhead (e.g., managing permissions, issues, and pull requests across multiple places). It also complicates feature development, since adding a single feature might require coordinated commits across three different repositories. This can easily lead to synchronization problems and build breakages.

## Decision
We will use a monorepo structure. All components, configuration files, schemas, frontends, backends, and documentation for the entire project will reside in a single, unified Git repository.
* The Go backend is located in the `/server` directory.
* The React client is located in the `/frontend` directory
* Project-wide documentation is located in `/docs`.
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
