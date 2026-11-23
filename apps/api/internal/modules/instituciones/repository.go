package instituciones

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrInvalidID    = errors.New("identificador de institución inválido")
	ErrInvalidActor = errors.New("identificador de administrador inválido")
	ErrNotFound     = errors.New("institución no encontrada")
	ErrStorage      = errors.New("no se pudo completar la operación de instituciones")
)

type Sede struct {
	DistritoID pgtype.UUID
	Nombre     string
}

type Institucion struct {
	ID                pgtype.UUID
	Nombre            string
	CodigoCSS         *string
	Activo            bool
	Sede              *Sede
	UbicacionHeredada *string
}

type Repository struct{ pool *pgxpool.Pool }

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) FindByID(ctx context.Context, id pgtype.UUID) (Institucion, error) {
	if !validID(id) {
		return Institucion{}, ErrInvalidID
	}
	return findByID(ctx, r.pool, id)
}

// Deactivate preserves the institution and its project references. actor is
// the authenticated portal administrator, recorded using the core convention.
func (r *Repository) Deactivate(ctx context.Context, id, actor pgtype.UUID) (Institucion, error) {
	if !validID(id) {
		return Institucion{}, ErrInvalidID
	}
	if !validID(actor) {
		return Institucion{}, ErrInvalidActor
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Institucion{}, storageError(err)
	}
	defer func() {
		// Rollback still has to release the connection when the request expired.
		cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 3*time.Second)
		defer cancel()
		_ = tx.Rollback(cleanup)
	}()
	if _, err := tx.Exec(ctx, `
		SELECT pg_catalog.set_config('search_path', 'pg_catalog, public', true),
		       pg_catalog.set_config('app.actor', $1, true)`, "admin:"+actor.String()); err != nil {
		return Institucion{}, storageError(err)
	}
	if _, err := tx.Exec(ctx, `
		UPDATE public.institutions SET activo = false
		WHERE id = $1 AND activo`, id); err != nil {
		return Institucion{}, storageError(err)
	}
	result, err := findByID(ctx, tx, id)
	if err != nil {
		return Institucion{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Institucion{}, storageError(err)
	}
	return result, nil
}

type rowQuerier interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

func findByID(ctx context.Context, db rowQuerier, id pgtype.UUID) (Institucion, error) {
	var result Institucion
	var distritoID pgtype.UUID
	var distritoNombre pgtype.Text
	err := db.QueryRow(ctx, `
		SELECT i.id, i.nombre, i.codigo_css, i.activo, d.id, d.nombre, i.ubicacion
		FROM public.institutions AS i
		LEFT JOIN public.geo_districts AS d ON d.id = i.district_id
		WHERE i.id = $1`, id).Scan(
		&result.ID, &result.Nombre, &result.CodigoCSS, &result.Activo,
		&distritoID, &distritoNombre, &result.UbicacionHeredada,
	)
	if err != nil {
		return Institucion{}, storageError(err)
	}
	if distritoID.Valid {
		result.Sede = &Sede{DistritoID: distritoID, Nombre: distritoNombre.String}
	}
	return result, nil
}

func validID(id pgtype.UUID) bool {
	return id.Valid && id.Bytes != [16]byte{}
}

func storageError(err error) error {
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return ErrNotFound
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return err
	default:
		return ErrStorage
	}
}
