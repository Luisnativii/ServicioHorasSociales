#!/usr/bin/env bash
set -euo pipefail
child=''
stop() {
  if [ -n "$child" ]; then
    kill -TERM "$child" 2>/dev/null || true
    wait "$child" 2>/dev/null || true
    child=''
  fi
}
trap 'stop; exit 0' TERM INT
snapshot() {
  find cmd internal -type f -name '*.go' -print0 | sort -z | xargs -0 sha256sum
  sha256sum go.mod go.sum
}
previous=''
while true; do
  current=$(snapshot | sha256sum)
  if [ "$current" != "$previous" ]; then
    previous="$current"
    # Conservar el binario activo si una edición no compila.
    if go build -o /tmp/hs-api-next ./cmd/api; then
      stop
      mv /tmp/hs-api-next /tmp/hs-api
      /tmp/hs-api &
      child=$!
    fi
  fi
  if [ -n "$child" ] && ! kill -0 "$child" 2>/dev/null; then
    wait "$child" || exit "$?"
    exit 1
  fi
  # Intervalo del observador; la disponibilidad se comprueba con healthchecks.
  sleep 1 &
  wait "$!" || true
done
