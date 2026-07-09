# ADR 0003: Use sqlc

## Status
Accepted

## Context
When interacting with PostgreSQL in Go, developers generally choose between three patterns:
1. **Raw `database/sql` or `pgx`:** Highly performant, but requires writing tedious boilerplate code to map query parameters, execute statements, scan rows into structs, and handle null values.
2. **Object-Relational Mappers (ORMs) like GORM:** Reduce boilerplate by abstracting SQL, but introduce performance overhead due to reflection, make query optimization harder, and hide the actual SQL execution.
3. **Compile-time Code Generators:** Generate type-safe database code directly from raw SQL statements and schemas.

To maintain high performance and database query clarity while eliminating repetitive boilerplate and scan errors, we need a solution that keeps SQL first-class while remaining type-safe.

## Decision
We will use `sqlc` to generate clean, type-safe Go code from raw SQL schemas and queries. Schema migrations will be managed using Goose format (`db/migrations`), and sqlc will parse these migrations and the query files (`db/queries`) to generate repository handlers under `internal/repository`.

## Consequences
* **Positive:**
  * **Compile-Time Type Safety:** Any mismatch between query parameters or table column types and Go variables is caught at compile time rather than runtime.
  * **No Boilerplate:** Automatically generates Go structs and database method wrappers, eliminating manual `rows.Scan()` and parameter mapping.
  * **SQL-First Development:** Developers write raw, optimized SQL queries, which makes explaining, debugging, and profiling queries straightforward.
  * **Zero ORM Overhead:** Generated code uses standard Go database drivers directly without reflection or runtime query-building overhead.
* **Negative:**
  * **Code Generation Step:** Developers must run `sqlc generate` whenever the database schema or query files change.
  * **Dynamic Query Limits:** Building highly dynamic SQL queries (e.g., dynamically adding multiple optional search filters or complex sorting clauses based on user inputs) is difficult in `sqlc` and may occasionally require falling back to manual sql builder libraries or raw `database/sql` code.
