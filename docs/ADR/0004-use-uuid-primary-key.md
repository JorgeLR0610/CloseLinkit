# ADR 0004: Use UUID Primary Key

## Status
Accepted

## Context
When designing database schemas, selecting the type of primary key is a foundational decision. Traditional options include:
1. **Auto-incrementing Integers (e.g., BIGINT/BIGSERIAL):** Simple, fast, and highly efficient for indexing. However, they expose the system to sequential enumeration attacks (where attackers guess or scrape resources by incrementing IDs) and reveal business metrics (such as the total number of shortened links). They also create challenges when merging databases or syncing data across distributed nodes.
2. **Universally Unique Identifiers (UUIDs):** 128-bit values that are globally unique. They do not leak sequence information and can be safely generated offline or in distributed nodes without database coordination.

For `CloseLinkit`, we need to ensure that the internal identifiers for database records (such as the `urls` table entries) are secure, non-guessable, and scale-friendly.

## Decision
We will use UUIDs as the primary keys for database tables, starting with the `urls` table.
* The database schema defines the primary key as `id UUID PRIMARY KEY DEFAULT gen_random_uuid()`, generating random UUIDs (UUID v4) natively within PostgreSQL.
* The Go backend represents and parses these IDs using the `github.com/google/uuid` library.

## Consequences
* **Positive:**
  * **Security & Privacy:** Prevents ID enumeration attacks. Clients cannot guess valid resource IDs by incrementing numbers. It also conceals business volume (total link count) from external observers.
  * **Distributed Generation:** Allows different services or client applications to generate unique IDs independently without querying the database for the next sequence value.
  * **Consistency:** Establishes a uniform primary key type across all future tables (e.g., users, logs, statistics).
* **Negative:**
  * **Storage Overhead:** UUIDs occupy 16 bytes compared to 8 bytes for 64-bit bigint columns, which increases both table and index sizes.
  * **Write Performance (Index Fragmentation):** Because UUID v4 values are random, inserts do not occur sequentially. This can cause index leaf fragmentation and frequent page splits in highly active B-Tree indexes as the dataset grows beyond memory capacity.
  * **Readability:** Long hex strings (e.g., `550e8400-e29b-41d4-a716-446655440000`) are less readable than integers during debugging and database administration.
