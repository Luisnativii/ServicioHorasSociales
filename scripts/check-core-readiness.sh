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
SELECT md5(string_agg(to_jsonb(p)::text, ',' ORDER BY id)) FROM public.projects AS p WHERE p.titulo <> 'Proyecto API #50' AND p.titulo NOT LIKE 'Proyecto prueba cierre #53 %' AND p.titulo NOT LIKE 'Proyecto prueba horas #65 %';
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
if [ "$(db_sql <<'SQL'
SELECT to_regclass('horas_sociales.proyecto_clasificacion') IS NOT NULL;
SQL
)" != t ]; then
  {
    printf 'BEGIN;\n'
    sed '/^-- +goose Down/,$d' apps/api/db/migrations/00003_catalogos_clasificacion.sql
    sed '/^-- +goose Down/,$d' apps/api/db/migrations/00005_proyectos.sql
    printf 'COMMIT;\n'
  } | db_sql
fi
db_sql <<'SQL'
GRANT SELECT, UPDATE ON horas_sociales.tipo_proyecto, horas_sociales.area, horas_sociales.modalidad TO hs_core_probe;
GRANT SELECT, INSERT ON horas_sociales.proyecto_clasificacion TO hs_core_probe;
GRANT SELECT, INSERT, UPDATE ON horas_sociales.proyecto_responsable TO hs_core_probe;
INSERT INTO horas_sociales.tipo_proyecto (id,nombre) VALUES ('00000000-0000-4000-8000-000000050001','Tipo sintético #50') ON CONFLICT (id) DO NOTHING;
INSERT INTO horas_sociales.area (id,nombre) VALUES
 ('00000000-0000-4000-8000-000000050002','Área sintética A #50'),
 ('00000000-0000-4000-8000-000000050003','Área sintética B #50') ON CONFLICT (id) DO NOTHING;
INSERT INTO horas_sociales.modalidad (id,nombre) VALUES ('00000000-0000-4000-8000-000000050004','Modalidad sintética #50') ON CONFLICT (id) DO NOTHING;
DELETE FROM horas_sociales.proyecto_responsable WHERE proyecto_id IN (SELECT id FROM public.projects WHERE titulo='Proyecto API #50');
DELETE FROM horas_sociales.proyecto_clasificacion WHERE proyecto_id IN (SELECT id FROM public.projects WHERE titulo='Proyecto API #50');
DELETE FROM public.projects WHERE titulo='Proyecto API #50';
SQL
db_sql < deploy/core-local/proyectos-permisos.sql
db_sql < deploy/core-local/participaciones-permisos.sql
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
INSERT INTO public.projects (id, institution_id, district_id, titulo, descripcion, correlativo_css, atributos, ubicacion)
VALUES ('00000000-0000-4000-8000-000000046003', '00000000-0000-4000-8000-000000046001',
        (SELECT id FROM public.geo_districts WHERE nombre = 'Jujutla'), 'Proyecto sintético', 'Prueba de conservación', '00050',
        '{"area_css":"Área histórica literal"}', 'Ubicación heredada')
ON CONFLICT (id) DO UPDATE SET correlativo_css=EXCLUDED.correlativo_css,
  atributos=EXCLUDED.atributos, ubicacion=EXCLUDED.ubicacion;

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
db_sql <<'SQL'
BEGIN;
INSERT INTO public.projects (id,institution_id,titulo,descripcion,cupos,estado)
VALUES ('00000000-0000-4000-8000-000000059001','00000000-0000-4000-8000-000000046002','Proyecto sintético de cupo #59','Prueba de asignación',1,'ACTIVO'),
       ('00000000-0000-4000-8000-000000059002','00000000-0000-4000-8000-000000046002','Proyecto sintético cerrado #59','Prueba de inscripción histórica',NULL,'FINALIZADO')
