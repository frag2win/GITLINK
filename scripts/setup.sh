#!/usr/bin/env bash
# =============================================================================
# setup.sh — Initial Setup for P2P Git Hosting Platform
# =============================================================================
# This script:
#   1. Creates required data directories
#   2. Generates a peer identity key (Ed25519) if one doesn't exist
#   3. Copies .env.example → .env if .env doesn't exist
#   4. Installs service dependencies
#   5. Builds all Docker images
#   6. Starts the platform via docker-compose
# =============================================================================

set -euo pipefail

# ---------------------------------------------------------------------------
# Configuration
# ---------------------------------------------------------------------------

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DATA_DIR="${PROJECT_ROOT}/data"
REPOS_DIR="${DATA_DIR}/repos"
DB_DIR="${DATA_DIR}/db"
BIN_DIR="${PROJECT_ROOT}/bin"
PEER_KEY_DIR="${DATA_DIR}/identity"
PEER_KEY_FILE="${PEER_KEY_DIR}/peer.key"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# ---------------------------------------------------------------------------
# Helper functions
# ---------------------------------------------------------------------------

info()  { echo -e "${BLUE}ℹ${NC}  $*"; }
ok()    { echo -e "${GREEN}✅${NC} $*"; }
warn()  { echo -e "${YELLOW}⚠${NC}  $*"; }
error() { echo -e "${RED}❌${NC} $*" >&2; }

check_command() {
    if ! command -v "$1" &> /dev/null; then
        error "$1 is not installed. Please install it before running setup."
        echo "   See: $2"
        return 1
    fi
    ok "$1 found: $(command -v "$1")"
}

# ---------------------------------------------------------------------------
# Step 1: Check prerequisites
# ---------------------------------------------------------------------------

echo ""
echo "=============================="
echo "  P2P Git — Initial Setup"
echo "=============================="
echo ""

info "Checking prerequisites..."

MISSING=0
check_command "docker"         "https://docs.docker.com/get-docker/"         || MISSING=1
check_command "docker"         "https://docs.docker.com/compose/install/"    || true  # compose is a subcommand
check_command "go"             "https://go.dev/dl/"                          || MISSING=1
check_command "rustc"          "https://rustup.rs/"                          || MISSING=1
check_command "cargo"          "https://rustup.rs/"                          || MISSING=1
check_command "node"           "https://nodejs.org/"                         || MISSING=1
check_command "npm"            "https://nodejs.org/"                         || MISSING=1

# Check Docker Compose (v2 plugin)
if docker compose version &> /dev/null; then
    ok "docker compose found: $(docker compose version --short 2>/dev/null || echo 'v2')"
else
    error "docker compose (v2) is not available."
    MISSING=1
fi

if [ "$MISSING" -eq 1 ]; then
    error "Some prerequisites are missing. Please install them and re-run this script."
    exit 1
fi

echo ""

# ---------------------------------------------------------------------------
# Step 2: Create data directories
# ---------------------------------------------------------------------------

info "Creating data directories..."

mkdir -p "${REPOS_DIR}"
mkdir -p "${DB_DIR}"
mkdir -p "${BIN_DIR}"
mkdir -p "${PEER_KEY_DIR}"

ok "Created: ${REPOS_DIR}"
ok "Created: ${DB_DIR}"
ok "Created: ${BIN_DIR}"
ok "Created: ${PEER_KEY_DIR}"

echo ""

# ---------------------------------------------------------------------------
# Step 3: Generate peer identity key
# ---------------------------------------------------------------------------

info "Checking peer identity key..."

if [ -f "${PEER_KEY_FILE}" ]; then
    ok "Peer key already exists at ${PEER_KEY_FILE}"
