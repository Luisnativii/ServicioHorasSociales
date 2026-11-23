-- Fixture para los entornos aislados del portal. Reproduce el bloque
-- institucional de las migraciones 008 y 015, sin registrarlas como aplicadas.
BEGIN;
SELECT pg_advisory_xact_lock(48);
ALTER TABLE public.institutions
    ADD COLUMN IF NOT EXISTS contacto_nombre text
        CONSTRAINT ck_institutions_contacto_nombre CHECK (btrim(contacto_nombre) <> ''),
    ADD COLUMN IF NOT EXISTS contacto_email text
        CONSTRAINT ck_institutions_contacto_email CHECK (contacto_email ~* '^[^@\s]+@[^@\s]+\.[^@\s]+$'),
    ADD COLUMN IF NOT EXISTS telefono text
        CONSTRAINT ck_institutions_telefono CHECK (telefono ~ '^[0-9+\-\s]{8,20}$'),
    ADD COLUMN IF NOT EXISTS contacto_extension text
        CONSTRAINT ck_institutions_contacto_extension CHECK (
            contacto_extension IS NULL
            OR (contacto_extension ~ '[^[:space:]]' AND telefono IS NOT NULL)
        );

GRANT SELECT (id, nombre, codigo_css, activo, district_id, ubicacion,
              contacto_nombre, contacto_email, telefono, contacto_extension),
      INSERT (nombre, district_id, contacto_nombre, contacto_email, telefono, contacto_extension),
      UPDATE (activo, nombre, district_id, contacto_nombre, contacto_email, telefono, contacto_extension)
    ON public.institutions TO :"app_role";
GRANT INSERT (tabla, registro_id, operacion, datos_antes, datos_despues, actor)
    ON public.auditoria TO :"app_role";
COMMIT;
