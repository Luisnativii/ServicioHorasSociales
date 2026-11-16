package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// CoreCatalogProbe checks the core data with the same role used by the API.
type CoreCatalogProbe struct {
	Pool *pgxpool.Pool
}

func (p CoreCatalogProbe) Ping(ctx context.Context) error {
	var accessible bool
	// An empty catalog is valid. Referencing both tables and the FK columns
	// still verifies the schema and permissions before serving private routes.
	return p.Pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM public.careers AS c
			JOIN public.faculties AS f ON f.id = c.faculty_id
		)
	`).Scan(&accessible)
}
