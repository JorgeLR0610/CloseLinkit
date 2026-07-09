# ADR 0002: Use PostgreSQL

## Status
Accepted

## Context
The URL shortening service `CloseLinkit` must store persistent associations between short codes and original URLs.
To support functionalities like expiration times, click counts, and future features (e.g., user authentication, custom aliases, regional redirect logic, and analytics), we need a database that provides robust transaction guarantees (ACID), handles concurrent reads/writes efficiently, and allows flexible relational modeling.
We evaluated options including NoSQL databases (e.g., MongoDB, Redis) and relational databases (e.g., SQLite, PostgreSQL). While Redis is excellent for caching, NoSQL systems lack the strict transactional guarantees and rich querying capabilities needed for core analytical relationships in future phases. SQLite is serverless and easy to run but lacks support for high concurrent write volumes (due to file-level locking) and distributed setups.

## Decision
We will use PostgreSQL as the primary relational database to persist all core application data. For local development, PostgreSQL will run in a containerized environment managed via Docker Compose (`compose.yaml`).

## Consequences
* **Positive:**
  * **Consistency & Transactions:** Strong ACID compliance ensures correct transaction execution, which is crucial for atomic metric updates (such as click counts).
  * **Read Performance:** Supports rich indexing types (e.g., B-Tree indexes on the `short_code` field) to ensure sub-millisecond lookup latency during redirection flows.
  * **Ecosystem Support:** Excellent integration with Go packages (such as `pgx`) and query generators like `sqlc`.
  * **Extensibility:** The relational schema can easily accommodate future extensions like user management, roles, and complex analytics.
* **Negative:**
  * **Operational Overhead:** Requires running and configuring a database server process, adding slight complexity to the infrastructure compared to SQLite.
  * **Memory Footprint:** PostgreSQL has higher baseline memory and resource utilization than lightweight alternatives.
