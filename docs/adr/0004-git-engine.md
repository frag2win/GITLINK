# ADR 0004: Decoupled Git Engine in Rust

## Status
Accepted

## Context
Standard Git platforms typically embed `libgit2` directly into their primary web server (e.g., using Go's `git2go` bindings). However, cgo bindings introduce severe memory management complexities, panics, memory leaks, and threading models that conflict with Go's scheduler.

## Decision
We isolated all Git operations in a standalone Rust binary (`git-server`) leveraging the native `git2` Rust crate. The Go API server communicates with the Rust engine purely over a socket via protobuf IPC.

## Consequences
- **Positive**: Guaranteed memory safety for Git operations, zero cgo overhead in the main API server, and independent scaling and crash resilience of the Git engine.
- **Negative**: Increases deployment complexity (requires orchestrating two binaries) and introduces IPC latency overhead.
