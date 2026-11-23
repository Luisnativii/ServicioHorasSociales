BEGIN;
-- Adaptación del contrato local. El operador del core gestiona la migración compartida.
DO $check$
BEGIN
    IF NOT EXISTS (
        SELECT FROM pg_constraint
        WHERE conrelid='public.projects'::regclass AND conname='ck_projects_estado'
          AND strpos(pg_get_constraintdef(oid), 'ABANDONADO') > 0
    ) THEN
        ALTER TABLE public.projects DROP CONSTRAINT ck_projects_estado;
        ALTER TABLE public.projects ADD CONSTRAINT ck_projects_estado
            CHECK (estado IN ('BORRADOR','ACTIVO','PAUSADO','FINALIZADO','CANCELADO','ABANDONADO'));
    END IF;
END
$check$;
GRANT UPDATE (estado, fecha_cierre) ON public.projects TO :"app_role";
GRANT UPDATE (estado, cerrada_en) ON public.project_enrollments TO :"app_role";
GRANT SELECT (fecha_horas) ON public.hour_movements TO :"app_role";
COMMIT;
