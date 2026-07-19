#!/usr/bin/env bash
# =============================================================================
# generate-proto.sh — Compile protobuf schema for Go
# =============================================================================

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo "Generating Go protobuf code..."
cd "$PROJECT_ROOT"
protoc --proto_path=proto --go_out=services/api-server/proto/generated --go_opt=paths=source_relative proto/git_commands.proto

echo "✅ Protobuf code generated successfully."
