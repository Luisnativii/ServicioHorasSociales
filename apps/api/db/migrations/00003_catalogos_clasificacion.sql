-- +goose Up
CREATE TABLE horas_sociales.tipo_proyecto (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  nombre text NOT NULL,
  activo boolean NOT NULL DEFAULT true,
  CONSTRAINT tipo_proyecto_nombre_formato CHECK (
    nombre <> '' AND nombre !~ '^[[:space:]]' AND nombre !~ '[[:space:]]$'
  )
);

CREATE UNIQUE INDEX tipo_proyecto_nombre_canonico_unique
  ON horas_sociales.tipo_proyecto (lower(nombre));

CREATE TABLE horas_sociales.area (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  nombre text NOT NULL,
  activo boolean NOT NULL DEFAULT true,
  CONSTRAINT area_nombre_formato CHECK (
    nombre <> '' AND nombre !~ '^[[:space:]]' AND nombre !~ '[[:space:]]$'
  )
);

CREATE UNIQUE INDEX area_nombre_canonico_unique
  ON horas_sociales.area (lower(nombre));

-- +goose Down
-- Los catálogos podrán quedar referenciados por la historia de proyectos.
-- +goose StatementBegin
DO $$ BEGIN
  RAISE EXCEPTION 'Rollback de catálogos de clasificación deshabilitado para conservar referencias';
END $$;
-- +goose StatementEnd
