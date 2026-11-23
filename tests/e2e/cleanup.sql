-- El run numérico identifica exclusivamente los datos creados por Chromium.
BEGIN;
CREATE TEMP TABLE check80_project AS SELECT id FROM public.projects WHERE titulo='Proyecto E2E80 '||:'run';
CREATE TEMP TABLE check80_enrollment AS
    SELECT id FROM public.project_enrollments WHERE project_id IN (SELECT id FROM check80_project);
CREATE TEMP TABLE check80_movement AS
    SELECT id FROM public.hour_movements WHERE enrollment_id IN (SELECT id FROM check80_enrollment);
DELETE FROM horas_sociales.autor_movimiento WHERE movimiento_id IN (SELECT id FROM check80_movement);
DELETE FROM horas_sociales.cierre_evento WHERE proyecto_id IN (SELECT id FROM check80_project);
DELETE FROM public.hour_movements WHERE id IN (SELECT id FROM check80_movement);
DELETE FROM public.project_enrollments WHERE id IN (SELECT id FROM check80_enrollment);
DELETE FROM horas_sociales.proyecto_responsable WHERE proyecto_id IN (SELECT id FROM check80_project);
DELETE FROM horas_sociales.proyecto_clasificacion WHERE proyecto_id IN (SELECT id FROM check80_project);
DELETE FROM public.projects WHERE id IN (SELECT id FROM check80_project);
DELETE FROM public.students WHERE correo_institucional='e2e80-'||:'run'||'@uca.edu.sv';
DELETE FROM public.institutions WHERE nombre='Institución E2E80 '||:'run';
COMMIT;
