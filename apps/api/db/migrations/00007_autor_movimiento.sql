-- +goose Up
CREATE TABLE horas_sociales.autor_movimiento (
    movimiento_id uuid PRIMARY KEY,
    administrador_id uuid NOT NULL
);

-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN
    RAISE EXCEPTION 'Rollback de atribución de horas deshabilitado';
END $$;
-- +goose StatementEnd
