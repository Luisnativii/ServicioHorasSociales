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

-- name: PageTiposProyecto :many
WITH filtrado AS (
    SELECT id, nombre, activo
    FROM horas_sociales.tipo_proyecto
    WHERE (activo OR sqlc.arg(incluir_inactivos)::boolean)
      AND strpos(lower(nombre), lower(sqlc.arg(busqueda)::text)) > 0
)
SELECT pagina.id,
       COALESCE(pagina.nombre, '')::text AS nombre,
       COALESCE(pagina.activo, false)::boolean AS activo,
       total.cantidad
FROM (SELECT count(*) AS cantidad FROM filtrado) AS total
LEFT JOIN (
    SELECT id, nombre, activo
    FROM filtrado
    ORDER BY lower(nombre), id
    LIMIT sqlc.arg(limite)::integer OFFSET sqlc.arg(desplazamiento)::bigint
) AS pagina ON true
ORDER BY lower(pagina.nombre), pagina.id;

-- name: PageAreas :many
WITH filtrado AS (
    SELECT id, nombre, activo
    FROM horas_sociales.area
    WHERE (activo OR sqlc.arg(incluir_inactivos)::boolean)
      AND strpos(lower(nombre), lower(sqlc.arg(busqueda)::text)) > 0
)
SELECT pagina.id,
       COALESCE(pagina.nombre, '')::text AS nombre,
       COALESCE(pagina.activo, false)::boolean AS activo,
       total.cantidad
FROM (SELECT count(*) AS cantidad FROM filtrado) AS total
LEFT JOIN (
    SELECT id, nombre, activo
    FROM filtrado
    ORDER BY lower(nombre), id
    LIMIT sqlc.arg(limite)::integer OFFSET sqlc.arg(desplazamiento)::bigint
) AS pagina ON true
ORDER BY lower(pagina.nombre), pagina.id;
