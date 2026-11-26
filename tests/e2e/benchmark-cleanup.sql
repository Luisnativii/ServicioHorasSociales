-- Remove only synthetic rows created by benchmark.sql in hs-test-core.
BEGIN;
DELETE FROM public.hour_movements WHERE id IN (
  SELECT md5('hs-bench80-hours-'||student||'-'||entry)::uuid
  FROM generate_series(1,10000) student CROSS JOIN generate_series(1,20) entry
);
DELETE FROM public.project_enrollments WHERE id IN (
  SELECT md5('hs-bench80-enrollment-'||n)::uuid FROM generate_series(1,10000) n
);
DELETE FROM public.students WHERE id IN (
  SELECT md5('hs-bench80-student-'||n)::uuid FROM generate_series(1,10000) n
);
DELETE FROM public.projects WHERE id IN (
  SELECT md5('hs-bench80-project-'||n)::uuid FROM generate_series(1,1000) n
);
COMMIT;
