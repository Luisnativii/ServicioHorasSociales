package catalogos

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LugarTerritorial struct {
	ID      string  `json:"id"`
	Codigo  string  `json:"codigo"`
	Nombre  string  `json:"nombre"`
	PadreID *string `json:"padre_id,omitempty"`
}

type TerritorioRepository struct {
	pool *pgxpool.Pool
}

func NewTerritorioRepository(pool *pgxpool.Pool) *TerritorioRepository {
	return &TerritorioRepository{pool: pool}
}

func (r *TerritorioRepository) ListDepartamentos(ctx context.Context, limite int, offset int64, _ pgtype.UUID) ([]LugarTerritorial, int64, error) {
	return r.listar(ctx, `SELECT id, codigo, nombre, NULL::uuid AS padre_id FROM public.geo_departments`, limite, offset)
}

func (r *TerritorioRepository) ListMunicipios(ctx context.Context, limite int, offset int64, departamento pgtype.UUID) ([]LugarTerritorial, int64, error) {
	return r.listar(ctx, `SELECT id, codigo, nombre, department_id AS padre_id
		FROM public.geo_municipalities WHERE department_id = $3`, limite, offset, departamento)
}

func (r *TerritorioRepository) ListDistritos(ctx context.Context, limite int, offset int64, municipio pgtype.UUID) ([]LugarTerritorial, int64, error) {
	return r.listar(ctx, `SELECT id, codigo, nombre, municipality_id AS padre_id
		FROM public.geo_districts WHERE municipality_id = $3`, limite, offset, municipio)
}

func (r *TerritorioRepository) listar(ctx context.Context, seleccion string, limite int, offset int64, padre ...pgtype.UUID) ([]LugarTerritorial, int64, error) {
	args := []any{limite, offset}
	for _, id := range padre {
		args = append(args, id)
	}
	// La selección es SQL fijo del repositorio. El total y la página usan la misma instantánea.
	rows, err := r.pool.Query(ctx, `WITH catalogo AS (`+seleccion+`)
		SELECT pagina.id::text, pagina.codigo, pagina.nombre, pagina.padre_id::text, total.cantidad
		FROM (SELECT count(*) AS cantidad FROM catalogo) AS total
		LEFT JOIN (
			SELECT id, codigo, nombre, padre_id FROM catalogo ORDER BY codigo, id LIMIT $1 OFFSET $2
		) AS pagina ON true
		ORDER BY pagina.codigo, pagina.id`, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	lugares := make([]LugarTerritorial, 0)
	var total int64
	for rows.Next() {
		var id, codigo, nombre, padreID pgtype.Text
		if err := rows.Scan(&id, &codigo, &nombre, &padreID, &total); err != nil {
			return nil, 0, err
		}
		if id.Valid {
			lugar := LugarTerritorial{ID: id.String, Codigo: codigo.String, Nombre: nombre.String}
			if padreID.Valid {
				lugar.PadreID = &padreID.String
			}
			lugares = append(lugares, lugar)
		}
	}
	return lugares, total, rows.Err()
}
