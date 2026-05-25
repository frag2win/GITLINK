# Development Guide — P2P Git Hosting Platform

## Prerequisites

Before you begin, make sure you have the following tools installed:

| Tool             | Version  | Installation                                      |
|------------------|----------|---------------------------------------------------|
| Go               | 1.22+    | https://go.dev/dl/                                |
| Rust             | 1.77+    | `curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs \| sh` |
| Node.js          | 20 LTS+  | https://nodejs.org/ or use `nvm`                   |
| Docker           | 24.0+    | https://docs.docker.com/get-docker/                |
| Docker Compose   | 2.20+    | Included with Docker Desktop                       |
| Make             | any      | Pre-installed on macOS/Linux; use `choco install make` on Windows |
| golangci-lint    | latest   | `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest` |

---

## Initial Setup

```bash
# 1. Clone the repository
git clone https://github.com/your-org/p2p-git.git
cd p2p-git

# 2. Copy the environment template and adjust if needed
cp .env.example .env

# 3. Run the automated setup (creates directories, generates peer key, installs deps)
make setup

# Alternatively, do it manually:
mkdir -p data/repos data/db bin
```

---

## Building Each Service

### API Server (Go + Fiber)

```bash
# Build
make build-api

# Or manually:
cd services/api-server
go build -o ../../bin/api-server ./cmd/server

# Run tests
cd services/api-server
go test -v -race ./...

# Run locally (outside Docker)
./bin/api-server \
  --port 3000 \
  --git-socket ./data/git.sock \
  --p2p-socket ./data/p2p.sock \
  --db-path ./data/db/api.db
```

### Git Server (Rust)

```bash
# Build
make build-git

# Or manually:
cd services/git-server
cargo build --release
cp target/release/git-server ../../bin/

# Run tests
cargo test

# Run locally
./bin/git-server \
  --socket-path ./data/git.sock \
  --repos-path ./data/repos
```

### libp2p Node (Go)

```bash
# Build
make build-libp2p

# Or manually:
cd services/libp2p-node
go build -o ../../bin/libp2p-node ./cmd/node

# Run tests
go test -v -race ./...

# Run locally
./bin/libp2p-node \
  --key-path ./data/peer.key \
  --p2p-socket ./data/p2p.sock \
  --git-socket ./data/git.sock \
  --listen-port 4001
```

### Web UI (React + TypeScript)

```bash
# Build
make build-ui

# Or manually:
cd ui
npm ci
npm run build      # Production build → ui/dist/

# Development server with hot reload
npm run dev        # → http://localhost:5173 (proxies API to :3000)

# Run tests
npm test
```

---

## Running Locally

### With Docker (Recommended)

```bash
# Start all services in the background
make up

# View logs
docker compose logs -f

# View logs for a specific service
docker compose logs -f api-server

# Stop everything
make down
```

### Development Mode

```bash
# Start with docker-compose.dev.yml overlay (enables live-reload, verbose logging)
make dev

# Or use the convenience script
bash scripts/dev.sh
```

### Without Docker (Manual)

If you prefer to run services directly on your host for debugging:

```bash
# Terminal 1 — Git Server
./bin/git-server --socket-path /tmp/git.sock --repos-path ./data/repos

# Terminal 2 — libp2p Node
./bin/libp2p-node --key-path ./data/peer.key --p2p-socket /tmp/p2p.sock --git-socket /tmp/git.sock

# Terminal 3 — API Server
./bin/api-server --port 3000 --git-socket /tmp/git.sock --p2p-socket /tmp/p2p.sock --db-path ./data/db/api.db

# Terminal 4 — Web UI dev server
cd ui && npm run dev
```

---

## Running Tests

```bash
# All tests
make test

# Individual services
make test-api       # Go tests for api-server
make test-git       # Rust tests for git-server
make test-libp2p    # Go tests for libp2p-node
make test-ui        # React/Jest tests for web UI

# With coverage (Go)
cd services/api-server && go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# With coverage (Rust)
cd services/git-server && cargo tarpaulin --out Html

# With coverage (UI)
cd ui && npm test -- --coverage
```

---

## Code Quality

```bash
# Format all code
make fmt

# Lint all code
make lint
```

### Formatting Details

| Service      | Formatter                  |
|-------------|---------------------------|
| Go services | `go fmt`                  |
| Rust service | `cargo fmt`              |
| Web UI      | Prettier                  |

### Linting Details

| Service      | Linter                    |
|-------------|---------------------------|
| Go services | `golangci-lint`           |
| Rust service | `cargo clippy`           |
| Web UI      | ESLint                    |

---

## Project Conventions

### Git Workflow

- Use feature branches: `feature/`, `fix/`, `docs/`
- Write conventional commits: `feat:`, `fix:`, `docs:`, `refactor:`, `test:`
- Keep PRs small and focused on a single change

### Code Style

- **Go** — Follow standard `gofmt` conventions. Use `internal/` for unexported packages.
- **Rust** — Follow `rustfmt` defaults. Use `clippy` with `-D warnings`.
- **TypeScript** — Follow the project's ESLint + Prettier config. Prefer functional components and hooks.

### Directory Layout Conventions

```
services/<name>/
  ├── cmd/           # Binary entrypoint(s)
  ├── internal/      # Private packages (Go) or modules (Rust)
  ├── pkg/           # Public packages (if any)
  ├── Dockerfile     # Multi-stage production build
  └── go.mod / Cargo.toml
```

---

## Troubleshooting

### Docker socket permission errors

```bash
# Add your user to the docker group
sudo usermod -aG docker $USER
# Log out and back in for changes to take effect
```

### Port 3000 already in use

```bash
# Change the port in .env
echo "API_PORT=3001" >> .env
make up
```

### Git server healthcheck failing

```bash
# Check if the Unix socket exists
docker exec p2p-git-server ls -la /run/git/

# Check git-server logs
docker compose logs git-server
```

### libp2p node can't discover peers

```bash
# Ensure mDNS is enabled
grep MDNS_ENABLED .env

# Check that the host firewall allows UDP multicast on port 5353
# On Linux:
sudo ufw allow 5353/udp

# Check libp2p logs for discovery events
docker compose logs libp2p-node | grep -i "discovered"
```
