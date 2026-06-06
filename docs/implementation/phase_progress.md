# Phase-Wise Implementation Progress

This document tracks the actual implementation progress against the PRD's phase definitions.

## Phase 1: Local MVP & Core Architecture
**Goal:** Two people on the same local network can use this instead of GitHub via standard Git commands and a local React UI.

### Step 1: The Contract Layer (Implemented)
- **Status:** Done
- **Files:** `proto/git_commands.proto`
- **Details:** Defined the Protobuf schemas for all Git operations, acting as the strict communication contract between the Go API server and the Rust Git server.

### Step 2: The Rust Git Storage Vault (In Progress)
- **Status:** Partially Implemented
- **Files:** `services/git-server/src/git/*`
- **Details:** 
  - Raw `git2-rs` operations exist for repository lifecycle (`repository.rs`), commits (`commits.rs`), and objects (`objects.rs`).
  - Implemented `sanitize.rs` for path traversal protection and validation rules. Added missing `InvalidRepoName` error variant to `error.rs`.
  - *Pending:* The Unix Domain Socket server (`socket/server.rs`) and the handlers to execute these Git operations over the socket.

### Step 3: The Go API Gateway & Database (Scaffolded)
- **Status:** Pending
- **Files:** `services/api-server/*`
- **Details:** The directory structure and boilerplate (26-byte stub files) exist, but no actual HTTP endpoints, Unix socket client, or SQLite schemas (`001_initial.sql`) are implemented yet.

### Step 4: The React Frontend Foundation (Scaffolded)
- **Status:** Pending
- **Files:** `ui/src/*`
- **Details:** Base Vite/React configuration exists (`App.tsx`, `main.tsx`), but the specific views, API clients, and components are pending.

---

## Phase 2: Cross-Network P2P (libp2p)
- **Status:** Scaffolded (Pending)
- **Details:** Only the directory structure under `services/libp2p-node/` exists.

## Phase 3: Offline Queue & Resilience
- **Status:** Pending

## Phase 4 & 5: Self-Hostable Collaboration Platform
- **Status:** Pending
