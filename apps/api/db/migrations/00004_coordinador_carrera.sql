-- +goose Up
CREATE TABLE horas_sociales.coordinador_carrera (
  carrera_id uuid PRIMARY KEY CHECK (carrera_id <> '00000000-0000-0000-0000-000000000000'),
  nombre text CHECK (
    nombre IS NULL OR (char_length(nombre) BETWEEN 1 AND 200 AND nombre !~ '^[[:space:]]|[[:space:]]$')
  )
);

-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN
  RAISE EXCEPTION 'Rollback deshabilitado para conservar los nombres de coordinadores';
END $$;
-- +goose StatementEnd
