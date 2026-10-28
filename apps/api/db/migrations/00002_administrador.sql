-- +goose Up
CREATE TABLE horas_sociales.administrador (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  correo text NOT NULL CONSTRAINT administrador_correo_unique UNIQUE,
  contrasena_hash text NOT NULL,
  activo boolean NOT NULL DEFAULT true,
  creado_en timestamptz NOT NULL DEFAULT now(),
  actualizado_en timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT administrador_correo_canonico CHECK (correo <> '' AND correo = lower(btrim(correo))),
  CONSTRAINT administrador_hash_formato CHECK (contrasena_hash LIKE '$argon2id$%')
);

-- +goose Down
-- Conservar identidades que podrán estar referenciadas por la historia.
-- +goose StatementBegin
DO $$ BEGIN
  RAISE EXCEPTION 'Rollback de identidad deshabilitado para conservar cuentas';
END $$;
-- +goose StatementEnd
