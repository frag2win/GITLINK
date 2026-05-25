# [PROJECT NAME] — Product Requirements Document

> **Local-First, Privacy-Preserving, Peer-to-Peer Git Platform**

| Field | Value |
|-------|-------|
| Version | 1.2 |
| Status | Draft |
| Date | May 2026 |
| Author | Founder |
| Phase Coverage | All Phases (1–5) |
| License | Open Source |

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [Problem Statement](#2-problem-statement)
3. [Vision & Principles](#3-vision--principles)
4. [Target Users](#4-target-users)
5. [Technology Stack](#5-technology-stack)
6. [System Architecture](#6-system-architecture)
7. [Docker Networking Architecture](#7-docker-networking-architecture)
8. [Security Architecture](#8-security-architecture)
9. [Feature Roadmap by Phase](#9-feature-roadmap-by-phase)
10. [Offline & Resilience Behavior](#10-offline--resilience-behavior)
11. [Non-Functional Requirements](#11-non-functional-requirements)
12. [Open Source Dependencies](#12-open-source-dependencies)
13. [Out of Scope](#13-out-of-scope)
14. [Success Metrics](#14-success-metrics)
15. [Risks & Mitigations](#15-risks--mitigations)
16. [Appendix — Key Concepts](#16-appendix--key-concepts)

---

## 1. Executive Summary

> **A privacy-first, local-first Git hosting platform that runs entirely on your own hardware. Contributors connect directly peer-to-peer using libp2p — no GitHub, no cloud, no third party ever sees your code.**

This project is born from a single conviction: your code belongs to you. Not to a corporation, not to a cloud provider, not to any third party. Every line of code you write, every commit, every branch — all of it should live on hardware you physically control.

The platform solves a problem no existing tool cleanly addresses: how do you give contributors seamless access to your local Git server when they are on a completely different network, without routing your code through any external server?

The answer is **libp2p** — the same peer-to-peer networking stack that powers IPFS and the Ethereum network. Contributors anywhere in the world connect directly to your machine, through NAT, through firewalls, with end-to-end encryption — and your code never leaves your hardware.

The backend is built on a hybrid **Go + Rust** architecture. Go owns the network edge — handling thousands of concurrent P2P connections with goroutine-level concurrency. Rust owns the storage vault — parsing raw Git binary data with compile-time memory safety guarantees that make entire classes of exploits impossible to ship.

---

## 2. Problem Statement

### 2.1 The Core Problem

Every popular Git hosting solution today requires your code to live on someone else's server:

| Platform | Who Owns Your Data |
|----------|--------------------|
| GitHub | Microsoft |
| GitLab.com | GitLab Inc. |
| Bitbucket | Atlassian |
| Gitea Cloud | Gitea Ltd. |

Even self-hosted solutions like Gitea or GitLab require a dedicated server — meaning either you pay for a VPS (third-party infrastructure) or you run a machine that needs to be publicly accessible on the internet.

### 2.2 Specific Pain Points

- **Code privacy** — no control over who can access your repository data at the infrastructure level
- **Cost** — paid hosting, VPS fees, storage costs
- **Connectivity dependency** — contributors cannot work if the central server is down
- **Network requirement** — no clean solution for local-first hosting with cross-network P2P access
- **Data sovereignty** — subject to terms of service that may claim rights over hosted content
- **Censorship** — repositories can be taken down without recourse

---

## 3. Vision & Principles

### 3.1 Core Vision

> **"Your machine is the server. Always. Cloud is optional backup — never the source of truth."**

### 3.2 Non-Negotiable Principles

| Principle | What It Means in Practice |
|-----------|--------------------------|
| **Local First** | All repos live on your hardware. No exceptions. |
| **Zero Third Parties** | No paid services, no external coordinators, no cloud dependencies. |
| **Open Source Only** | Every dependency must be open source. Full stack auditable. |
| **Privacy by Default** | Data never leaves your hardware unless you explicitly mirror it. |
| **Least Privilege** | Every component runs with minimum required permissions. |
| **No Money Motive** | Built for privacy and code protection. Not a commercial product. |

---

## 4. Target Users

### 4.1 Primary User

> Solo developer who wants complete ownership and privacy of their codebase.

### 4.2 Extended Users

| User Type | Use Case |
|-----------|----------|
| Solo Developer | Personal projects, full privacy, runs on laptop or Raspberry Pi |
| Small Team (2–10) | Startup or agency, shared local server on office hardware |
| Privacy-conscious org | Regulated industries, air-gapped requirements, no cloud policy |
| Open source maintainer | Self-sovereign hosting, no dependence on platform goodwill |

---

## 5. Technology Stack

### 5.1 The Hybrid Core

To satisfy the strict networking requirements of peer-to-peer NAT traversal while maintaining absolute cryptographic and memory safety at the filesystem level, the backend architecture splits responsibilities across two highly specialized systems languages.

Go owns the **network edge** — everything that touches the internet, routes requests, and manages concurrent peer connections. Rust owns the **storage vault** — everything that touches raw `.git` binary data on disk. These two domains have fundamentally different threat profiles, and each language is the best tool in existence for its specific domain.

| Layer | Technology | Why | License |
|-------|-----------|-----|---------|
| **Network Edge (P2P)** | Go + `go-libp2p` | Reference libp2p implementation. Goroutines handle thousands of concurrent DHT lookups with sub-millisecond scheduling jitter — critical for DCUtR hole punching timing. | Apache 2.0 |
| **Storage Vault (Git)** | Rust + `git2-rs` | Compile-time memory safety via borrow checker. Mathematical immunity from buffer overflows and use-after-free when parsing raw `.git` binary pack files. Zero GC overhead. | MIT |
| **API Gateway** | Go + Fiber | Hyper-fast HTTP routing. Serves static React frontend directly — eliminates a container. Bridges browser ↔ git-server via Unix Domain Socket. | MIT |
| **Frontend** | React + Vite (static) | Pure static build. No Node.js runtime in production. Go serves compiled output directly. Fast HMR in development. | MIT |
| **Database** | SQLite → PostgreSQL | Zero-config embedded state for local metadata. `modernc.org/sqlite` in Go (pure Go, no cgo). Swappable to PostgreSQL in Phase 3. | MIT / PostgreSQL |
| **Containerization** | Docker + Docker Compose | Strict process isolation. Surgical network jailing per container role. | Apache 2.0 |
| **Encryption** | Noise Protocol (via go-libp2p) | `Noise_XX_25519_ChaChaPoly_SHA256`. Unified identity and transport encryption. Forward secrecy per session. | MIT |

### 5.2 Why Go for Networking — The Technical Case

DCUtR hole punching requires both peers to send packets **simultaneously within milliseconds**. This is where Node.js fails and Go excels:

```
Node.js event loop under load:
- Single-threaded, callbacks queued
- 1000 concurrent DHT lookups = 1000 queued callbacks
- Event loop jitter under load: 10–50ms
- One slow callback blocks all others
- Hole punch timing becomes unreliable

Go goroutine scheduler:
- M:N threading — goroutines mapped to real OS threads
- 1000 concurrent DHT lookups = 1000 independent goroutines
- Scheduler jitter: sub-millisecond
- One slow lookup never blocks others
- Hole punch timing is deterministic
```

`go-libp2p` is also the **reference implementation** — every major protocol decision in libp2p is validated against Go first. It is the same stack that runs IPFS in production globally.

### 5.3 Why Rust for Git Storage — The Technical Case

The git-server parses raw binary `.git` format files — pack files, object headers, delta chains. This is one of the most dangerous parsing surfaces in systems programming. Rust eliminates entire exploit classes at compile time:

```
Buffer overflow (crafted pack file with malformed length header):
- C: reads past buffer boundary → arbitrary code execution
- Rust: compiler rejects the code. Cannot ship.

Use-after-free (object reference after memory freed):
- C: undefined behavior, exploitable
- Rust: borrow checker catches at compile time. Cannot ship.

Path traversal (repo name: "../../etc/passwd"):
- C: requires manual bounds checking, easy to miss
- Rust: type system enforces sanitization patterns

Integer overflow (pack file size calculation):
- C: silent wraparound, heap corruption
- Rust: checked arithmetic by default, panics instead of corrupts
```

`git2-rs` provides safe Rust bindings over `libgit2` — the same C library powering GitHub, GitLab, and Gitea — with Rust's safety guarantees wrapped around every call.

### 5.4 Container Logic and Language Mapping

| Container | Language | Network Mode | Filesystem Access | Role |
|-----------|---------|-------------|------------------|------|
| `libp2p-node` | Go | `host` | Identity key only (read-only) | P2P networking, DHT, hole punching |
| `git-server` | Rust | `none` | `/repos` only (read-write) | Git operations, pack file parsing |
| `api-server` | Go | Internal bridge | `/repos` + `/db` + Unix socket | HTTP routing, auth, serves UI static files |

> The ui-server container is **eliminated**. The Go api-server serves the compiled React/Vite static build directly. One less container, one less attack surface, zero Node.js runtime in production.

### 5.5 libp2p Networking Components (Go Implementation)

| Component | Purpose |
|-----------|---------|
| TCP Transport | Core multiplexed data transport across networks |
| Kademlia DHT | Fully decentralized peer discovery across the internet |
| mDNS | Zero-config peer discovery on local and air-gapped networks |
| Noise Protocol | `Noise_XX_25519_ChaChaPoly_SHA256` — end-to-end encryption and mutual auth |
| DCUtR | Coordinates NAT hole punching with millisecond timing precision |
| AutoNAT | Detects NAT type, selects optimal connection strategy automatically |
| Circuit Relay v2 | Fallback relay pipeline for symmetric NAT environments |
| Yamux | High-performance stream multiplexing over a single P2P connection |

### 5.6 Inter-Process Communication — Unix Domain Sockets

The Go api-server communicates with the Rust git-server exclusively via Unix Domain Sockets (`.sock` files), never TCP:

```
TCP (even 127.0.0.1):          Unix Domain Socket:
- Full kernel network stack    - Kernel memory copy, direct
- IP header processing         - Zero network stack overhead
- TCP handshake overhead       - No IP processing
- ~15–20μs latency             - ~2–5μs latency
- Reachable via network        - Filesystem permission only
                               - Cannot be accessed remotely
```

For streaming large pack files during clone operations, the latency difference compounds over millions of small reads. The security property — unreachable from any network path — is stronger than any firewall rule.

---

## 6. System Architecture

### 6.1 High-Level Container Layout

```
Host Machine (Your Hardware — Arch Linux / Raspberry Pi / any Linux)
│
├── Container: libp2p-node  [Go]           [network_mode: host]
│   └── go-libp2p reference implementation
│   └── Goroutine per DHT lookup — true concurrency
│   └── Sub-millisecond DCUtR timing precision
│   └── NO filesystem access except identity key (read-only)
│   └── Communicates via Unix socket to api-server
│
├── Container: git-server   [Rust]         [network_mode: none]
│   └── git2-rs — safe bindings over libgit2
│   └── Borrow checker eliminates buffer overflows at compile time
│   └── Parses pack files, manages objects, handles refs
│   └── ZERO network interface — physically unreachable from network
│   └── Mounts /repos only — nothing else visible
│
├── Container: api-server   [Go + Fiber]   [bridge: internal]
│   └── Fiber HTTP router — serves REST API and React static build
│   └── Owns SQLite / PostgreSQL connection
│   └── Bridges libp2p-node ↔ git-server via Unix sockets
│   └── Exposed on 127.0.0.1:3000 only (localhost, never 0.0.0.0)
│
└── Volume: /repos
    └── The only place repository data ever lives
    └── Mounted ONLY in git-server and api-server
    └── Never visible to libp2p-node
```

### 6.2 Request Flow — Browser to Repo

```
User opens browser → localhost:3000
        ↓
api-server (Go/Fiber) serves React static files
        ↓
User action triggers API call → api-server
        ↓
api-server authenticates request (SSH key / peer identity)
        ↓
api-server writes request to Unix socket → git-server
        ↓
git-server (Rust) parses request, reads/writes /repos
        ↓
Response streams back via Unix socket → api-server → browser
```

### 6.3 Request Flow — Remote Contributor Push

```
Contributor runs: git push origin main
        ↓
Contributor's Git client → libp2p encrypted tunnel
        ↓
libp2p-node (Go) receives stream on host network
        ↓
Writes to Unix socket → api-server
        ↓
api-server validates contributor Peer ID + repo permissions
        ↓
Writes pack data to Unix socket → git-server
        ↓
git-server (Rust) safely parses and applies pack file to /repos
        ↓
Success response flows back through same path
```

### 6.4 Data Flow by Scenario

| Scenario | What Happens |
|----------|-------------|
| Same network push | Contributor → mDNS discovery → direct TCP → libp2p-node → Unix socket → git-server → /repos |
| Different network push | Contributor → DHT lookup → hole punch → direct P2P → libp2p-node → Unix socket → git-server → /repos |
| Host offline — contributor commits | Commits stored locally on contributor device, queue flagged pending |
| Host comes back online | libp2p detects host peer → auto-flush queue → sync complete |
| Backup mirror (optional) | Host pushes one-way mirror to self-hosted VPS — local always stays primary |

---

## 7. Docker Networking Architecture

> **This section addresses the most critical engineering challenge in the entire stack.**

### 7.1 The Problem — Docker's Hidden Double NAT

Standard Docker creates a bridge network with its own internal virtual router. This introduces a second NAT layer on top of your home or office router:

```
Internet
    ↓
Your Router              ← NAT Layer 1 (expected)
    ↓
Host Machine (Arch Linux)
    ↓
Docker Bridge            ← NAT Layer 2 (the problem)
    ↓
libp2p container         → sees IP: 172.17.0.x (fake Docker-internal IP)
```

**Why this completely breaks libp2p:**

When libp2p performs NAT hole punching, it reports its observed address to the relay peer. With Docker bridge networking, libp2p reports `172.17.0.x` — a Docker-internal IP meaningless to the outside world. DCUtR's millisecond-precision simultaneous punch gets thrown off by the extra translation layer, causing silent connection failures that are extremely hard to debug.

### 7.2 The Solution — Surgical `network_mode: host` for libp2p Only

Give **only the libp2p container** host networking. Every other container keeps strict isolation.

```
Internet
    ↓
Your Router              ← NAT Layer 1 (only layer now)
    ↓
Host Machine = libp2p-node container (same network namespace)
    ↓
libp2p-node sees real IP: 192.168.1.x  ✅
Hole punching works correctly           ✅
DHT announces real reachable address    ✅
AutoNAT reports accurate NAT type       ✅
```

### 7.3 The Trade-Off and Why It Is Acceptable

| Security Property | Bridge Network | Host Network |
|-------------------|---------------|-------------|
| Network namespace isolation | ✅ Full | ❌ Shared with host |
| Can sniff host network traffic | ❌ Cannot | ⚠️ Theoretically possible |
| Access to /repos | ❌ Cannot | ❌ Cannot (not mounted) |
| Access to database | ❌ Cannot | ❌ Cannot (not mounted) |
| Access to git-server | ❌ Cannot | ❌ Cannot (network_mode: none) |
| Access to api-server | ❌ Cannot | ❌ Cannot (internal network) |

The libp2p-node's only job is networking. Even fully compromised with host network access, an attacker cannot reach repos, database, or other containers. Worst case is network-level damage. The container is destroyed and rebuilt from the Go binary in seconds.

### 7.4 git-server Gets `network_mode: none`

The container that touches your repos has **zero network interface**:

```
Attacker fully compromises libp2p-node container
    ↓
Tries to reach git-server via TCP/UDP
    ↓
git-server has NO network interface whatsoever
    ↓
No TCP, no UDP, no ICMP — nothing
    ↓
Dead end ✅ — repos physically unreachable from network layer
```

The only path to repo data is through the api-server Unix socket, which enforces auth on every request.

### 7.5 Final docker-compose.yml

```yaml
version: '3.8'

services:

  # ── P2P NETWORKING (Go) ─────────────────────────────────────────
  libp2p-node:
    image: libp2p-node:latest        # Built from Go binary
    container_name: libp2p-node

    # THE NETWORKING FIX: bypasses Docker bridge NAT so go-libp2p
    # can see the real host network for hole punching precision
    network_mode: "host"

    user: "1001:1001"
    read_only: true
    cap_drop:
      - ALL
    cap_add:
      - NET_BIND_SERVICE             # only capability needed
    tmpfs:
      - /tmp:size=64m                # writable memory, no disk
    volumes:
      - peer-identity:/identity:ro   # own Ed25519 key, read-only
    deploy:
      resources:
        limits:
          cpus: '0.30'
          memory: 256M

  # ── GIT SERVER (Rust) ───────────────────────────────────────────
  git-server:
    image: git-server:latest         # Built from Rust binary
    container_name: git-server

    # ZERO NETWORK: physically unreachable from any network path
    # communicates only via unix socket with api-server
    network_mode: "none"

    user: "1002:1002"
    read_only: true
    cap_drop:
      - ALL
    volumes:
      - /path/to/repos:/repos:rw     # repos read-write
      - git-socket:/socket:rw        # unix socket only
    deploy:
      resources:
        limits:
          cpus: '0.50'
          memory: 512M

  # ── API SERVER + UI (Go + Fiber) ────────────────────────────────
  api-server:
    image: api-server:latest         # Built from Go binary, includes React static build
    container_name: api-server
    networks:
      - internal                     # internal bridge, no internet

    user: "1003:1003"
    read_only: true
    cap_drop:
      - ALL
    volumes:
      - /path/to/repos:/repos:rw
      - /path/to/db:/db:rw
      - git-socket:/socket:rw
      - p2p-socket:/p2p-socket:ro    # read commands from libp2p-node
    ports:
      - "127.0.0.1:3000:3000"        # localhost only, never 0.0.0.0
    deploy:
      resources:
        limits:
          cpus: '0.50'
          memory: 512M

networks:
  internal:
    driver: bridge
    internal: true                   # no external internet on this network

volumes:
  peer-identity:
  git-socket:
  p2p-socket:
```

> **Note:** The ui-server container is eliminated. The Go api-server binary embeds and serves the compiled React/Vite static build directly via Fiber's static file middleware. Zero Node.js in production.

### 7.6 Security Properties Summary

| Container | Language | Network | Filesystem | Repos Reachable | Internet Reachable |
|-----------|---------|---------|-----------|----------------|-------------------|
| libp2p-node | Go | host | None (identity key only) | ❌ No | ✅ Yes (required) |
| git-server | Rust | none | /repos only | ✅ Yes | ❌ No |
| api-server | Go | internal | /repos + /db | ✅ Yes | ❌ No |

### 7.7 Why This Beats systemd Bare-Metal

| Property | systemd jail | This Architecture |
|----------|-------------|------------------|
| Hole punching works | ✅ Yes | ✅ Yes (host network on libp2p-node) |
| RCE recovery | Reboot or manual cleanup | `docker restart` — Go binary rebuilt in seconds |
| Filesystem isolation | Manual cgroup config | Guaranteed by Linux namespaces |
| Malware persistence after restart | Possible | Impossible — container rebuilt from immutable image |
| Portability | Host OS specific | Runs identically on any Linux in minutes |
| Repo access if libp2p compromised | ⚠️ Possible | ❌ Impossible — not mounted, not networked |
| Language safety in git parser | Depends on implementation | ✅ Rust borrow checker — compile-time guaranteed |

---

## 8. Security Architecture

### 8.1 Security Layers

| Layer | Mechanism |
|-------|-----------|
| Identity | Peer ID = hash of Ed25519 public key. Cannot be spoofed without private key. |
| Transport encryption | Noise XX — `Noise_XX_25519_ChaChaPoly_SHA256`. Forward secrecy per session. |
| Authentication | SSH key only. No passwords ever accepted at any layer. |
| Authorization | Per-repo access control. Contributor A cannot see Repo B unless explicitly granted. |
| Commit integrity | Every commit signed with contributor's key. Tampering cryptographically detectable. |
| Memory safety | Rust borrow checker eliminates buffer overflow and use-after-free in git-server at compile time. |
| Container isolation | Docker least privilege. Breach contained to that container only. |
| Filesystem isolation | git-server (Rust) sees /repos only. Nothing else on host visible. |
| Network isolation | git-server has zero network interface. Unreachable from any network path. |
| At-rest encryption | Full disk encryption on host machine (LUKS on Linux). |
| Audit logging | Every push, pull, clone logged: timestamp, Peer ID, repo name, operation. |
| Key revocation | Remove contributor key → immediate access termination. |

### 8.2 Threat Model

| Attack | Likelihood | Severity | Defense |
|--------|-----------|----------|---------|
| Peer ID spoofing | Low | Critical | Noise XX — requires private key, mathematically impossible to fake |
| Man in the middle | Low | Critical | Noise XX + Peer ID hash verification — MITM immediately detected |
| Unauthorized push | Medium | High | SSH key auth + per-repo ACL in api-server |
| Buffer overflow in pack file parser | Low | Critical | Rust borrow checker — this class of exploit cannot be compiled |
| Path traversal via repo name | Medium | High | Rust type system + path jailing in api-server |
| Zero day in Go runtime | Low | Critical | Docker isolation — `docker restart` recovers in seconds |
| Compromised contributor key | Medium | High | Commit signing + audit log + instant key revocation |
| Malicious contributor | Medium | High | Protected branches + required review + audit log |
| Physical machine theft | Low | Critical | Full disk encryption (LUKS) |
| DHT poisoning | Low | Medium | Private closed network — strangers cannot join |
| Relay node compromise | Low | Low | E2E encryption — relay sees only ciphertext |
| Compromised libp2p-node container | Low | Medium | git-server has no network — repos unreachable from container |

---

## 9. Feature Roadmap by Phase

---

### Phase 1 — Local MVP

> **Goal: Two people on the same network can use this instead of GitHub.**

**Estimated build time: 2–3 months solo**

| Feature | Description |
|---------|-------------|
| Repo creation | Create bare Git repos via UI. Stored in /repos. Managed by Rust git-server. |
| Clone / Push / Pull | Standard Git over SSH and HTTP. Compatible with any Git client. |
| File browser | Browse repo files, view raw content in React UI. |
| Commit history | List commits with author, message, timestamp, hash. |
| Branch management | Create, switch, delete branches. View all branches. |
| Contributor management | Add contributors by SSH public key. Grant per-repo access. |
| Local UI | React/Vite static build served by Go/Fiber on localhost:3000. |
| Docker deployment | Single `docker-compose up`. Three containers. All least privilege. |
| mDNS discovery | Automatic peer discovery on local network. Zero config. |
| Audit log | All Git operations logged: timestamp, user, operation, repo. |

---

### Phase 2 — Cross-Network P2P

> **Goal: Contributor in a different city connects to your local machine seamlessly.**

**Estimated build time: 3–4 months additional**

| Feature | Description |
|---------|-------------|
| go-libp2p integration | Full libp2p-node container. Peer ID from Ed25519 key pair. |
| Kademlia DHT | Decentralized peer discovery across internet. No coordinator. |
| NAT hole punching | DCUtR via go-libp2p. Direct P2P through home and office NAT. |
| AutoNAT detection | Detects NAT type, selects optimal connection strategy. |
| Circuit relay fallback | Relay connection when hole punching fails on symmetric NAT. |
| Peer invite system | Share Peer ID to invite contributors. One approval step. |
| Offline commit queue | Commits queued on contributor device when host offline. |
| Auto-sync on reconnect | Queue flushes automatically when host comes back online. |
| Connection status UI | Clear indicator: online / offline / syncing / queued commits. |
| Noise encryption | All traffic encrypted with `Noise_XX_25519_ChaChaPoly_SHA256`. |

---

### Phase 3 — Self-Hostable Platform

> **Goal: Any developer installs this in one command and uses it as their GitHub replacement.**

**Estimated build time: 4–6 months additional**

| Feature | Description |
|---------|-------------|
| SSH key management UI | Add, remove, rotate SSH keys from web UI. |
| Protected branches | Prevent force push. Require review before merge. |
| Pull requests | Open, review, comment, merge PRs via UI. |
| Commit signing verification | Verify GPG and SSH signed commits. Show verified badge. |
| Webhook support | Trigger external scripts on push, PR open, merge. |
| REST API | Full API for all operations. Enables CLI and integrations. |
| PostgreSQL support | Swap from SQLite to PostgreSQL for larger teams. |
| Docker one-liner install | `curl` script → `docker-compose up` → running in 2 minutes. |
| Optional VPS mirror | One-way mirror to self-hosted VPS. Local always stays primary. |
| Key revocation | Instantly revoke access. All future pushes rejected. |

---

### Phase 4 — Collaboration Platform

> **Goal: Feature parity with Gitea for teams up to 50 people.**

**Estimated build time: 6–9 months additional**

| Feature | Description |
|---------|-------------|
| Issue tracker | Create, assign, label, close issues. Linked to commits and PRs. |
| Organizations | Group repos under an org. Team-level access control. |
| Code review UI | Inline diff comments. Review request workflow. |
| CI/CD webhook triggers | Trigger external pipelines (Drone, Woodpecker, Forgejo Actions). |
| Repo forking | Fork repos within the platform. Sync with upstream. |
| Release management | Tag releases, attach binaries, write release notes. |
| Activity feed | Dashboard: recent commits, PRs, issues across all repos. |
| Email notifications | Notify on PRs, reviews, mentions via local SMTP relay. |
| Repo templates | Create new repos from predefined templates. |
| Advanced permissions | Branch-level permissions. Role-based access control. |

---

### Phase 5 — Gitea-Scale

> **Goal: Thousands of self-hosted instances worldwide. Production-grade reliability.**

**Estimated build time: 12–18 months additional, small team**

| Feature | Description |
|---------|-------------|
| Horizontal scaling | Multiple api-server instances behind a load balancer. |
| Object storage for repos | MinIO or S3-compatible storage for repository data. |
| Multi-device hosting | Server across multiple machines. Distributed repo storage. |
| Federation | Two instances share repos with each other over libp2p. |
| Plugin system | Extension API for third-party features. |
| Admin dashboard | System health, storage, active connections, audit logs. |
| Migration tools | Import from GitHub, GitLab, Gitea with full history. |
| Performance benchmarks | Documented targets and regression test suite. |

---

## 10. Offline & Resilience Behavior

### 10.1 When Host Goes Offline

| Event | System Behavior |
|-------|----------------|
| Host closes laptop | libp2p peer unreachable. Contributors see "Host Offline" in UI. |
| Contributor tries to push | Push queued locally on contributor device. Clear UI confirmation. |
| Contributor continues working | All local Git operations work normally. Commits queue locally. |
| Host comes back online | go-libp2p detects peer available. Auto-flush begins immediately. |
| Queue flush completes | All queued commits pushed. UI shows "Synced". Conflict UI if needed. |
| Multiple contributors queued | Pushes replayed in timestamp order. Standard Git conflict resolution. |

### 10.2 Device Options for Always-On Hosting

| Device | Best For |
|--------|----------|
| Developer's laptop | Solo work, occasional collaboration |
| Old PC left running | Small startup, always-on, no extra cost |
| Raspberry Pi 4/5 (~$70) | Dedicated silent server, low power, always-on |
| Mini PC (Intel NUC etc.) | Larger teams, more storage, better reliability |
| Self-hosted VPS (optional) | Backup mirror only — never primary |

---

## 11. Non-Functional Requirements

| Requirement | Phase 1 | Phase 3 | Phase 5 |
|-------------|---------|---------|---------|
| Repo clone speed (LAN) | Full wire speed | Full wire speed | Full wire speed |
| Repo clone speed (P2P) | N/A | > 10 MB/s | > 50 MB/s |
| Peer discovery time | N/A | < 30 seconds | < 10 seconds |
| Concurrent P2P connections | N/A | 100 | 10,000+ |
| UI load time | < 1 second | < 1 second | < 500ms |
| API response time (p99) | < 100ms | < 50ms | < 20ms |
| Max repos | Unlimited (disk) | Unlimited | Unlimited |
| Max contributors | 10 | 50 | 1000+ |
| Docker startup time | < 30 seconds | < 20 seconds | < 10 seconds |
| Audit log retention | 90 days | 1 year | Configurable |
| Memory per container | < 512M | < 512M | Configurable |

---

## 12. Open Source Dependencies

> Every dependency must be open source with a permissive license. No proprietary tools, no paid services, no external coordinators. Zero Node.js in production.

### Go Dependencies

| Package | License | Purpose |
|---------|---------|---------|
| `go-libp2p` | Apache 2.0 | P2P networking, NAT traversal, DHT, encryption |
| `gofiber/fiber` | MIT | HTTP routing, static file serving for React UI |
| `modernc.org/sqlite` | MIT | Pure Go SQLite — no cgo required |
| `lib/pq` | MIT | PostgreSQL driver (Phase 3+) |

### Rust Dependencies

| Crate | License | Purpose |
|-------|---------|---------|
| `git2` | MIT / Apache 2.0 | Safe bindings over libgit2 for all Git operations |
| `tokio` | MIT | Async runtime for Unix socket server |
| `serde` | MIT / Apache 2.0 | Serialization for socket protocol messages |
| `rusqlite` | MIT | SQLite bindings if git-server needs local state |

### Frontend Dependencies

| Package | License | Purpose |
|---------|---------|---------|
| `react` | MIT | UI framework |
| `vite` | MIT | Build tool — produces static files served by Go |
| `tailwindcss` | MIT | UI styling |

### Infrastructure

| Tool | License | Purpose |
|------|---------|---------|
| `docker` | Apache 2.0 | Container isolation and orchestration |
| `git` | GPL-2.0 | Core version control (system dependency) |

---

## 13. Out of Scope

The following will **not** be built:

- Monetization of any kind — no freemium, no enterprise tiers, no ads, ever
- Cloud-hosted SaaS version — local-first by design
- Mobile native apps — web UI works on mobile browsers
- Built-in CI/CD runner — webhooks enable external tools
- Built-in email server — SMTP relay configuration only
- Windows native installer in Phase 1 — Docker handles cross-platform
- Proprietary protocol extensions — stay compatible with standard Git clients
- Node.js in production — eliminated entirely in favour of Go serving static React build

---

## 14. Success Metrics

| Phase | Success Looks Like |
|-------|-------------------|
| Phase 1 | Two people replace GitHub for a project on local network. Zero data leaks. |
| Phase 2 | Two contributors on different ISPs collaborate with zero external servers. |
| Phase 3 | Any developer installs and runs it in under 5 minutes via Docker. |
| Phase 4 | A startup of 10 uses it as sole Git platform for 6 months. |
| Phase 5 | 100+ developers self-host globally. Community contributes to codebase. |

---

## 15. Risks & Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|-----------|
| Rust compile times slow iteration | High | Medium | `cargo watch` for dev, pre-built Docker images for testing, only git-server is Rust |
| Go/Rust Unix socket protocol mismatch | Medium | High | Define strict protobuf/JSON schema for socket messages from day one |
| go-libp2p API changes between versions | Medium | Medium | Pin exact version, integration tests on every update |
| NAT traversal failure on strict networks | Medium | Medium | Circuit relay fallback always available — connectivity preserved |
| Single developer bandwidth | High | High | Phased roadmap — Phase 1 is fully useful standalone |
| git2-rs / libgit2 vulnerability | Low | High | Docker isolation contains blast radius, update policy enforced |
| Data loss from disk failure | Low | Critical | Optional VPS mirror, documented /repos volume backup process |
| Physical machine theft | Low | Critical | LUKS full disk encryption, documented setup in install guide |

---

## 16. Appendix — Key Concepts

### go-libp2p
The reference implementation of the libp2p networking stack, written in Go. Powers IPFS in production globally. Provides TCP transport, Kademlia DHT, mDNS, Noise encryption, DCUtR hole punching, AutoNAT, Circuit Relay, and Yamux multiplexing. Every major libp2p protocol decision is validated against this implementation first.

### Rust Borrow Checker
The Rust compiler's ownership system that enforces memory safety at compile time with zero runtime overhead. It makes buffer overflows, use-after-free, and data races impossible to ship — not just unlikely. For a git-server parsing untrusted binary pack files, this is not a nice-to-have. It eliminates the most dangerous class of remote code execution vulnerabilities before the binary is ever built.

### git2-rs
Safe Rust bindings over `libgit2` — the same C library that powers GitHub, GitLab, and Gitea. Provides a complete API for all Git operations: reading objects, writing pack files, managing references, walking commit history. Rust's borrow checker wraps every call with memory safety guarantees the underlying C library cannot provide.

### Kademlia DHT
A distributed hash table algorithm for decentralized peer discovery. Peers find each other in O(log n) hops with no central server. Each peer stores routing information for peers whose ID is numerically close to their own in XOR space. On a network of 1 million peers, any peer is reachable in approximately 20 hops.

### Noise Protocol — `Noise_XX_25519_ChaChaPoly_SHA256`
The specific Noise handshake pattern used by go-libp2p. Provides mutual authentication (both sides verify each other's identity) and forward secrecy (each session uses ephemeral keys that are destroyed after use). Your Peer ID is derived from your Ed25519 public key — identity and encryption are unified. No certificate authority required.

### DCUtR (Direct Connection Upgrade through Relay)
A libp2p protocol that coordinates NAT hole punching with millisecond timing precision using goroutines in the Go implementation. Both peers connect to a relay, exchange their observed public addresses, synchronise clocks via relay round-trip measurement, then punch through their NATs simultaneously. The relay drops out after direct connection is established.

### Unix Domain Sockets
IPC mechanism where processes communicate via a filesystem path rather than a network address. Approximately 4–8x lower latency than TCP on localhost. Cannot be accessed from any network path — only processes with filesystem permission to the socket file can connect. Used for all communication between api-server (Go) and git-server (Rust).

### network_mode: host (libp2p-node only)
Gives the libp2p-node container direct access to the host's real network interface, bypassing Docker's internal bridge NAT. Required for go-libp2p hole punching to work correctly — without it, libp2p reports a fake Docker-internal IP to relay peers and all cross-network connections fail silently. Acceptable trade-off: the libp2p-node container has no access to repository data, the database, or other containers.

### Least Privilege (Docker)
Each container runs with the minimum permissions it strictly needs. `cap_drop: ALL` removes all Linux capabilities. Only `NET_BIND_SERVICE` is re-added for libp2p-node. `read_only: true` makes the container filesystem immutable. `user: 1000+` ensures no process runs as root. A compromised container cannot escalate privileges, access other containers' data, or persist malware after a restart.

### Circuit Relay
When NAT hole punching fails — typically on symmetric NAT used by some corporate networks and mobile carriers — go-libp2p automatically falls back to routing traffic through a relay peer found via DHT. The relay forwards encrypted bytes it cannot read. Privacy is fully maintained. Connectivity is preserved at the cost of some added latency.

---

*End of Document — Version 1.2 — Go + Rust hybrid architecture — All phases subject to revision*
