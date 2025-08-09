#!/usr/bin/env bash
set -euo pipefail
IMAGE=${IMAGE:-go-url-shortener:latest}
PORT=${PORT:-8080}

docker run --rm -it \
  -p "${PORT}:8080" \
  -e PORT=8080 \
  -e BASE_URL="http://localhost:${PORT}" \
  -e DB_PATH="/data/urls.db" \
  -v urls_data:/data \
  "$IMAGE"
