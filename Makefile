# =============================================================================
# Makefile — Local-First P2P Git Hosting Platform
# =============================================================================
# Build, test, and run commands for all three backend services and the web UI.
# =============================================================================

.PHONY: all build build-api build-git build-libp2p build-ui \
        up down dev test clean fmt lint help setup

# Default target
all: build

# ---------------------------------------------------------------------------
# Build targets
# ---------------------------------------------------------------------------

## Build all services
build: build-api build-git build-libp2p build-ui
	@echo "✅ All services built successfully."

## Build the Go + Fiber API server
build-api:
	@echo "🔨 Building api-server (Go + Fiber)..."
	cd services/api-server && go build -o ../../bin/api-server ./cmd/server
	@echo "   ✅ api-server built → bin/api-server"

## Build the Rust git backend server
build-git:
	@echo "🔨 Building git-server (Rust)..."
	cd services/git-server && cargo build --release
	cp services/git-server/target/release/git-server bin/git-server
	@echo "   ✅ git-server built → bin/git-server"

## Build the Go libp2p networking daemon
build-libp2p:
	@echo "🔨 Building libp2p-node (Go)..."
	cd services/libp2p-node && go build -o ../../bin/libp2p-node ./cmd/node
	@echo "   ✅ libp2p-node built → bin/libp2p-node"

## Build the React web UI
build-ui:
	@echo "🔨 Building web UI (React + TypeScript)..."
	cd ui && npm ci && npm run build
	@echo "   ✅ UI built → ui/dist"

# ---------------------------------------------------------------------------
# Docker targets
# ---------------------------------------------------------------------------

## Start all services via Docker Compose (detached)
up:
	@echo "🚀 Starting all services..."
	docker compose up -d --build
	@echo "   ✅ Services are running. API available at http://localhost:$${API_PORT:-3000}"

## Stop all services and remove containers
down:
	@echo "🛑 Stopping all services..."
	docker compose down
	@echo "   ✅ All services stopped."

## Start services in development mode with live reload
dev:
	@echo "🔧 Starting development environment..."
	@mkdir -p data/repos data/db
	docker compose -f docker-compose.yml -f docker-compose.dev.yml up --build
	@echo "   Press Ctrl+C to stop."

# ---------------------------------------------------------------------------
# Quality targets
# ---------------------------------------------------------------------------

## Run all tests across every service
test: test-api test-git test-libp2p test-ui
	@echo "✅ All tests passed."

test-api:
	@echo "🧪 Testing api-server..."
	cd services/api-server && go test -v -race ./...

test-git:
	@echo "🧪 Testing git-server..."
	cd services/git-server && cargo test

test-libp2p:
	@echo "🧪 Testing libp2p-node..."
	cd services/libp2p-node && go test -v -race ./...

test-ui:
	@echo "🧪 Testing web UI..."
	cd ui && npm test -- --watchAll=false

## Format all source code
fmt:
	@echo "🎨 Formatting code..."
	cd services/api-server && go fmt ./...
	cd services/libp2p-node && go fmt ./...
	cd services/git-server && cargo fmt
	cd ui && npx prettier --write "src/**/*.{ts,tsx,css}"
	@echo "   ✅ All code formatted."

## Run linters across every service
lint:
	@echo "🔍 Linting code..."
	cd services/api-server && golangci-lint run ./...
	cd services/libp2p-node && golangci-lint run ./...
	cd services/git-server && cargo clippy -- -D warnings
	cd ui && npx eslint src/
	@echo "   ✅ Lint checks passed."

# ---------------------------------------------------------------------------
# Utility targets
# ---------------------------------------------------------------------------

## Initial project setup — create directories, generate identity, pull deps
setup:
	@echo "⚙️  Running initial setup..."
	@mkdir -p bin data/repos data/db
	@bash scripts/setup.sh
	@echo "   ✅ Setup complete."

## Clean all build artifacts, containers, volumes, and cached data
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -rf bin/
	rm -rf ui/dist ui/node_modules
	cd services/git-server && cargo clean 2>/dev/null || true
	docker compose down -v --rmi local 2>/dev/null || true
	@echo "   ✅ Clean complete."

## Print this help message
help:
	@echo ""
	@echo "P2P Git Hosting Platform — Available Commands"
	@echo "=============================================="
	@echo ""
	@echo "  Build:"
	@echo "    make build          Build all services"
	@echo "    make build-api      Build the API server (Go + Fiber)"
	@echo "    make build-git      Build the git server (Rust)"
	@echo "    make build-libp2p   Build the libp2p node (Go)"
	@echo "    make build-ui       Build the web UI (React)"
	@echo ""
	@echo "  Run:"
	@echo "    make up             Start services (Docker, detached)"
	@echo "    make down           Stop services"
	@echo "    make dev            Start in dev mode with live reload"
	@echo ""
	@echo "  Quality:"
	@echo "    make test           Run all tests"
	@echo "    make fmt            Format all code"
	@echo "    make lint           Lint all code"
	@echo ""
	@echo "  Utility:"
	@echo "    make setup          Initial project setup"
	@echo "    make clean          Remove all build artifacts"
	@echo "    make help           Show this help message"
	@echo ""
