#!/usr/bin/env sh
set -eu

cd "$(dirname "$0")"

echo "Building tf2-server-tui..."
go build -trimpath -ldflags="-s -w" -o tf2-server-tui .

echo
echo "Build complete: tf2-server-tui"
