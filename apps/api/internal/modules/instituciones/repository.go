package instituciones

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrInvalidID    = errors.New("identificador de institución inválido")
	ErrInvalidActor = errors.New("identificador de administrador inválido")
	ErrDuplicate    = errors.New("nombre de institución duplicado")
	ErrInvalidSede  = errors.New("distrito de sede inexistente")
	ErrNotFound     = errors.New("institución no encontrada")
	ErrStorage      = errors.New("no se pudo completar la operación de instituciones")
)

type Sede struct {
	DistritoID pgtype.UUID
	Nombre     string
}

type Contacto struct {
	Nombre, Correo, Telefono, Extension *string
}

type Institucion struct {
	Contacto          Contacto
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
	return r.write(ctx, actor, func(tx pgx.Tx) (pgtype.UUID, error) {
		_, err := tx.Exec(ctx, `UPDATE public.institutions SET activo = false WHERE id = $1 AND activo`, id)
		return id, err
	})
}

func (r *Repository) write(ctx context.Context, actor pgtype.UUID, mutate func(pgx.Tx) (pgtype.UUID, error)) (Institucion, error) {
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
	id, err := mutate(tx)
	if err != nil {
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

const institutionColumns = "i.id, i.nombre, i.codigo_css, i.activo, d.id, d.nombre, i.ubicacion, i.contacto_nombre, i.contacto_email, i.telefono, i.contacto_extension"
const institutionJoin = "LEFT JOIN public.geo_districts AS d ON d.id = i.district_id"

func findByID(ctx context.Context, db rowQuerier, id pgtype.UUID) (Institucion, error) {
	return scanInstitution(db.QueryRow(ctx, `
		SELECT `+institutionColumns+`
		FROM public.institutions AS i
		`+institutionJoin+`
		WHERE i.id = $1`, id))
}

type scanner interface{ Scan(...any) error }

func scanInstitution(row scanner, extra ...any) (Institucion, error) {
	var result Institucion
	var nombre pgtype.Text
	var activo pgtype.Bool
	var distritoID pgtype.UUID
	var distritoNombre pgtype.Text
	err := row.Scan(append([]any{
		&result.ID, &nombre, &result.CodigoCSS, &activo, &distritoID, &distritoNombre, &result.UbicacionHeredada,
		&result.Contacto.Nombre, &result.Contacto.Correo, &result.Contacto.Telefono, &result.Contacto.Extension,
	}, extra...)...)
	if err != nil {
		return Institucion{}, storageError(err)
	}
	result.Nombre, result.Activo = nombre.String, activo.Bool
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
	}
	var db *pgconn.PgError
	if errors.As(err, &db) {
		switch {
		case db.Code == "23505" && db.ConstraintName == "uq_institutions_nombre":
			return ErrDuplicate
		case db.Code == "23503" && db.ConstraintName == "fk_institutions_district":
			return ErrInvalidSede
		}
	}
	return ErrStorage
}
