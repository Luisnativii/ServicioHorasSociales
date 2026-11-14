#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
dc() { bash scripts/dc-test.sh "$@"; }
dc config --quiet
source scripts/db-roles-env.sh
check_db_roles_file .secrets/hs-test-db-roles.env
while IFS='=' read -r key value; do
  case "$key" in
    HS_APP_DB_PASSWORD) app_password=$value;;
    HS_MIGRATOR_DB_PASSWORD) migration_password=$value;;
  esac
done < .secrets/hs-test-db-roles.env
unset value
database="hs_catalogs_probe_$(date +%s)_$$"
created=0
umask 077
log=$(mktemp)
cleanup() {
  result=$?
  trap - EXIT
  if [ "$created" = 1 ]; then
    dc exec -T db sh -c 'exec dropdb -U "$POSTGRES_USER" "$1"' sh "$database" >/dev/null
  fi
  rm -f "$log"
  exit "$result"
}
trap cleanup EXIT
# Only a database newly created by this invocation may be removed by cleanup.
dc exec -T db sh -c 'exec createdb -U "$POSTGRES_USER" -O hs_migrator "$1"' sh "$database"
created=1
url="postgres://hs_migrator:$migration_password@db:5432/$database?sslmode=disable&connect_timeout=3"
for attempt in 1 2; do
  if ! dc run --rm -e "GOOSE_DBSTRING=$url" migrator up >"$log" 2>&1; then
    echo 'Falló la migración de catálogos en la base vacía de prueba; no se muestran credenciales.' >&2
    exit 1
  fi
done
version=$(dc exec -T db sh -c 'exec psql -X -U "$POSTGRES_USER" -d "$1" -Atc "SELECT max(version_id) FROM public.goose_db_version WHERE is_applied;"' sh "$database")
[ "$version" = 3 ]
counts=$(dc exec -T db sh -c 'exec psql -X -U "$POSTGRES_USER" -d "$1" -Atc "SELECT (SELECT count(*) FROM horas_sociales.tipo_proyecto), (SELECT count(*) FROM horas_sociales.area);"' sh "$database")
[ "$counts" = '0|0' ]
owners=$(dc exec -T db sh -c 'exec psql -X -U "$POSTGRES_USER" -d "$1" -Atc "SELECT string_agg(DISTINCT pg_get_userbyid(relowner), '\''|'\'') FROM pg_class WHERE oid IN ('\''horas_sociales.tipo_proyecto'\''::regclass, '\''horas_sociales.area'\''::regclass);"' sh "$database")
[ "$owners" = hs_migrator ]
dc exec -T db sh -c 'exec psql -X -q -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$1" -c "GRANT USAGE ON SCHEMA horas_sociales TO hs_app; GRANT SELECT, INSERT, UPDATE ON ALL TABLES IN SCHEMA horas_sociales TO hs_app;"' sh "$database"
app_url="postgres://hs_app:$app_password@db:5432/$database?sslmode=disable&connect_timeout=3"
dc run --rm --no-deps -e HS_TEST_CATALOGS=1 -e "DATABASE_URL=$app_url" api go test -race ./internal/modules/catalogos
unset app_password migration_password app_url
printf 'Catálogos: migración vacía/repetida, sin semillas, propietario migrator, unicidad, estados y repositorio real comprobados.\n'
