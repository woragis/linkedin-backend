#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"

cd "$ROOT/server"
go test ./...
go test -tags=integration ./integration/...
go test -tags=e2e ./e2e/...

cd "$ROOT/worker"
python -m pytest tests/ -q
