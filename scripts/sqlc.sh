#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
case "${1:-generate}" in generate|vet) ;; *) echo 'Uso: bash scripts/sqlc.sh generate|vet' >&2; exit 1;; esac
exec docker run --rm --user "$(id -u):$(id -g)" \
  -v "$PWD/apps/api:/src" -w /src \
  sqlc/sqlc:1.31.1@sha256:70f53171d27b2424e9358869975455a6e955a5aa8e58a998a270a6e34e525537 "${1:-generate}"
