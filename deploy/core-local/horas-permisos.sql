-- Reproduce únicamente las columnas de origen y concepto de la migración 005 del core.
BEGIN;
ALTER TABLE public.hour_movements
    ADD COLUMN IF NOT EXISTS origen text NOT NULL DEFAULT 'IMPORTACION',
    ADD COLUMN IF NOT EXISTS concepto text;
ALTER TABLE public.hour_movements ALTER COLUMN origen DROP DEFAULT;
DO $check$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_constraint WHERE conrelid='public.hour_movements'::regclass AND conname='ck_hour_movements_origen') THEN
        ALTER TABLE public.hour_movements ADD CONSTRAINT ck_hour_movements_origen CHECK (origen IN ('IMPORTACION','MANUAL'));
    END IF;
    IF NOT EXISTS (SELECT FROM pg_constraint WHERE conrelid='public.hour_movements'::regclass AND conname='ck_hour_movements_concepto') THEN
        ALTER TABLE public.hour_movements ADD CONSTRAINT ck_hour_movements_concepto CHECK (origen <> 'MANUAL' OR btrim(coalesce(concepto,'')) <> '');
    END IF;
END
$check$;
GRANT SELECT (id,enrollment_id,horas,naturaleza,fecha_horas,registrado_en,solicitud_id,origen,concepto,descripcion,registrado_por),
      INSERT (enrollment_id,horas,naturaleza,fecha_horas,solicitud_id,origen,concepto)
  ON public.hour_movements TO :"app_role";
COMMIT;
