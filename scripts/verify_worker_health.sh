#!/usr/bin/env bash
# Usage: ./scripts/verify_worker_health.sh [host:port]
# Default checks worker-realtime health on localhost:8081 after docker compose up.
set -euo pipefail
TARGET="${1:-http://127.0.0.1:8081/health}"
curl -sf "$TARGET" | grep -q '"status":"ok"'
echo "ok: $TARGET"
