-- name: CreateAdministrador :one
INSERT INTO horas_sociales.administrador (correo, contrasena_hash)
VALUES ($1, $2)
RETURNING *;

-- name: GetAdministradorPorCorreo :one
SELECT * FROM horas_sociales.administrador WHERE correo = $1;

-- name: GetAdministradorPorID :one
SELECT * FROM horas_sociales.administrador WHERE id = $1;

-- name: ListAdministradores :many
SELECT id, correo, activo, creado_en, actualizado_en
FROM horas_sociales.administrador
ORDER BY creado_en, id
LIMIT sqlc.arg(limite)::integer OFFSET sqlc.arg(desplazamiento)::bigint;

-- name: SetAdministradorActivo :one
UPDATE horas_sociales.administrador
SET activo = $2, actualizado_en = now()
WHERE id = $1
RETURNING *;

-- name: CountAdministradores :one
SELECT count(*)::int AS total FROM horas_sociales.administrador;