ON CONFLICT (id) DO NOTHING;
INSERT INTO public.students (id,carnet,nombres,apellidos,correo_institucional,career_id,activo)
VALUES
 ('00000000-0000-4000-8000-000000059011','00590011','Estudiante','A #59','a59@uca.edu.sv','00000000-0000-4000-8000-000000035002',true),
 ('00000000-0000-4000-8000-000000059012','00590012','Estudiante','B #59','b59@uca.edu.sv','00000000-0000-4000-8000-000000035002',true),
 ('00000000-0000-4000-8000-000000059013','00590013','Estudiante','Inactivo #59','c59@uca.edu.sv','00000000-0000-4000-8000-000000035002',false)
ON CONFLICT (id) DO NOTHING;
DELETE FROM public.project_enrollments WHERE project_id='00000000-0000-4000-8000-000000059001';
INSERT INTO public.project_enrollments(project_id,student_id,estado,asignada_en,cerrada_en)
VALUES ('00000000-0000-4000-8000-000000059002','00000000-0000-4000-8000-000000059011','FINALIZADA',
        '2026-01-02T12:00:00Z','2026-02-02T12:00:00Z')
ON CONFLICT ON CONSTRAINT uq_project_enrollments_student_project DO NOTHING;
DO $check$
BEGIN
  IF has_table_privilege('hs_core_probe','public.project_enrollments','DELETE')
    OR has_column_privilege('hs_core_probe','public.project_enrollments','estado','INSERT')
    OR has_column_privilege('hs_core_probe','public.project_enrollments','asignada_en','INSERT') THEN
    RAISE EXCEPTION 'Permisos de participación exceden el alta';
  END IF;
END
$check$;
COMMIT;
SQL
if [ "$(db_sql <<'SQL'
SELECT to_regclass('horas_sociales.cierre_evento') IS NOT NULL;
SQL
)" != t ]; then
  {
    printf 'BEGIN;\n'
    sed '/^-- +goose Down/,$d' apps/api/db/migrations/00006_cierres.sql
    printf 'COMMIT;\n'
  } | db_sql
fi
db_sql < deploy/core-local/cierres-permisos.sql
db_sql < deploy/core-local/horas-permisos.sql
if [ "$(db_sql <<'SQL'
SELECT to_regclass('horas_sociales.autor_movimiento') IS NOT NULL;
SQL
)" != t ]; then
  {
    printf 'BEGIN;\n'
    sed '/^-- +goose Down/,$d' apps/api/db/migrations/00007_autor_movimiento.sql
    printf 'COMMIT;\n'
  } | db_sql
fi
db_sql <<'SQL'
GRANT SELECT, INSERT ON horas_sociales.autor_movimiento TO hs_core_probe;
SQL
db_sql <<'SQL'
GRANT SELECT, INSERT ON horas_sociales.cierre_evento TO hs_core_probe;
BEGIN;
DELETE FROM horas_sociales.cierre_evento WHERE proyecto_id BETWEEN
  '00000000-0000-4000-8000-000000053001' AND '00000000-0000-4000-8000-000000053004';
DELETE FROM public.hour_movements WHERE enrollment_id IN
  (SELECT id FROM public.project_enrollments WHERE project_id BETWEEN
  '00000000-0000-4000-8000-000000053001' AND '00000000-0000-4000-8000-000000053004');
DELETE FROM public.project_enrollments WHERE project_id BETWEEN
  '00000000-0000-4000-8000-000000053001' AND '00000000-0000-4000-8000-000000053004';
INSERT INTO public.projects(id,institution_id,titulo,descripcion,estado,fecha_inicio)
SELECT ('00000000-0000-4000-8000-00000005300'||n)::uuid,
  '00000000-0000-4000-8000-000000046002','Proyecto prueba cierre #53 '||n,'Datos sintéticos','ACTIVO',current_date-2
