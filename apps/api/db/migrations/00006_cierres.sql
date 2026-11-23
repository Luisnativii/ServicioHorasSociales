-- +goose Up
CREATE TABLE horas_sociales.cierre_evento (
    proyecto_id uuid NOT NULL,
    participacion_id uuid,
    estado text NOT NULL CHECK (estado IN ('RETIRADA','ABANDONADO','CANCELADO','CANCELADA','FINALIZADO','FINALIZADA')),
    fecha_cierre date NOT NULL,
    cerrada_en timestamptz NOT NULL,
    actor_id uuid NOT NULL,
    motivo text,
    UNIQUE NULLS NOT DISTINCT (proyecto_id, participacion_id),
    CHECK ((participacion_id IS NULL) = (estado IN ('ABANDONADO','CANCELADO','FINALIZADO')))
);
CREATE UNIQUE INDEX cierre_evento_participacion ON horas_sociales.cierre_evento(participacion_id)
    WHERE participacion_id IS NOT NULL;

-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN
    RAISE EXCEPTION 'Rollback del historial de cierres deshabilitado';
END $$;
-- +goose StatementEnd
