#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
docker run --rm --user "$(id -u):$(id -g)" -v "$PWD:/repo:ro" -w /repo \
  node:22.23.4-bookworm-slim node --test tests/gitflow.test.cjs
