# Phase Implementation PRD

This document breaks down the high-level phases from the main PRD into concrete, file-by-file implementation tasks based on our established 184-file structure.

---

## Phase 1: Local MVP & Core Architecture
**Goal:** Two people on the same local network can use this instead of GitHub via standard Git commands and a local React UI.

### Step 1: The Contract Layer
Before writing backend logic, we must define the Unix Domain Socket communication contract between Go and Rust.
- `proto/git_commands.proto`: Define Protobuf schemas for Git operations (Clone, Push, Pull, ListCommits, RepoCRUD).
- `scripts/generate-proto.sh`: Generate the Go and Rust bindings.

### Step 2: The Rust Git Storage Vault
Implement the raw `git2-rs` operations securely.
- `services/git-server/src/git/sanitize.rs`: Implement path traversal protection and name validation.
- `services/git-server/src/git/repository.rs`: Bare repo initialization, opening, and deletion logic.
- `services/git-server/src/git/commits.rs` & `objects.rs`: Commit history walking and file tree browsing.
- `services/git-server/src/socket/server.rs`: The Tokio async Unix socket server that accepts connections from Go.

### Step 3: The Go API Gateway & Database
Bridge the Rust vault to the HTTP world and set up persistent state.
- `services/api-server/internal/database/migrations/001_initial.sql`: Set up SQLite tables for repos and contributors.
- `services/api-server/internal/socket/git_client.go`: Write the Unix socket client to talk to the Rust server.
- `services/api-server/internal/handlers/repos.go` & `commits.go`: Build the REST API endpoints.
- `services/api-server/internal/middleware/auth.go`: Implement SSH key authentication.

### Step 4: The React Frontend Foundation
Connect the UI to the API to display the local repos.
- `ui/src/api/client.ts`: The typed fetch wrapper for the Go REST API.
- `ui/src/pages/Dashboard.tsx` & `RepoList.tsx`: Build the repo listing views.
- `ui/src/components/git/FileBrowser.tsx`: The recursive file tree component.

---

## Phase 2: Cross-Network P2P (libp2p)
**Goal:** Contributors across the internet connect directly without central servers.

### Step 1: Identity & Networking Core
Establish cryptographic identities and pure P2P transports.
- `services/libp2p-node/internal/identity/keypair.go`: Generate/load the persistent Ed25519 node identity.
- `services/libp2p-node/internal/host/host.go`: Configure `go-libp2p` with TCP, Noise encryption, and Yamux.

### Step 2: Discovery & Hole Punching
Find peers across NAT boundaries.
- `services/libp2p-node/internal/discovery/mdns.go`: Local network discovery.
- `services/libp2p-node/internal/discovery/dht.go`: Internet-wide Kademlia DHT routing.
- `services/libp2p-node/internal/nat/dcutr.go`: Implement DCUtR for simultaneous NAT hole punching.

### Step 3: P2P Streams & The Bridge
Forward Git traffic from libp2p securely into the API Server.
- `services/libp2p-node/internal/protocol/git_protocol.go`: Handle incoming Git binary streams over libp2p.
- `services/libp2p-node/internal/bridge/socket.go`: Forward streams via Unix socket to the Go `api-server`.

---

## Phase 3: Offline Queue & Resilience
**Goal:** Uninterrupted workflow even when the host peer is offline.

### Step 1: The Offline Queue
- `services/libp2p-node/internal/bridge/queue.go`: Store pushed commits locally on the contributor device when the target peer is unreachable.
- `ui/src/store/connectionStore.ts`: Update React state to show "Offline / Queued" badge.

### Step 2: Auto-Sync
- `services/libp2p-node/internal/protocol/queue_protocol.go`: Background worker that flushes queued pack files the moment `go-libp2p` detects the host peer is back online.

---

## Phase 4 & 5: Self-Hostable Collaboration Platform
**Goal:** Issue tracking, PRs, and scalable architecture (Gitea parity).

- Database migration from SQLite to PostgreSQL (`services/api-server/internal/database/db.go`).
- Implement Issues, Pull Requests, and Code Review UI components in the React app.
- Build federation protocols so two separate GITLINK instances can share repos.
