# ============================================================
# libp2p-node  —  Multi-stage Docker build
# ============================================================

# ---- Stage 1: Build ----
FROM golang:alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /build

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /build/libp2p-node ./cmd/node

# ---- Stage 2: Run ----
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata git git-daemon

# Non-root user — no filesystem mounts except identity key
RUN adduser -D -u 1001 p2puser

WORKDIR /app

COPY --from=builder /build/libp2p-node /app/libp2p-node

# Identity key volume mount point
RUN mkdir -p /app/identity /app/queue && chown -R p2puser:p2puser /app

USER 1001

ENTRYPOINT ["/app/libp2p-node"]
