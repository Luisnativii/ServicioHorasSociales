#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
APP_VERSION=$(bash scripts/version.sh)
export APP_VERSION
exec docker compose -p hs-dev -f compose.yaml -f compose.dev.yaml "$@"
