# 🌐 P2P Git — Local-First Peer-to-Peer Git Hosting

A **local-first, peer-to-peer Git hosting platform** that lets you host, clone, and collaborate on Git repositories without any central server. Built on [libp2p](https://libp2p.io/) for peer discovery and data transfer, with a hardened three-container architecture that isolates the git backend from the network.

> **Status:** 🚧 Under active development

---

## ✨ Features

- **No central server** — repositories live on your machine and replicate peer-to-peer.
- **LAN-first discovery** — mDNS finds peers on your local network automatically.
- **Container-isolated git backend** — the Rust git server has _no network access_ (`network: none`); it communicates only via Unix domain sockets.
- **Noise-encrypted transport** — all peer-to-peer traffic is encrypted with the libp2p Noise protocol.
- **Per-repo access control** — fine-grained ACLs let you decide who can read or push to each repository.
- **Modern web UI** — React + TypeScript dashboard for managing repos, peers, and settings.

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        Host Machine                             │
│                                                                 │
│  ┌──────────────────┐   Unix Socket   ┌──────────────────────┐  │
│  │   libp2p-node    │◄──────────────►│     api-server        │  │
│  │   (Go, host net) │  /run/p2p/     │  (Go + Fiber, bridge) │  │
│  └────────┬─────────┘   libp2p.sock  └──────────┬───────────┘  │
│           │                                      │              │
│           │  Unix Socket                         │ Unix Socket  │
│           │  /run/git/git.sock                   │ /run/git/    │
│           │                                      │ git.sock     │
│           ▼                                      ▼              │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │              git-server  (Rust, network: none)            │  │
│  │                                                           │  │
│  │   • Bare repo management    • Pack-file serving           │  │
│  │   • Ref advertisement       • Object validation           │  │
│  └───────────────────────────────────────────────────────────┘  │
│                                                                 │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │                Web UI  (React + TypeScript)               │  │
│  │           http://localhost:3000  →  api-server             │  │
│  └───────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

| Service        | Language     | Network Mode   | Purpose                                      |
|----------------|-------------|----------------|----------------------------------------------|
| `libp2p-node`  | Go          | `host`         | Peer discovery (mDNS), data relay, Noise transport |
| `git-server`   | Rust        | `none`         | Git repository storage & pack-file operations |
| `api-server`   | Go + Fiber  | `bridge`       | REST/WebSocket API for web UI & CLI           |

All inter-service communication uses **Unix domain sockets** — no TCP between containers.

---

## 📋 Prerequisites

| Tool             | Minimum Version | Purpose                        |
|------------------|----------------|---------------------------------|
| Docker           | 24.0+          | Container runtime               |
| Docker Compose   | 2.20+          | Multi-container orchestration   |
| Go               | 1.22+          | API server & libp2p node builds |
| Rust             | 1.77+          | Git server build                |
| Node.js          | 20 LTS+        | Web UI build                    |
| Make             | any            | Build automation                |

---

## 🚀 Quick Start

```bash
# 1. Clone the repository
git clone https://github.com/your-org/p2p-git.git
cd p2p-git

# 2. Copy the environment template
cp .env.example .env

# 3. Run the setup script (creates dirs, generates peer key, builds images)
make setup

# 4. Start all services
make up

# 5. Open the web UI
#    → http://localhost:3000
```

### Development Mode

```bash
# Start with live-reload and verbose logging
make dev
```

---

## 📁 Project Structure

```
.
├── docker-compose.yml          # Three-service container orchestration
├── Makefile                    # Build, test, and run automation
├── .env.example                # Environment variable template
│
├── services/
│   ├── api-server/             # Go + Fiber HTTP API
│   │   ├── cmd/server/         # Entrypoint
│   │   ├── internal/           # Business logic
│   │   ├── Dockerfile
│   │   └── go.mod
│   │
│   ├── git-server/             # Rust git backend
│   │   ├── src/                # Source code
│   │   ├── Dockerfile
│   │   └── Cargo.toml
│   │
│   └── libp2p-node/            # Go libp2p networking daemon
│       ├── cmd/node/           # Entrypoint
│       ├── internal/           # Peer management, protocols
│       ├── Dockerfile
│       └── go.mod
│
├── ui/                         # React + TypeScript web frontend
│   ├── src/
│   ├── package.json
│   └── vite.config.ts
│
├── docs/                       # Architecture & development docs
│   ├── architecture.md
│   ├── development.md
│   └── security.md
│
├── scripts/                    # Automation scripts
│   ├── setup.sh
│   └── dev.sh
│
└── data/                       # Local runtime data (git-ignored)
    ├── repos/                  # Bare git repositories
    └── db/                     # SQLite database
```

---

## 🧪 Testing

```bash
# Run all tests
make test

# Test individual services
make test-api
make test-git
make test-libp2p
make test-ui
```

---

## 🔒 Security Model

- **Container isolation** — The git server runs with `network: none`; it cannot make or receive network connections.
- **Unix socket IPC** — All inter-service communication is local-only, through Unix domain sockets with filesystem permissions.
- **Noise protocol** — Peer-to-peer connections are authenticated and encrypted using the libp2p Noise handshake.
- **Per-repo ACLs** — Repository access is controlled per-peer with read/write permissions stored locally.

See [docs/security.md](docs/security.md) for the full threat model.

---

## 📄 License

This project is licensed under the **MIT License**. See [LICENSE](LICENSE) for details.

---

## 🤝 Contributing

Contributions are welcome! Please read the [development guide](docs/development.md) before submitting a pull request.

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-thing`)
3. Commit your changes (`git commit -m 'Add amazing thing'`)
4. Push to the branch (`git push origin feature/amazing-thing`)
5. Open a Pull Request
# GITLINK
