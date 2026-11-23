-- +goose Up
CREATE TABLE horas_sociales.modalidad (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  nombre text NOT NULL CHECK (nombre <> '' AND nombre !~ '^[[:space:]]' AND nombre !~ '[[:space:]]$'),
  activo boolean NOT NULL DEFAULT true
);
CREATE UNIQUE INDEX modalidad_nombre_canonico_unique ON horas_sociales.modalidad (lower(nombre));

CREATE TABLE horas_sociales.proyecto_clasificacion (
  proyecto_id uuid NOT NULL,
  vigente_desde timestamptz NOT NULL DEFAULT clock_timestamp(),
  tipo_id uuid NOT NULL REFERENCES horas_sociales.tipo_proyecto(id),
  area_id uuid NOT NULL REFERENCES horas_sociales.area(id),
  modalidad_id uuid NOT NULL REFERENCES horas_sociales.modalidad(id),
  PRIMARY KEY (proyecto_id, vigente_desde)
);

CREATE TABLE horas_sociales.proyecto_responsable (
  proyecto_id uuid PRIMARY KEY,
  nombre text NOT NULL CHECK (btrim(nombre) <> ''),
  correo text,
  telefono text,
  extension text CHECK (extension IS NULL OR telefono IS NOT NULL)
);

-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN
  RAISE EXCEPTION 'Rollback de proyectos deshabilitado para conservar historia';
END $$;
-- +goose StatementEnd
