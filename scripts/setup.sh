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

if [ ! -f .env ]; then
    cp .env.example .env
fi

if grep -q "your_jwt_secret_key_here_change_in_production" .env; then
    NEW_SECRET=$(openssl rand -base64 32)
    sed -i.bak "s|JWT_SECRET=.*|JWT_SECRET=${NEW_SECRET}|" .env && rm -f .env.bak
    echo "Generated fresh local JWT_SECRET in .env"
fi

echo "Starting containers..."

docker compose up -d

echo "Running database migrations..."

make migrate-up

echo "All set! You can try the app at http://localhost:5173"
