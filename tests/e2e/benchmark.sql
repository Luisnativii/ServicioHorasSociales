-- Semilla 80-v1. Solo se ejecuta sobre el core aislado de pruebas.
BEGIN;
INSERT INTO public.projects(id,institution_id,titulo,descripcion,naturaleza,estado)
SELECT md5('hs-bench80-project-'||n)::uuid,'00000000-0000-4000-8000-000000046002',
   'Bench80 proyecto '||n,'Carga sintética de referencia #80','EXTERNA','ACTIVO'
FROM generate_series(1,1000) n ON CONFLICT(id) DO NOTHING;
INSERT INTO public.students(id,carnet,nombres,apellidos,correo_institucional,career_id)
SELECT md5('hs-bench80-student-'||n)::uuid,lpad((80000000+n)::text,8,'0'),
   'Bench80','Estudiante '||n,'bench80-'||n||'@uca.edu.sv','00000000-0000-4000-8000-000000035002'
FROM generate_series(1,10000) n ON CONFLICT(id) DO NOTHING;
INSERT INTO public.project_enrollments(id,project_id,student_id,asignada_en)
SELECT md5('hs-bench80-enrollment-'||n)::uuid,
   md5('hs-bench80-project-'||((n-1)%1000+1))::uuid,
   md5('hs-bench80-student-'||n)::uuid,'2024-01-01T12:00:00Z'
FROM generate_series(1,10000) n ON CONFLICT(id) DO NOTHING;
INSERT INTO public.hour_movements(id,enrollment_id,horas,naturaleza,fecha_horas,origen)
SELECT md5('hs-bench80-hours-'||student||'-'||entry)::uuid,
   md5('hs-bench80-enrollment-'||student)::uuid, 1+(entry%8),
   CASE WHEN entry%2=1 THEN 'EXTERNA' ELSE 'INTERNA' END,
   date '2024-02-01' + ((student-1)*20+entry-1)%40, 'IMPORTACION'
FROM generate_series(1,10000) student CROSS JOIN generate_series(1,20) entry
ON CONFLICT(id) DO NOTHING;
COMMIT;
ANALYZE public.students, public.projects, public.project_enrollments, public.hour_movements;
SELECT 'Bench80 estudiantes='||count(*) FROM public.students WHERE nombres='Bench80';
SELECT 'Bench80 proyectos='||count(*) FROM public.projects WHERE titulo LIKE 'Bench80 proyecto %';
SELECT 'Bench80 movimientos='||count(*) FROM public.hour_movements h
JOIN public.project_enrollments e ON e.id=h.enrollment_id JOIN public.students s ON s.id=e.student_id WHERE s.nombres='Bench80';
SELECT 'Bench80 bytes de la base='||pg_database_size(current_database());
