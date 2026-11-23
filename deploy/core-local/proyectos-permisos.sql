-- Permisos para el entorno local del portal; el core compartido conserva su operación propia.
BEGIN;
GRANT SELECT (id,institution_id,district_id,titulo,descripcion,correlativo_css,naturaleza,estado,
              fecha_inicio,fecha_cierre,cupos,atributos,ubicacion),
      INSERT (institution_id,district_id,titulo,descripcion,naturaleza,estado,fecha_inicio,cupos),
      UPDATE (district_id,titulo,descripcion,naturaleza,fecha_inicio,cupos)
  ON public.projects TO :"app_role";
GRANT SELECT (id,project_id) ON public.project_enrollments TO :"app_role";
GRANT SELECT (enrollment_id) ON public.hour_movements TO :"app_role";
COMMIT;
