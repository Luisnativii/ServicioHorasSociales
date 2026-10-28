#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
for script in scripts/*.sh; do bash -n "$script"; done
bash scripts/dc-test.sh config --quiet
bash scripts/dc-test.sh run --rm --no-deps api bash -c \
  'test -z "$(gofmt -l cmd internal)" && go mod verify && go vet ./... && go test -race ./... && go build -o /tmp/hs-api-check ./cmd/api'
bash scripts/dc-test.sh run --rm --no-deps web npm run build
bash scripts/dc-test.sh exec -T web node --input-type=module -e '
  for (const path of ["/", "/api/health/live", "/api/health/ready"]) {
    const response = await fetch("http://127.0.0.1:8080" + path);
    if (!response.ok) throw new Error(`${path}: ${response.status}`);
    if (path.startsWith("/api") && (await response.json()).status !== "ok") throw new Error(path);
    console.log(`${path}: ${response.status}`);
  }
'
