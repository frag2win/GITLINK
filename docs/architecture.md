# Architecture — P2P Git Hosting Platform

## Overview

The platform follows a **three-container, local-first** architecture. Each container has a single responsibility and a deliberately minimal network surface. Inter-process communication (IPC) happens exclusively through **Unix domain sockets**, eliminating TCP-based attack vectors between services.

---

## Container Architecture

### 1. libp2p-node (Go)

| Property        | Value                          |
|-----------------|--------------------------------|
| Language        | Go 1.22+                      |
| Network mode    | `host`                         |
| IPC endpoints   | `/run/p2p/libp2p.sock` (serve) |
|                 | `/run/git/git.sock` (client)   |

**Responsibilities:**

- **Peer discovery** — Uses mDNS for LAN discovery and the libp2p Kademlia DHT for wider-area discovery.
- **Transport** — Listens for inbound libp2p connections using the Noise protocol for authentication and encryption.
- **Protocol multiplexing** — Handles the custom `/p2p-git/pack/1.0.0` and `/p2p-git/refs/1.0.0` protocols.
- **Data relay** — Receives git pack-file requests from remote peers and proxies them to the local git-server via its Unix socket.

The node runs in **host network mode** so it can:
- Bind directly to a host port for inbound peer connections.
- Participate in mDNS multicast on the LAN interface.
- Avoid Docker's NAT, which complicates libp2p's address advertisement.

### 2. git-server (Rust)

| Property        | Value                          |
|-----------------|--------------------------------|
| Language        | Rust 1.77+                     |
| Network mode    | `none`                         |
| IPC endpoints   | `/run/git/git.sock` (serve)    |

**Responsibilities:**

- **Repository management** — Creates, lists, and deletes bare Git repositories under `/repos`.
- **Pack-file operations** — Implements `upload-pack` (clone/fetch) and `receive-pack` (push) over the Unix socket.
- **Ref advertisement** — Serves repository reference lists (branches, tags) to the API server and libp2p node.
- **Object validation** — Verifies incoming pack-files before writing to disk.

The git-server is the **most security-critical** component. By running with `network: none`, it is physically impossible for it to make or accept TCP connections — even if an attacker achieves code execution inside the container.

### 3. api-server (Go + Fiber)

| Property        | Value                          |
|-----------------|--------------------------------|
| Language        | Go 1.22+ with Fiber framework |
| Network mode    | `bridge` (internal)            |
| IPC endpoints   | `/run/p2p/libp2p.sock` (client)|
|                 | `/run/git/git.sock` (client)   |

**Responsibilities:**

- **REST API** — Serves endpoints for repository CRUD, peer listing, and settings management.
- **WebSocket** — Provides real-time updates for clone/push progress and peer events.
- **Authentication** — Manages local authentication (no external auth providers needed).
- **Database** — Stores metadata (repo list, peer nicknames, ACLs) in a SQLite database at `/data/api.db`.

The API server is the only container with a Docker-mapped port (`3000`). It acts as the gateway between the web UI and the two backend services.

---

## IPC via Unix Domain Sockets

```
                    ┌────────────────────┐
                    │    Web UI / CLI     │
                    │   (localhost:3000)  │
                    └────────┬───────────┘
                             │ HTTP / WebSocket
                             ▼
┌──────────────┐  /run/p2p/  ┌──────────────────┐  /run/git/  ┌──────────────┐
│  libp2p-node │◄───────────►│    api-server     │◄───────────►│  git-server   │
│              │ libp2p.sock │                    │  git.sock   │              │
└──────┬───────┘             └────────────────────┘             └──────────────┘
       │                                                              ▲
       │                         /run/git/git.sock                    │
       └──────────────────────────────────────────────────────────────┘
```

**Why Unix sockets instead of TCP?**

1. **No network exposure** — Unix sockets are filesystem objects. They cannot be reached from outside the host, and containers with `network: none` can still use them.
2. **Lower latency** — No TCP handshake, no Nagle algorithm, no loopback overhead.
3. **Filesystem permissions** — Access can be restricted via standard Unix file permissions (owner, group, mode).
4. **No port conflicts** — No need to manage port allocations between services.

### Socket Protocol

Both sockets use a simple **length-prefixed JSON-RPC** protocol:

```
┌──────────┬───────────────────────────────────────┐
│ 4 bytes  │  JSON-RPC 2.0 message (UTF-8)         │
│ (length) │                                        │
└──────────┴───────────────────────────────────────┘
```

For binary data (pack-files), the response includes a `Content-Type: application/x-git-pack` header and the payload is streamed in 64 KiB chunks after the initial JSON-RPC response.

---

## Data Flow Examples

### Local Clone (Web UI)

1. User clicks "Clone" in the web UI.
2. `api-server` receives the HTTP request.
3. `api-server` sends a `refs/list` RPC to `git-server` via `/run/git/git.sock`.
4. `git-server` reads the bare repo and returns ref advertisement.
5. `api-server` sends an `upload-pack` RPC to `git-server`.
6. `git-server` generates a pack-file and streams it back.
7. `api-server` proxies the pack-file to the web UI via HTTP.

### Remote Clone (Peer-to-Peer)

1. Remote peer connects to `libp2p-node` via the Noise-encrypted transport.
2. Remote peer opens the `/p2p-git/refs/1.0.0` protocol stream.
3. `libp2p-node` proxies the request to `git-server` via `/run/git/git.sock`.
4. `git-server` returns the ref advertisement.
5. Remote peer opens `/p2p-git/pack/1.0.0` and sends a want/have negotiation.
6. `libp2p-node` proxies the negotiation to `git-server`.
7. `git-server` generates and streams the pack-file.
8. `libp2p-node` relays the pack-file to the remote peer over the encrypted stream.

---

## Volume Layout

| Volume          | Mount Path     | Purpose                                    |
|-----------------|----------------|--------------------------------------------|
| `peer-identity` | `/identity`    | Ed25519 peer key (persistent across restarts) |
| `git-socket`    | `/run/git`     | Unix socket for git-server IPC              |
| `p2p-socket`    | `/run/p2p`     | Unix socket for libp2p-node IPC             |
| Host bind mount | `/repos`       | Bare git repositories (user data)           |
| Host bind mount | `/data`        | SQLite database (API metadata)              |

---

## Technology Choices

| Component       | Choice          | Rationale                                            |
|-----------------|----------------|------------------------------------------------------|
| Networking      | libp2p (Go)    | Mature p2p stack with NAT traversal, mDNS, Noise     |
| Git backend     | Rust + git2    | Memory-safe, no GC pauses during pack-file generation|
| API framework   | Go + Fiber     | High-performance HTTP with familiar Express-like API |
| Database        | SQLite         | Zero-config, single-file, perfect for local-first    |
| Web UI          | React + Vite   | Fast dev experience, rich ecosystem                  |
| Transport auth  | Noise protocol | Mutual authentication, forward secrecy               |
