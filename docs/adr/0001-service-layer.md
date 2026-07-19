# ADR 0001: Service Layer Architecture

## Status
Accepted

## Context
As GITLINK expanded, the HTTP handlers accumulated business logic, database queries, and direct IPC calls to the Git engine. This caused tight coupling, making endpoints hard to test and hindering our ability to introduce features like authentication, auditing, and complex transactions.

## Decision
We introduced a strict Service Layer architectural pattern between the HTTP Handlers and the underlying infrastructure (Database and Git/IPC):

1. **HTTP Handlers**: Restricted to parsing requests, validating inputs, and returning responses. They contain zero business logic.
2. **Service Layer**: Contains domain-specific business rules.
   - `GitService`: Orchestrates all interactions with the Rust `git-server` via IPC.
   - `RepoService`: Manages repository lifecycle, coordinating both `GitService` and database states.
3. **Repository Layer**: Handles persistence (PostgreSQL) and exposes data access objects.

## Consequences
- **Positive**: High testability, decoupled layers, and clear boundaries of responsibility. Future protocols (like gRPC) can reuse the Service Layer.
- **Negative**: Adds boilerplate (interfaces, constructors) to what were previously simple functions.
