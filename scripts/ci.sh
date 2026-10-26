#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
cleanup() {
  result=$?
  trap - EXIT
  if [ "$result" -ne 0 ]; then bash scripts/dc.sh logs --no-color --tail=80 api web db || true; fi
  bash scripts/dc.sh down >/dev/null || true
  exit "$result"
}
trap cleanup EXIT
bash scripts/init-local.sh
bash scripts/check-gitflow.sh
bash scripts/dc.sh --profile tools build api web migrator
bash scripts/dc.sh up -d --wait --wait-timeout 180
bash scripts/dc.sh --profile tools run --rm migrator up
bash scripts/check.sh
bash scripts/check-integration.sh
bash scripts/check-runtime.sh
bash scripts/check-security.sh
