# ADR 0002: IPC Transport Abstraction

## Status
Accepted

## Context
The Go API server communicates with the Rust Git Engine over IPC. Initially, Unix Domain Sockets were hardcoded. This broke functionality on Windows, forcing the API server and Git Engine to run differently based on OS.

## Decision
We abstracted the IPC layer behind a uniform `Transport` interface capable of supporting multiple underlying mechanisms:
1. `UnixSocketTransport` for Linux/macOS.
2. `TcpLoopbackTransport` (127.0.0.1) as a reliable cross-platform fallback (especially for Windows).

The rest of the system (`GitClient`) operates on the abstract interface and is entirely agnostic to the transport layer.

## Consequences
- **Positive**: Cross-platform compatibility without littered `runtime.GOOS` checks. Allows for future transports like Windows Named Pipes.
- **Negative**: Slight overhead for TCP loopback compared to raw named pipes on Windows.