FROM generate_series(1,4) n
ON CONFLICT (id) DO UPDATE SET estado='ACTIVO',fecha_inicio=EXCLUDED.fecha_inicio,fecha_cierre=NULL;
INSERT INTO public.project_enrollments(id,project_id,student_id,asignada_en)
SELECT ('00000000-0000-4000-8000-0000000530'||p||n)::uuid,
  ('00000000-0000-4000-8000-00000005300'||p)::uuid,
  ('00000000-0000-4000-8000-00000005901'||n)::uuid, now()-interval '2 days'
FROM generate_series(1,3) p CROSS JOIN generate_series(1,2) n;
INSERT INTO public.hour_movements(enrollment_id,horas,naturaleza,fecha_horas,origen)
VALUES ('00000000-0000-4000-8000-000000053032',2,'EXTERNA',(now() AT TIME ZONE 'America/El_Salvador')::date,'IMPORTACION');
COMMIT;
SQL
db_sql <<'SQL'
BEGIN;
DELETE FROM horas_sociales.cierre_evento WHERE proyecto_id IN
  ('00000000-0000-4000-8000-000000065001','00000000-0000-4000-8000-000000065002');
DELETE FROM horas_sociales.autor_movimiento WHERE movimiento_id IN (
    SELECT id FROM public.hour_movements WHERE enrollment_id IN
      ('00000000-0000-4000-8000-000000065011','00000000-0000-4000-8000-000000065012'));
DELETE FROM public.hour_movements WHERE enrollment_id IN
      ('00000000-0000-4000-8000-000000065011','00000000-0000-4000-8000-000000065012');
INSERT INTO public.projects (id,institution_id,titulo,descripcion,naturaleza,estado)
VALUES ('00000000-0000-4000-8000-000000065001','00000000-0000-4000-8000-000000046002','Proyecto prueba horas #65 1','Datos sintéticos','INTERNA','ACTIVO'),
       ('00000000-0000-4000-8000-000000065002','00000000-0000-4000-8000-000000046002','Proyecto prueba horas #65 2','Datos sintéticos','EXTERNA','ACTIVO')
ON CONFLICT(id) DO UPDATE SET estado='ACTIVO',fecha_cierre=NULL;
INSERT INTO public.project_enrollments(id,project_id,student_id,asignada_en)
VALUES ('00000000-0000-4000-8000-000000065011','00000000-0000-4000-8000-000000065001','00000000-0000-4000-8000-000000059011',now()-interval '1 day'),
       ('00000000-0000-4000-8000-000000065012','00000000-0000-4000-8000-000000065002','00000000-0000-4000-8000-000000059011',now()-interval '1 day')
ON CONFLICT (id) DO UPDATE SET estado='ASIGNADA',cerrada_en=NULL,asignada_en=EXCLUDED.asignada_en;
INSERT INTO public.hour_movements(enrollment_id,horas,naturaleza,fecha_horas,origen)
VALUES ('00000000-0000-4000-8000-000000065011',20,'EXTERNA',(now() AT TIME ZONE 'America/El_Salvador')::date-80,'IMPORTACION');
DO $check$
BEGIN
  IF has_table_privilege('hs_core_probe','public.hour_movements','UPDATE, DELETE')
     OR has_column_privilege('hs_core_probe','public.hour_movements','registrado_por','INSERT') THEN
    RAISE EXCEPTION 'Privilegios del libro demasiado amplios';
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
DO $check$
BEGIN
  IF (SELECT count(*) FROM horas_sociales.proyecto_clasificacion h JOIN public.projects p ON p.id=h.proyecto_id
      WHERE p.titulo='Proyecto API #50') <> 2 THEN
    RAISE EXCEPTION 'No se conservó la anterior clasificación del proyecto';
  END IF;
END
$check$;
DO $check$
BEGIN
  IF (SELECT sum(h.horas) FROM public.hour_movements h
      WHERE h.enrollment_id='00000000-0000-4000-8000-000000053032') IS DISTINCT FROM 2::bigint THEN
    RAISE EXCEPTION 'El cierre alteró las horas existentes';
  END IF;
END
$check$;
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
