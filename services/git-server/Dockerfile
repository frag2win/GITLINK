# =============================================================================
# Stage 1: Builder — compile the git-server binary
# =============================================================================
FROM rust:slim-bookworm AS builder

# Install build dependencies for libgit2 (git2-rs)
RUN apt-get update && apt-get install -y --no-install-recommends \
    pkg-config \
    libssl-dev \
    cmake \
    libgit2-dev \
    protobuf-compiler \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /build

# Copy proto definitions
COPY proto/ /proto/

# Copy manifests first for dependency caching
COPY services/git-server/Cargo.toml services/git-server/Cargo.lock* ./

# Copy build script so prost can generate protobuf structs
COPY services/git-server/build.rs ./

# Create a dummy main.rs to pre-build dependencies
RUN mkdir -p src && echo "fn main() {}" > src/main.rs
RUN cargo build --release || true

# Copy actual source code and rebuild
COPY services/git-server/src/ src/
RUN touch src/main.rs && touch build.rs && cargo build --release

# =============================================================================
# Stage 2: Runner — minimal runtime image
# =============================================================================
FROM debian:bookworm-slim AS runner

# Install runtime dependencies for libgit2 and git itself
RUN apt-get update && apt-get install -y --no-install-recommends \
    libgit2-1.5 \
    ca-certificates \
    git \
    && rm -rf /var/lib/apt/lists/*

# Create directories for repos and socket
RUN mkdir -p /repos /run/git && chown -R 1002:1002 /repos /run/git

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