else
    info "Generating new Ed25519 peer identity key..."

    # Generate a 32-byte random seed for Ed25519
    # This is a placeholder — the actual libp2p node will generate a proper
    # libp2p-compatible key on first boot if this file doesn't exist.
    # For now, we create a random seed that the node will use.
    if command -v openssl &> /dev/null; then
        openssl genpkey -algorithm ed25519 -out "${PEER_KEY_FILE}" 2>/dev/null
        ok "Generated Ed25519 key via openssl → ${PEER_KEY_FILE}"
    else
        # Fallback: generate 32 random bytes as a seed
        dd if=/dev/urandom bs=32 count=1 2>/dev/null > "${PEER_KEY_FILE}"
        ok "Generated random key seed → ${PEER_KEY_FILE}"
    fi

    # Restrict permissions
    chmod 600 "${PEER_KEY_FILE}"
    ok "Set key permissions to 600 (owner-only)"
fi

echo ""

# ---------------------------------------------------------------------------
# Step 4: Create .env file if it doesn't exist
# ---------------------------------------------------------------------------

info "Checking .env file..."

if [ -f "${PROJECT_ROOT}/.env" ]; then
    ok ".env file already exists"
else
    cp "${PROJECT_ROOT}/.env.example" "${PROJECT_ROOT}/.env"
    ok "Created .env from .env.example"
    warn "Review ${PROJECT_ROOT}/.env and adjust settings if needed."
fi

echo ""

# ---------------------------------------------------------------------------
# Step 5: Install dependencies
# ---------------------------------------------------------------------------

info "Installing Go dependencies for api-server..."
if [ -f "${PROJECT_ROOT}/services/api-server/go.mod" ]; then
    (cd "${PROJECT_ROOT}/services/api-server" && go mod download)
    ok "api-server dependencies installed"
else
    warn "services/api-server/go.mod not found — skipping"
fi

info "Installing Go dependencies for libp2p-node..."
if [ -f "${PROJECT_ROOT}/services/libp2p-node/go.mod" ]; then
    (cd "${PROJECT_ROOT}/services/libp2p-node" && go mod download)
    ok "libp2p-node dependencies installed"
else
    warn "services/libp2p-node/go.mod not found — skipping"
fi

info "Installing Rust dependencies for git-server..."
if [ -f "${PROJECT_ROOT}/services/git-server/Cargo.toml" ]; then
    (cd "${PROJECT_ROOT}/services/git-server" && cargo fetch)
    ok "git-server dependencies installed"
else
    warn "services/git-server/Cargo.toml not found — skipping"
fi

info "Installing Node.js dependencies for web UI..."
if [ -f "${PROJECT_ROOT}/ui/package.json" ]; then
    (cd "${PROJECT_ROOT}/ui" && npm ci)
    ok "UI dependencies installed"
else
    warn "ui/package.json not found — skipping"
fi

echo ""

# ---------------------------------------------------------------------------
# Step 6: Build Docker images
# ---------------------------------------------------------------------------

info "Building Docker images..."

if docker compose -f "${PROJECT_ROOT}/docker-compose.yml" build; then
    ok "Docker images built successfully"
else
    warn "Docker image build failed — some service directories may not exist yet."
    warn "You can build later with: make up"
fi

echo ""

# ---------------------------------------------------------------------------
# Step 7: Start services
# ---------------------------------------------------------------------------

info "Starting services..."

if docker compose -f "${PROJECT_ROOT}/docker-compose.yml" up -d; then
    ok "All services started!"
    echo ""
    echo "=============================="
    echo "  Setup Complete!"
    echo "=============================="
    echo ""
    echo "  🌐 Web UI:    http://localhost:${API_PORT:-3000}"
    echo "  📁 Repos:     ${REPOS_DIR}"
    echo "  🗄️  Database:  ${DB_DIR}/api.db"
    echo "  🔑 Peer Key:  ${PEER_KEY_FILE}"
    echo ""
    echo "  Useful commands:"
    echo "    make up     — Start services"
    echo "    make down   — Stop services"
    echo "    make dev    — Development mode"
    echo "    make test   — Run all tests"
    echo "    make help   — Show all commands"
    echo ""
else
    warn "Services failed to start. Check logs with: docker compose logs"
fi
