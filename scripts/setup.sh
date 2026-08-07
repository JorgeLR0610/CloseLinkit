#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
cd "$PROJECT_ROOT"

if ! command -v docker >/dev/null 2>&1; then
    echo "Docker is not installed." >&2
    exit 1
fi

if ! docker info >/dev/null 2>&1; then
    echo "Docker daemon is not running. Please start Docker and try again." >&2
    exit 1
fi

cp .env.example .env

echo "Starting containers..."

docker compose up -d

echo "Running database migrations..."

make migrate-up

echo "All set! You can try the app at http://localhost:5173"
