-- name: FindFacultad :one
SELECT id, nombre, activo FROM public.faculties WHERE id = $1;

-- name: FindCarrera :one
SELECT c.id, c.faculty_id, c.nombre, c.activo, f.activo AS facultad_activa
FROM public.careers AS c
JOIN public.faculties AS f ON f.id = c.faculty_id
WHERE c.id = $1;

-- name: ListFacultades :many
SELECT id, nombre, activo
FROM public.faculties
WHERE (sqlc.arg(incluir_inactivas)::boolean OR activo)
  AND strpos(nombre_normalizado, public.fn_normalize_text(sqlc.arg(busqueda)::text)) > 0
ORDER BY nombre_normalizado, id
LIMIT sqlc.arg(limite)::integer OFFSET sqlc.arg(desplazamiento)::bigint;

-- name: CountFacultades :one
SELECT count(*)
FROM public.faculties
WHERE (sqlc.arg(incluir_inactivas)::boolean OR activo)
  AND strpos(nombre_normalizado, public.fn_normalize_text(sqlc.arg(busqueda)::text)) > 0;

-- name: ListCarreras :many
SELECT c.id, c.faculty_id, c.nombre, c.activo, f.activo AS facultad_activa
FROM public.careers AS c
JOIN public.faculties AS f ON f.id = c.faculty_id
WHERE (sqlc.arg(incluir_inactivas)::boolean OR (c.activo AND f.activo))
  AND (sqlc.narg(facultad_id)::uuid IS NULL OR c.faculty_id = sqlc.narg(facultad_id)::uuid)
  AND strpos(c.nombre_normalizado, public.fn_normalize_text(sqlc.arg(busqueda)::text)) > 0
ORDER BY c.nombre_normalizado, c.id
LIMIT sqlc.arg(limite)::integer OFFSET sqlc.arg(desplazamiento)::bigint;

-- name: CountCarreras :one
SELECT count(*)
FROM public.careers AS c
JOIN public.faculties AS f ON f.id = c.faculty_id
WHERE (sqlc.arg(incluir_inactivas)::boolean OR (c.activo AND f.activo))
  AND (sqlc.narg(facultad_id)::uuid IS NULL OR c.faculty_id = sqlc.narg(facultad_id)::uuid)
  AND strpos(c.nombre_normalizado, public.fn_normalize_text(sqlc.arg(busqueda)::text)) > 0;
