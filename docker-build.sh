#!/usr/bin/env bash
set -euo pipefail
IMAGE=${IMAGE:-go-url-shortener:latest}
docker build -t "$IMAGE" .
echo "Built $IMAGE"