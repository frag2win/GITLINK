# =============================================================================
# Stage 1: Builder — compile the git-server binary
# =============================================================================
FROM rust:1.77-slim-bookworm AS builder

# Install build dependencies for libgit2 (git2-rs)
RUN apt-get update && apt-get install -y --no-install-recommends \
    pkg-config \
    libssl-dev \
    cmake \
    libgit2-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /build

# Copy manifests first for dependency caching
COPY Cargo.toml Cargo.lock* ./

# Create a dummy main.rs to pre-build dependencies
RUN mkdir -p src && echo "fn main() {}" > src/main.rs
RUN cargo build --release || true

# Copy actual source code and rebuild
COPY src/ src/
RUN touch src/main.rs && cargo build --release

# =============================================================================
# Stage 2: Runner — minimal runtime image
# =============================================================================
FROM debian:bookworm-slim AS runner

# Install runtime dependencies for libgit2
RUN apt-get update && apt-get install -y --no-install-recommends \
    libgit2-1.5 \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Create directories for repos and socket
RUN mkdir -p /repos /socket

# Copy the compiled binary from builder
COPY --from=builder /build/target/release/git-server /usr/local/bin/git-server

# Run as non-root user (UID 1002)
USER 1002

# No network exposure — communication is via Unix Domain Socket only
# Mount points:
#   /repos  — bare Git repositories storage
#   /socket — Unix domain socket for IPC
VOLUME ["/repos", "/socket"]

ENTRYPOINT ["git-server"]
