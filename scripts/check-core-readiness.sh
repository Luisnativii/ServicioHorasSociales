#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
dc() { bash scripts/dc-test.sh --core "$@"; }
db_sql() {
  dc exec -T db sh -c 'exec psql -X -qAt -v ON_ERROR_STOP=1 -v app_role=hs_core_probe -U "$POSTGRES_USER" -d "$POSTGRES_DB"'
}
snapshot() {
  db_sql <<'SQL'
SELECT md5(string_agg(to_jsonb(m)::text, ',' ORDER BY name)) FROM public.schema_migrations AS m;
SELECT md5(string_agg(to_jsonb(f)::text, ',' ORDER BY id)) FROM public.faculties AS f;
SELECT md5(string_agg(to_jsonb(c)::text, ',' ORDER BY id)) FROM public.careers AS c;
SELECT md5(string_agg(to_jsonb(d)::text, ',' ORDER BY id)) FROM public.geo_departments AS d;
SELECT md5(string_agg(to_jsonb(m)::text, ',' ORDER BY id)) FROM public.geo_municipalities AS m;
SELECT md5(string_agg(to_jsonb(d)::text, ',' ORDER BY id)) FROM public.geo_districts AS d;
SELECT md5(string_agg(to_jsonb(u)::text, ',' ORDER BY id)) FROM public.users AS u;
SELECT md5(string_agg(to_jsonb(p)::text, ',' ORDER BY id)) FROM public.projects AS p;
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
GRANT SELECT ON public.faculties, public.careers,
  public.geo_departments, public.geo_municipalities, public.geo_districts TO hs_core_probe;
INSERT INTO public.faculties (id, nombre)
VALUES ('00000000-0000-4000-8000-000000000139', 'Facultad sintética de verificación')
ON CONFLICT (id) DO NOTHING;
INSERT INTO public.faculties (id, nombre, activo) VALUES
  ('00000000-0000-4000-8000-000000035001', 'Facultad Académica A', true),
  ('00000000-0000-4000-8000-000000035003', 'Facultad Académica B', false)
ON CONFLICT (id) DO NOTHING;
INSERT INTO public.careers (id, faculty_id, nombre, activo) VALUES
  ('00000000-0000-4000-8000-000000035002', '00000000-0000-4000-8000-000000035001', 'Carrera Académica A', true),
  ('00000000-0000-4000-8000-000000035004', '00000000-0000-4000-8000-000000035003', 'Carrera Académica B', true),
  ('00000000-0000-4000-8000-000000035005', '00000000-0000-4000-8000-000000035001', 'Carrera Académica C', false)
ON CONFLICT (id) DO NOTHING;
DO $check$
BEGIN
  IF (SELECT count(*) = 3 AND bool_and(length(trim(checksum)) = 64) FROM public.schema_migrations) IS NOT TRUE
     OR has_schema_privilege('hs_core_probe', 'public', 'CREATE')
     OR has_table_privilege('hs_core_probe', 'public.faculties', 'INSERT, UPDATE, DELETE')
     OR has_table_privilege('hs_core_probe', 'public.careers', 'INSERT, UPDATE, DELETE')
     OR has_table_privilege('hs_core_probe', 'public.geo_departments', 'INSERT, UPDATE, DELETE, TRUNCATE')
     OR has_table_privilege('hs_core_probe', 'public.geo_municipalities', 'INSERT, UPDATE, DELETE, TRUNCATE')
     OR has_table_privilege('hs_core_probe', 'public.geo_districts', 'INSERT, UPDATE, DELETE, TRUNCATE')
     OR has_table_privilege('hs_core_probe', 'public.users', 'SELECT, INSERT, UPDATE, DELETE')
     OR to_regclass('public.goose_db_version') IS NOT NULL THEN
    RAISE EXCEPTION 'Inventario o privilegios inesperados en hs-test-core';
  END IF;
END
$check$;
COMMIT;
SQL
db_sql < deploy/core-local/territory/load.sql
# Probar el SQL de la migración del portal en el core aislado. El arranque
# local aplica la misma migración con Goose y su propia cuenta de migrador.
if [ "$(db_sql <<'SQL'
SELECT to_regclass('horas_sociales.coordinador_carrera') IS NOT NULL;
SQL
)" != t ]; then
  {
    printf 'BEGIN;\nCREATE SCHEMA IF NOT EXISTS horas_sociales;\n'
    sed '/^-- +goose Down/,$d' apps/api/db/migrations/00004_coordinador_carrera.sql
    printf 'COMMIT;\n'
  } | db_sql
fi
db_sql <<'SQL'
GRANT USAGE ON SCHEMA horas_sociales TO hs_core_probe;
GRANT SELECT, INSERT, UPDATE ON horas_sociales.coordinador_carrera TO hs_core_probe;
SQL
db_sql < deploy/core-local/contacto-institucional.sql
db_sql < deploy/core-local/estudiantes-vigencias.sql
db_sql <<'SQL'
BEGIN;
INSERT INTO public.careers (id, faculty_id, nombre, activo)
VALUES ('00000000-0000-4000-8000-000000055002', '00000000-0000-4000-8000-000000035001', 'Programa sintético de expedientes', true)
ON CONFLICT (id) DO UPDATE SET nombre = EXCLUDED.nombre
WHERE careers.nombre IS DISTINCT FROM EXCLUDED.nombre;
DELETE FROM public.students WHERE carnet IN ('00550001', '00550002');
-- Un expediente anterior a la captura de vigencias tiene inicio desconocido.
ALTER TABLE public.students DISABLE TRIGGER trg_student_career_periods;
INSERT INTO public.students (carnet, nombre_completo, correo_institucional, career_id)
VALUES ('00550001', 'ARCHIVO, ESTUDIANTE', 'archivo55@uca.edu.sv', '00000000-0000-4000-8000-000000035002');
ALTER TABLE public.students ENABLE TRIGGER trg_student_career_periods;
DO $check$
BEGIN
  IF has_column_privilege('hs_core_probe', 'public.students', 'dui_hash', 'SELECT, UPDATE')
     OR has_column_privilege('hs_core_probe', 'public.students', 'dui_cifrado', 'SELECT, UPDATE')
     OR has_table_privilege('hs_core_probe', 'public.student_career_periods', 'INSERT, UPDATE, DELETE') THEN
    RAISE EXCEPTION 'Privilegios inesperados para expedientes';
  END IF;
END
$check$;
COMMIT;
SQL
db_sql <<'SQL'
BEGIN;
DELETE FROM public.institutions
WHERE nombre IN ('Institución API prueba', 'Institución API editada', 'Institución API sin sede');
INSERT INTO public.institutions (id, nombre, codigo_css, district_id)
VALUES ('00000000-0000-4000-8000-000000046001', 'Institución sintética con sede', '00046',
        (SELECT id FROM public.geo_districts WHERE nombre = 'San Francisco Menéndez'))
ON CONFLICT (id) DO UPDATE SET activo = true WHERE NOT institutions.activo;
INSERT INTO public.institutions (id, nombre, ubicacion)
VALUES ('00000000-0000-4000-8000-000000046002', 'Institución sintética histórica', 'Descripción histórica de sede')
ON CONFLICT (id) DO NOTHING;
INSERT INTO public.projects (id, institution_id, district_id, titulo, descripcion)
VALUES ('00000000-0000-4000-8000-000000046003', '00000000-0000-4000-8000-000000046001',
        (SELECT id FROM public.geo_districts WHERE nombre = 'Jujutla'), 'Proyecto sintético', 'Prueba de conservación')
ON CONFLICT (id) DO NOTHING;

DO $check$
BEGIN
  IF NOT has_column_privilege('hs_core_probe', 'public.institutions', 'activo', 'UPDATE')
     OR has_table_privilege('hs_core_probe', 'public.institutions', 'SELECT, INSERT, UPDATE, DELETE, TRUNCATE')
     OR NOT has_column_privilege('hs_core_probe', 'public.institutions', 'nombre', 'INSERT')
     OR NOT has_column_privilege('hs_core_probe', 'public.institutions', 'nombre', 'UPDATE')
     OR NOT has_column_privilege('hs_core_probe', 'public.institutions', 'district_id', 'INSERT')
     OR NOT has_column_privilege('hs_core_probe', 'public.institutions', 'district_id', 'UPDATE')
     OR has_column_privilege('hs_core_probe', 'public.institutions', 'codigo_css', 'INSERT, UPDATE')
     OR has_column_privilege('hs_core_probe', 'public.institutions', 'ubicacion', 'INSERT, UPDATE')
     OR has_table_privilege('hs_core_probe', 'public.projects', 'INSERT, UPDATE, DELETE')
     OR has_table_privilege('hs_core_probe', 'public.auditoria', 'SELECT, UPDATE, DELETE') THEN
    RAISE EXCEPTION 'Privilegios de instituciones inesperados en hs-test-core';
  END IF;
END
$check$;
COMMIT;
SQL
institution_audit_count() {
  db_sql <<'SQL'
SELECT count(*) FROM public.auditoria
WHERE tabla = 'institutions' AND registro_id = '00000000-0000-4000-8000-000000046001'
  AND operacion = 'UPDATE' AND actor = 'admin:00000000-0000-4000-8000-000000000046'
  AND datos_antes->>'activo' = 'true' AND datos_despues->>'activo' = 'false';
SQL
}
audit_before=$(institution_audit_count)
before=$(snapshot)
dc build api
dc up -d --no-build --no-deps api api-missing
dc run --rm --no-deps -T probe
dc run --rm --no-deps --build -T catalogo-test
db_sql <<'SQL'
BEGIN;
SELECT set_config('app.actor', 'core:synthetic', true);
UPDATE public.students SET career_id = '00000000-0000-4000-8000-000000035002'
WHERE carnet = '00550001';
DO $check$
BEGIN
  IF (SELECT count(*) FROM public.student_career_periods h JOIN public.students s ON s.id = h.student_id
      WHERE s.carnet = '00550001') <> 3
     OR NOT EXISTS (SELECT 1 FROM public.student_career_periods h JOIN public.students s ON s.id = h.student_id
      WHERE s.carnet = '00550001' AND h.valid_until IS NULL AND h.actor = 'core:synthetic' AND h.career_id = s.career_id) THEN
    RAISE EXCEPTION 'No se conservó el historial al escribir desde el core';
  END IF;
END
$check$;
ROLLBACK;
SQL
[ "$(institution_audit_count)" -eq "$((audit_before + 1))" ]
institution_result=$(db_sql <<'SQL'
SELECT NOT i.activo AND i.codigo_css = '00046'
       AND sede.nombre = 'San Francisco Menéndez' AND ejecucion.nombre = 'Jujutla'
FROM public.institutions AS i
JOIN public.projects AS p ON p.institution_id = i.id
JOIN public.geo_districts AS sede ON sede.id = i.district_id
JOIN public.geo_districts AS ejecucion ON ejecucion.id = p.district_id
WHERE i.id = '00000000-0000-4000-8000-000000046001' AND p.id = '00000000-0000-4000-8000-000000046003';
SQL
)
[ "$institution_result" = t ]
echo 'hs-test-core: institución inactiva con auditoría única y proyecto vinculado intacto.'
[ "$(snapshot)" = "$before" ] || { echo 'La API alteró el catálogo o el registro de migraciones del core.' >&2; exit 1; }
echo 'hs-test-core: se conservaron el catálogo y el registro de migraciones.'
dc stop db
dc run --rm --no-deps -T probe --db-down
dc up -d --wait --wait-timeout 90 db
dc run --rm --no-deps -T probe
[ "$(snapshot)" = "$before" ] || { echo 'El reinicio cambió los datos de prueba.' >&2; exit 1; }
echo 'hs-test-core: PostgreSQL se recuperó y readiness volvió a 200.'
