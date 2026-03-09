#!/usr/bin/env bash
# build.sh — Cross-compile downloader for Linux amd64
# Usage: ./scripts/build.sh

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
OUT_DIR="$REPO_ROOT/build"

mkdir -p "$OUT_DIR"

GOOS=linux GOARCH=amd64 go build -o "$OUT_DIR/downloader-linux-amd64" "$REPO_ROOT/cmd/downloader/"

echo "Built: $OUT_DIR/downloader-linux-amd64"
