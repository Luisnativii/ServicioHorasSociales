#!/usr/bin/env bash
# Prueba local sin puertos publicados ni despliegue externo.
set -euo pipefail
cd "$(dirname "$0")/.."
runtime() { bash scripts/dc-test.sh --runtime "$@"; }
runtime config --quiet
cleanup() { result=$?; trap - EXIT; runtime down >/dev/null; exit "$result"; }
trap cleanup EXIT
runtime build api web
runtime up -d --wait --wait-timeout 90
for service in api web; do
  container=$(runtime ps -q "$service")
  user=$(docker inspect --format '{{.Config.User}}' "$container")
  case "$user" in ''|0|0:*|root|root:*) echo "$service se ejecuta como root" >&2; exit 1;; esac
  printf '%s: usuario %s\n' "$service" "$user"
done
runtime exec -T web sh -eu <<'SH'
wget -q -O /tmp/index.html http://127.0.0.1:8080/
grep -q '/assets/' /tmp/index.html
for route in live ready; do
  body=$(wget -q -O - "http://127.0.0.1:8080/api/health/$route")
  test "$body" = '{"status":"ok"}'
  printf '%s: %s\n' "$route" "$body"
done
SH
printf 'Runtime API y Nginx comprobados sin root y sin publicar puertos.\n'
