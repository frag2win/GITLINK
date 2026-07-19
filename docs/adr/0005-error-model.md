# ADR 0005: Unified Error Model for IPC

## Status
Accepted

## Context
When defining the Protobuf RPC interface, each operation initially defined its own explicit error structure (e.g., `CreateRepoError`, `DeleteRepoError`). This led to massive duplication and tedious error mapping logic in the Go API server, forcing handlers to type-assert or inspect complex nested structures for every single endpoint.

## Decision
We adopted a unified `GitError` model to be returned by all endpoints.
```protobuf
message GitError {
  string code = 1;
  string message = 2;
}
```
All protobuf response payloads contain a `GitError` field alongside an optional result payload. In the Go API layer, these error responses are deserialized into a native `ipc.GitError` type that implements the standard `error` interface.

## Consequences
- **Positive**: Massively simplifies the Go error handling boilerplate and guarantees a consistent error structure across the IPC boundary.
- **Negative**: Certain complex operational errors (like `MergeConflictList`) cannot fit into the basic `GitError` and require secondary explicit structures on specific endpoints.
