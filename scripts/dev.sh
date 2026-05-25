#!/usr/bin/env bash
# =============================================================================
# dev.sh — Development Convenience Script for P2P Git Hosting Platform
# =============================================================================
# Starts all services in development mode with:
#   • Verbose logging (LOG_LEVEL=debug)
#   • Live-reload where supported
#   • All logs streamed to the terminal
#   • Automatic directory creation
#   • Graceful shutdown on Ctrl+C
# =============================================================================

set -euo pipefail

# ---------------------------------------------------------------------------
# Configuration
# ---------------------------------------------------------------------------

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DATA_DIR="${PROJECT_ROOT}/data"
REPOS_DIR="${DATA_DIR}/repos"
DB_DIR="${DATA_DIR}/db"

# Override log level for development
export LOG_LEVEL="debug"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# ---------------------------------------------------------------------------
# Helper functions
# ---------------------------------------------------------------------------

info()  { echo -e "${BLUE}ℹ${NC}  $*"; }
ok()    { echo -e "${GREEN}✅${NC} $*"; }
warn()  { echo -e "${YELLOW}⚠${NC}  $*"; }

# ---------------------------------------------------------------------------
# Trap for graceful shutdown
# ---------------------------------------------------------------------------

cleanup() {
    echo ""
    info "Shutting down development services..."
    docker compose -f "${PROJECT_ROOT}/docker-compose.yml" down 2>/dev/null || true

    # Kill any background processes we started
    if [ -n "${UI_PID:-}" ]; then
        kill "${UI_PID}" 2>/dev/null || true
    fi

    ok "Development environment stopped."
    exit 0
}

trap cleanup SIGINT SIGTERM

# ---------------------------------------------------------------------------
# Banner
# ---------------------------------------------------------------------------

echo ""
echo -e "${CYAN}╔══════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║   P2P Git — Development Mode             ║${NC}"
echo -e "${CYAN}╚══════════════════════════════════════════╝${NC}"
echo ""

# ---------------------------------------------------------------------------
# Step 1: Ensure data directories exist
# ---------------------------------------------------------------------------

info "Ensuring data directories exist..."
mkdir -p "${REPOS_DIR}" "${DB_DIR}"
ok "Data directories ready"

# ---------------------------------------------------------------------------
# Step 2: Ensure .env exists
# ---------------------------------------------------------------------------

if [ ! -f "${PROJECT_ROOT}/.env" ]; then
    warn ".env not found — copying from .env.example"
    cp "${PROJECT_ROOT}/.env.example" "${PROJECT_ROOT}/.env"
    ok "Created .env from .env.example"
fi

# Force debug logging in dev mode
if ! grep -q "LOG_LEVEL=debug" "${PROJECT_ROOT}/.env"; then
    info "Setting LOG_LEVEL=debug for development"
fi

# ---------------------------------------------------------------------------
# Step 3: Check if docker-compose.dev.yml exists
# ---------------------------------------------------------------------------

COMPOSE_FILES="-f ${PROJECT_ROOT}/docker-compose.yml"

if [ -f "${PROJECT_ROOT}/docker-compose.dev.yml" ]; then
    COMPOSE_FILES="${COMPOSE_FILES} -f ${PROJECT_ROOT}/docker-compose.dev.yml"
    info "Using docker-compose.dev.yml overlay"
else
    info "No docker-compose.dev.yml found — using base config with debug logging"
fi

# ---------------------------------------------------------------------------
# Step 4: Build and start backend services
# ---------------------------------------------------------------------------

info "Building and starting backend services..."
echo ""

# shellcheck disable=SC2086
if ! docker compose ${COMPOSE_FILES} up --build -d; then
    warn "Docker services failed to start. Continuing with UI only..."
fi

echo ""
ok "Backend services started"

# ---------------------------------------------------------------------------
# Step 5: Start UI dev server (with hot reload)
# ---------------------------------------------------------------------------

if [ -f "${PROJECT_ROOT}/ui/package.json" ]; then
    info "Starting web UI dev server..."
    (cd "${PROJECT_ROOT}/ui" && npm run dev) &
    UI_PID=$!
    ok "UI dev server starting (PID: ${UI_PID})"
else
    warn "ui/package.json not found — skipping UI dev server"
fi

echo ""

# ---------------------------------------------------------------------------
# Step 6: Stream logs
# ---------------------------------------------------------------------------

echo -e "${CYAN}══════════════════════════════════════════${NC}"
echo -e "  ${GREEN}API Server:${NC}    http://localhost:${API_PORT:-3000}"
echo -e "  ${GREEN}Web UI:${NC}        http://localhost:5173"
echo -e "  ${GREEN}Log Level:${NC}     ${LOG_LEVEL}"
echo -e "  ${GREEN}Repos Path:${NC}    ${REPOS_DIR}"
echo -e "${CYAN}══════════════════════════════════════════${NC}"
echo ""
info "Streaming backend logs... Press Ctrl+C to stop."
echo ""

# Stream Docker Compose logs (blocks until Ctrl+C)
# shellcheck disable=SC2086
docker compose ${COMPOSE_FILES} logs -f --tail=50
