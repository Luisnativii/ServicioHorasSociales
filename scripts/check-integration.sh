#!/usr/bin/env bash
# Prueba local con interrupción temporal de hs-dev; conserva el volumen.
set -euo pipefail
cd "$(dirname "$0")/.."
dc() { bash scripts/dc.sh "$@"; }
sql() { dc exec -T db sh -c 'exec psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB" -At' <<< "$1"; }
probe() {
  dc exec -T web node --input-type=module - "$1" "$2" <<'JS'
const [route, expected] = process.argv.slice(2);
const end = Date.now() + 30000;
while (true) {
  try {
    const response = await fetch(`http://127.0.0.1:8080/api/health/${route}`, { signal: AbortSignal.timeout(3000) });
    const body = await response.json();
    if (response.status === Number(expected) && body.status === (expected === '200' ? 'ok' : 'unavailable')) {
      console.log(`${route}: ${response.status} ${body.status}`);
      break;
    }
  } catch {}
  if (Date.now() >= end) throw new Error(`${route} no alcanzó HTTP ${expected}`);
  await new Promise(resolve => setTimeout(resolve, 500));
}
JS
}
validation_schema="hs_validation_$(date +%s)_$$"
created=0
cleanup() {
  result=$?
  trap - EXIT
  dc up -d --wait --wait-timeout 180 db api web >/dev/null || true
  if [ "$created" = 1 ]; then sql "DROP SCHEMA IF EXISTS $validation_schema CASCADE;" >/dev/null || true; fi
  exit "$result"
}
trap cleanup EXIT

dc up -d --wait --wait-timeout 180 db api web
dc --profile tools run --rm migrator up
probe live 200
probe ready 200

dc stop db
probe live 200
probe ready 503
dc up -d --wait --wait-timeout 90 db
probe ready 200

volume_created=$(docker volume inspect hs-dev_pgdata --format '{{.CreatedAt}}')
sql "CREATE SCHEMA $validation_schema; CREATE TABLE $validation_schema.probe (value text PRIMARY KEY); INSERT INTO $validation_schema.probe VALUES ('persistencia-verificada');" >/dev/null
created=1
dc down
dc up -d --wait --wait-timeout 180
[ "$(sql "SELECT value FROM $validation_schema.probe;")" = 'persistencia-verificada' ]
[ "$(docker volume inspect hs-dev_pgdata --format '{{.CreatedAt}}')" = "$volume_created" ]
[ "$(sql "SELECT count(*) FROM public.goose_db_version WHERE version_id = 1 AND is_applied;")" = '1' ]
[ "$(sql "SELECT count(*) FROM information_schema.schemata WHERE schema_name = 'horas_sociales';")" = '1' ]
probe live 200
probe ready 200
printf 'Recuperación, migración y persistencia verificadas sin eliminar el volumen.\n'
