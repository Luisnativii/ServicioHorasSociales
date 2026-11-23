package instituciones

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type DatosInstitucion struct {
	Nombre     string
	DistritoID pgtype.UUID
}

type Filtros struct {
	Busqueda         string
	IncluirInactivas bool
	Limite           int32
	Desplazamiento   int64
}

func (r *Repository) Create(ctx context.Context, actor pgtype.UUID, datos DatosInstitucion) (Institucion, error) {
	return r.write(ctx, actor, func(tx pgx.Tx) (pgtype.UUID, error) {
		var id pgtype.UUID
		err := tx.QueryRow(ctx, `
			INSERT INTO public.institutions (nombre, district_id)
			VALUES ($1, $2) RETURNING id`, datos.Nombre, datos.DistritoID).Scan(&id)
		return id, err
	})
}

func (r *Repository) Update(ctx context.Context, id, actor pgtype.UUID, datos DatosInstitucion) (Institucion, error) {
	if !validID(id) {
		return Institucion{}, ErrInvalidID
	}
	return r.write(ctx, actor, func(tx pgx.Tx) (pgtype.UUID, error) {
		var updated pgtype.UUID
		err := tx.QueryRow(ctx, `
			UPDATE public.institutions SET nombre = $2, district_id = $3
			WHERE id = $1 RETURNING id`, id, datos.Nombre, datos.DistritoID).Scan(&updated)
		return updated, err
	})
}

func (r *Repository) List(ctx context.Context, filtros Filtros) ([]Institucion, int64, error) {
	rows, err := r.pool.Query(ctx, `
		WITH filtradas AS (
			SELECT id, nombre, codigo_css, activo, district_id, ubicacion
			FROM public.institutions
			WHERE (activo OR $1)
			  AND (strpos(lower(nombre), lower($2)) > 0
			    OR strpos(lower(COALESCE(codigo_css, '')), lower($2)) > 0)
		)
		SELECT `+institutionColumns+`, total.cantidad
		FROM (SELECT count(*) AS cantidad FROM filtradas) AS total
		LEFT JOIN (
			SELECT * FROM filtradas ORDER BY lower(nombre), id LIMIT $3 OFFSET $4
		) AS i ON true
		`+institutionJoin+`
		ORDER BY lower(i.nombre), i.id`,
		filtros.IncluirInactivas, filtros.Busqueda, filtros.Limite, filtros.Desplazamiento)
	if err != nil {
		return nil, 0, storageError(err)
	}
	defer rows.Close()

	lista := make([]Institucion, 0)
	var total int64
	for rows.Next() {
		item, err := scanInstitution(rows, &total)
		if err != nil {
			return nil, 0, err
		}
		if item.ID.Valid {
			lista = append(lista, item)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, 0, storageError(err)
	}
	return lista, total, nil
}
