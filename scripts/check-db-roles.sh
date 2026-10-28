#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
dc() { bash scripts/dc-test.sh "$@"; }
bash scripts/provision-db-roles.sh test
source scripts/db-roles-env.sh
check_db_roles_file .secrets/hs-test-db-roles.env
while IFS='=' read -r key value; do
  case "$key" in HS_APP_DB_PASSWORD) password=$value;; esac
done < .secrets/hs-test-db-roles.env
unset value
app_sql() {
  dc exec -T -e "PGPASSWORD=$password" db sh -c 'exec psql -X -q -w -h 127.0.0.1 -U hs_app -d "$POSTGRES_DB" -v ON_ERROR_STOP=1 -At' <<< "$1"
}
cleanup() {
  result=$?
  trap - EXIT
  dc exec -T db sh -c 'exec psql -X -q -U "$POSTGRES_USER" -d "$POSTGRES_DB" -v ON_ERROR_STOP=1' <<'SQL' >/dev/null
DROP TABLE IF EXISTS horas_sociales.hs_privilege_probe;
DROP TABLE IF EXISTS horas_sociales.hs_role_probe;
DROP TABLE IF EXISTS horas_sociales.hs_forbidden;
DROP TABLE IF EXISTS public.hs_forbidden;
SQL
  exit "$result"
}
trap cleanup EXIT
dc run --rm db-role-probe
dc run --rm db-role-probe
[ "$(app_sql "INSERT INTO horas_sociales.hs_privilege_probe VALUES (1, 'dato-sintetico'); SELECT note FROM horas_sociales.hs_privilege_probe WHERE id = 1;")" = dato-sintetico ]
for statement in \
  'CREATE TABLE horas_sociales.hs_forbidden (id integer);' \
  'CREATE TABLE public.hs_forbidden (id integer);' \
  'ALTER TABLE horas_sociales.hs_privilege_probe ADD COLUMN forbidden integer;' \
  'SELECT * FROM public.goose_db_version;'; do
  if app_sql "$statement" >/dev/null 2>&1; then
    echo 'La cuenta de aplicación obtuvo un privilegio no permitido.' >&2
    exit 1
  fi
done
[ "$(app_sql 'SELECT current_user;')" = hs_app ]
before=$(shasum -a 256 .secrets/hs-test-db-roles.env)
bash scripts/provision-db-roles.sh test
[ "$(shasum -a 256 .secrets/hs-test-db-roles.env)" = "$before" ]
[ "$(app_sql 'SELECT note FROM horas_sociales.hs_privilege_probe WHERE id = 1;')" = dato-sintetico ]
dc up -d --wait --wait-timeout 180 db api web
dc exec -T web node -e 'fetch("http://127.0.0.1:8080/api/health/ready").then(async r=>{if(r.status!==200||(await r.json()).status!=="ok")process.exit(1)}).catch(()=>process.exit(1))'
printf 'hs-test: Goose con rol migrator, DML permitido, DDL/revisión Goose denegados, provisión repetible y API limitada comprobados.\n'
