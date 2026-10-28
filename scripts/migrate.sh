#!/usr/bin/env bash
# Crear únicamente la base fundacional con la cuenta original. Todo cambio
# posterior corre con el migrator dedicado, también en una instalación vacía.
set -euo pipefail
cd "$(dirname "$0")/.."
case "${1:-}" in
  dev) dc() { bash scripts/dc.sh "$@"; };;
  test) dc() { bash scripts/dc-test.sh "$@"; };;
  *) echo 'Uso: bash scripts/migrate.sh dev|test' >&2; exit 1;;
esac
dc config --quiet
dc build migrator
dc up -d --wait --wait-timeout 90 db
exists=$(dc exec -T db sh -c 'exec psql -X -U "$POSTGRES_USER" -d "$POSTGRES_DB" -Atc "SELECT to_regclass('\''public.goose_db_version'\'') IS NOT NULL;"')
applied=f
if [ "$exists" = t ]; then
  applied=$(dc exec -T db sh -c 'exec psql -X -U "$POSTGRES_USER" -d "$POSTGRES_DB" -Atc "SELECT EXISTS (SELECT FROM public.goose_db_version WHERE version_id = 1 AND is_applied);"')
fi
if [ "$applied" != t ]; then dc run --rm migrator up-to 1; fi
bash scripts/provision-db-roles.sh "$1"
dc run --rm migrator up
