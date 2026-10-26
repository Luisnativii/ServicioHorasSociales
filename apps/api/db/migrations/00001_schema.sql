-- +goose Up
CREATE SCHEMA IF NOT EXISTS horas_sociales;

-- +goose Down
-- Esta migración fundacional solo admite avance; no simular un rollback.
-- +goose StatementBegin
DO $$ BEGIN
  RAISE EXCEPTION 'Rollback del esquema fundacional deshabilitado para conservar datos';
END $$;
-- +goose StatementEnd
