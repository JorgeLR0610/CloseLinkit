# ADR 0005: Use Fixed-Length Short Codes

## Status
Accepted

## Context
A primary value proposition of a URL shortener is producing the shortest possible link to replace long, complex URLs.
The short code representation dictates both the link length and the total addressable namespace.
To represent these codes, we use a Base62 alphabet (consisting of 62 alphanumeric characters: `0-9`, `a-z`, `A-Z`).
We need to select a code length. Selecting a code length is a trade-off:
* **Short codes (e.g., 4–5 characters):** Produce very short URLs but have a limited namespace (e.g., $62^5 \approx 916$ million combinations). This increases the probability of hash collisions, forcing the application to perform multiple database checks and retries during creation.
* **Long codes (e.g., 8–10 characters):** Offer practically infinite namespaces but yield longer URLs, defeating the purpose of URL shortening.

We need a code length that is standard, preserves brevity, and guarantees an extremely low collision rate.

## Decision
We will use fixed-length short codes of exactly 7 characters generated using the Base62 character set.
* In the database schema, this is enforced by defining the column as `short_code VARCHAR(7) UNIQUE NOT NULL`.
* Using 7 characters provides $62^7 = 3,521,614,606,208$ (approx. 3.52 trillion) unique combinations.

## Consequences
* **Positive:**
  * **High Capacity:** A namespace of 3.5 trillion unique combinations is more than sufficient for the lifetime of the application, even under high write volumes.
  * **Extremely Low Collision Probability:** Randomly generated or hashed 7-character codes have a very low probability of collision, minimizing retry cycles during insertion.
  * **Aesthetics and Brevity:** Matches industry-standard shortener links (e.g. `myshortener.com/a8X2kLp`), fitting easily in SMS, social media posts, or print materials.
  * **Database Optimization:** A fixed maximum length of 7 characters allows PostgreSQL to optimize storage layout and index performance for the unique constraint.
* **Negative:**
  * **Collision Handling Still Required:** Even though the collision probability is low, the application must include retry logic when inserting a generated short code, in the rare event that a duplicate is generated.
  * **Rigid Length:** If we ever need to support dynamic or user-custom lengths shorter than 7 characters, we will need to change the application validation logic and handle a variable-length database column.
