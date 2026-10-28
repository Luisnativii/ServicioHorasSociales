#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
bash scripts/dc-test.sh run --rm --no-deps web npm audit --audit-level=low
bash scripts/dc-test.sh run --rm --no-deps -e GOMODCACHE=/tmp/vuln-mod api \
  go run golang.org/x/vuln/cmd/govulncheck@v1.9.0 ./...
