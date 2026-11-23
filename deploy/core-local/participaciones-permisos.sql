BEGIN;
GRANT SELECT (id, project_id, student_id, estado, asignada_en, cerrada_en),
      INSERT (project_id, student_id) ON public.project_enrollments TO :"app_role";
COMMIT;
