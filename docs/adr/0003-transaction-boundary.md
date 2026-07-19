# ADR 0003: Transaction Boundaries and Compensation

## Status
Accepted

## Context
Operations like `CreateRepository` involve two heterogeneous systems: the PostgreSQL database (for metadata and ownership) and the Git Engine (for filesystem bare repos). Because these systems do not share a two-phase commit (2PC) coordinator, a standard database transaction cannot roll back filesystem changes in the event of failure.

## Decision
We adopted an atomic database boundary with a manual compensation mechanism for external systems:
1. Operations affecting the Git filesystem are executed **first**.
2. If successful, the database transaction begins, recording repository metadata and audit logs atomically.
3. If the database transaction fails and rolls back, a manual compensation action (`compensateFailedCreate`) explicitly deletes the newly created Git repository.

## Consequences
- **Positive**: Prevents orphaned database records that point to non-existent Git repositories. Database changes remain strictly atomic.
- **Negative**: There remains a tiny window for orphaned Git repositories if the application hard-crashes between Git creation and database commit (requiring async reconciliation/cleanup later).
