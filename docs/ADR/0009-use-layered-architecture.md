# ADR 0009: Use Layered Architecture

## Status
Accepted

## Context
As the complexity of the backend API grows, it becomes critical to maintain a clean codebase where the HTTP transport logic, business rules, and database access are decoupled. Mixing these concerns in single functions makes the code hard to read, maintain, and test.

*Note: The backend was designed following a layered architecture from the beginning of the project. This ADR documents the rationale behind that decision.*

## Decision
Adopt a layered architecture (also known as N-Tier architecture) for the backend. The system is divided into distinct layers, primarily:
1. **Handler/Transport Layer:** Responsible for parsing HTTP requests and formatting responses.
2. **Service Layer:** Contains the core business logic.
3. **Repository/Data Access Layer:** Handles interactions with the PostgreSQL database.

## Consequences
* **Positive:**
  * **Separation of Concerns:** Each layer has a distinct responsibility, making the code much easier to understand and maintain.
  * **Low Coupling:** The layers can evolve independently. For example, changing the database or the HTTP framework does not require rewriting the core business logic.
  * **Testability:** It greatly facilitates testing. Business logic can be tested in isolation by mocking the repository layer, and handlers can be tested by mocking the service layer.
* **Negative:**
  * **Added Complexity:** It requires writing more boilerplate code (e.g., interfaces and dependency injection) compared to a simple, flat structure.
  * **Data Transformation:** Data often needs to be mapped or translated as it passes between layers (e.g., from database rows to domain models to response DTOs).
