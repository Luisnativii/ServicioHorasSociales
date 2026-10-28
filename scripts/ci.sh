#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
bash scripts/init-test.sh
cleanup() {
  result=$?
  trap - EXIT
  if [ "$result" -ne 0 ]; then bash scripts/dc-test.sh logs --no-color --tail=80 api web db || true; fi
  bash scripts/dc-test.sh --runtime down >/dev/null || true
  bash scripts/dc-test.sh down >/dev/null || true
  exit "$result"
}
trap cleanup EXIT
bash scripts/check-gitflow.sh
bash scripts/check-test-isolation.sh
bash scripts/dc-test.sh build api web migrator
bash scripts/dc-test.sh up -d --wait --wait-timeout 180 db api web
bash scripts/dc-test.sh run --rm migrator up
bash scripts/check-db-roles.sh
bash scripts/check.sh
bash scripts/check-integration.sh
bash scripts/check-runtime.sh
bash scripts/check-security.sh
