# ============================================================
# api-server  —  Multi-stage Docker build
# ============================================================

# ---- Stage 1: Build ----
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /build

# Cache dependency downloads
COPY go.mod go.sum ./
RUN go mod download

# Copy source and compile a statically-linked binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /build/api-server ./cmd/server

# ---- Stage 2: Run ----
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

# Non-root user for security
RUN adduser -D -u 1003 appuser

WORKDIR /app

COPY --from=builder /build/api-server /app/api-server

# Default directories for data volumes
RUN mkdir -p /app/data /app/repos && chown -R appuser:appuser /app

USER 1003

EXPOSE 3000

ENTRYPOINT ["/app/api-server"]
