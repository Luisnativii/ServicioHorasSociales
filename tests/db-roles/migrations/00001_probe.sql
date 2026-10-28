-- +goose Up
CREATE TABLE horas_sociales.hs_privilege_probe (
  id integer PRIMARY KEY,
  note text NOT NULL
);

-- +goose Down
DROP TABLE horas_sociales.hs_privilege_probe;
