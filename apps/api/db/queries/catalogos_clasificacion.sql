-- name: CreateTipoProyecto :one
INSERT INTO horas_sociales.tipo_proyecto (nombre)
VALUES ($1)
RETURNING *;

-- name: GetTipoProyectoPorID :one
SELECT * FROM horas_sociales.tipo_proyecto WHERE id = $1;

-- name: ListTiposProyecto :many
SELECT *
FROM horas_sociales.tipo_proyecto
WHERE activo OR sqlc.arg(incluir_inactivos)::boolean
ORDER BY lower(nombre), id;

-- name: UpdateTipoProyectoNombre :one
UPDATE horas_sociales.tipo_proyecto
SET nombre = $2
WHERE id = $1
RETURNING *;

-- name: SetTipoProyectoActivo :one
UPDATE horas_sociales.tipo_proyecto
SET activo = $2
WHERE id = $1
RETURNING *;

-- name: CreateArea :one
INSERT INTO horas_sociales.area (nombre)
VALUES ($1)
RETURNING *;

-- name: GetAreaPorID :one
SELECT * FROM horas_sociales.area WHERE id = $1;

-- name: ListAreas :many
SELECT *
FROM horas_sociales.area
WHERE activo OR sqlc.arg(incluir_inactivos)::boolean
ORDER BY lower(nombre), id;

-- name: UpdateAreaNombre :one
UPDATE horas_sociales.area
SET nombre = $2
WHERE id = $1
RETURNING *;

-- name: SetAreaActivo :one
UPDATE horas_sociales.area
SET activo = $2
WHERE id = $1
RETURNING *;
