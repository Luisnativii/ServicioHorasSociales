#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
dc() { bash scripts/dc-test.sh --core "$@"; }
db_sql() {
  dc exec -T db sh -c 'exec psql -X -qAt -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB"'
}
snapshot() {
  db_sql <<'SQL'
SELECT md5(string_agg(to_jsonb(m)::text, ',' ORDER BY name)) FROM public.schema_migrations AS m;
SELECT md5(string_agg(to_jsonb(f)::text, ',' ORDER BY id)) FROM public.faculties AS f;
SELECT md5(string_agg(to_jsonb(c)::text, ',' ORDER BY id)) FROM public.careers AS c;
SELECT to_regclass('public.goose_db_version') IS NULL;
SQL
}
cleanup() {
  result=$?
  trap - EXIT
  if [ "$result" -ne 0 ]; then dc logs --tail=40 api api-missing db >&2 || true; fi
  dc down >/dev/null || true
  exit "$result"
}
bash scripts/init-test.sh
dc config --quiet
trap cleanup EXIT
dc up -d --wait --wait-timeout 150 db
db_sql <<'SQL'
BEGIN;
SELECT pg_advisory_xact_lock(139);
\getenv password POSTGRES_PASSWORD
SELECT format('CREATE ROLE hs_core_probe LOGIN PASSWORD %L', :'password')
WHERE NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'hs_core_probe') \gexec
GRANT USAGE ON SCHEMA public TO hs_core_probe;
GRANT SELECT ON public.faculties, public.careers TO hs_core_probe;
INSERT INTO public.faculties (id, nombre)
VALUES ('00000000-0000-4000-8000-000000000139', 'Facultad sintética de verificación')
ON CONFLICT (id) DO NOTHING;
DO $check$
BEGIN
  IF (SELECT count(*) = 3 AND bool_and(length(trim(checksum)) = 64) FROM public.schema_migrations) IS NOT TRUE
     OR has_schema_privilege('hs_core_probe', 'public', 'CREATE')
     OR has_table_privilege('hs_core_probe', 'public.careers', 'UPDATE')
     OR has_table_privilege('hs_core_probe', 'public.users', 'SELECT')
     OR to_regclass('public.goose_db_version') IS NOT NULL THEN
    RAISE EXCEPTION 'Inventario o privilegios inesperados en hs-test-core';
  END IF;
END
$check$;
COMMIT;
SQL
before=$(snapshot)
dc build api
dc up -d --no-build --no-deps api api-missing
dc run --rm --no-deps -T probe
[ "$(snapshot)" = "$before" ] || { echo 'La API alteró el catálogo o el registro de migraciones del core.' >&2; exit 1; }
echo 'hs-test-core: se conservaron el catálogo y el registro de migraciones.'
dc stop db
dc run --rm --no-deps -T probe --db-down
dc up -d --wait --wait-timeout 90 db
dc run --rm --no-deps -T probe
[ "$(snapshot)" = "$before" ] || { echo 'El reinicio cambió los datos de prueba.' >&2; exit 1; }
echo 'hs-test-core: PostgreSQL se recuperó y readiness volvió a 200.'
