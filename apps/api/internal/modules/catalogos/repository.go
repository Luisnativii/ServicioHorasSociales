// Package catalogos owns persistence for independent project types and areas.
package catalogos

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	catalogdb "horas-sociales/api/internal/modules/catalogos/db"
)

var (
	ErrInvalidName = errors.New("nombre de catálogo inválido")
	ErrInvalidID   = errors.New("identificador de catálogo inválido")
	ErrDuplicate   = errors.New("el nombre ya existe en el catálogo")
	ErrNotFound    = errors.New("opción de catálogo no encontrada")
	ErrStorage     = errors.New("no se pudo completar la operación de catálogo")
)

type TipoProyecto struct {
	ID     pgtype.UUID
	Nombre string
	Activo bool
}

type Area struct {
	ID     pgtype.UUID
	Nombre string
	Activo bool
}

type Repository struct{ queries *catalogdb.Queries }

func NewRepository(db catalogdb.DBTX) *Repository { return &Repository{catalogdb.New(db)} }

// VisibleName preserves the entered capitalization while removing exterior whitespace.
func VisibleName(input string) (string, error) {
	name := strings.TrimSpace(input)
	if name == "" {
		return "", ErrInvalidName
	}
	return name, nil
}

func (r *Repository) CreateTipo(ctx context.Context, nombre string) (TipoProyecto, error) {
	visible, err := VisibleName(nombre)
	if err != nil {
		return TipoProyecto{}, err
	}
	row, err := r.queries.CreateTipoProyecto(ctx, visible)
	return fromTipoRow(row), storageError(err)
}

func (r *Repository) FindTipoByID(ctx context.Context, id pgtype.UUID) (TipoProyecto, error) {
	if !validID(id) {
		return TipoProyecto{}, ErrInvalidID
	}
	row, err := r.queries.GetTipoProyectoPorID(ctx, id)
	return fromTipoRow(row), storageError(err)
}

func (r *Repository) ListTipos(ctx context.Context, includeInactive bool) ([]TipoProyecto, error) {
	rows, err := r.queries.ListTiposProyecto(ctx, includeInactive)
	if err != nil {
		return nil, storageError(err)
	}
	result := make([]TipoProyecto, 0, len(rows))
	for _, row := range rows {
		result = append(result, fromTipoRow(row))
	}
	return result, nil
}

func (r *Repository) RenameTipo(ctx context.Context, id pgtype.UUID, nombre string) (TipoProyecto, error) {
	if !validID(id) {
		return TipoProyecto{}, ErrInvalidID
	}
	visible, err := VisibleName(nombre)
	if err != nil {
		return TipoProyecto{}, err
	}
	row, err := r.queries.UpdateTipoProyectoNombre(ctx, catalogdb.UpdateTipoProyectoNombreParams{ID: id, Nombre: visible})
	return fromTipoRow(row), storageError(err)
}

func (r *Repository) SetTipoActive(ctx context.Context, id pgtype.UUID, active bool) (TipoProyecto, error) {
	if !validID(id) {
		return TipoProyecto{}, ErrInvalidID
	}
	row, err := r.queries.SetTipoProyectoActivo(ctx, catalogdb.SetTipoProyectoActivoParams{ID: id, Activo: active})
	return fromTipoRow(row), storageError(err)
}

func (r *Repository) CreateArea(ctx context.Context, nombre string) (Area, error) {
	visible, err := VisibleName(nombre)
	if err != nil {
		return Area{}, err
	}
	row, err := r.queries.CreateArea(ctx, visible)
	return fromAreaRow(row), storageError(err)
}

func (r *Repository) FindAreaByID(ctx context.Context, id pgtype.UUID) (Area, error) {
	if !validID(id) {
		return Area{}, ErrInvalidID
	}
	row, err := r.queries.GetAreaPorID(ctx, id)
	return fromAreaRow(row), storageError(err)
}

func (r *Repository) ListAreas(ctx context.Context, includeInactive bool) ([]Area, error) {
	rows, err := r.queries.ListAreas(ctx, includeInactive)
	if err != nil {
		return nil, storageError(err)
	}
	result := make([]Area, 0, len(rows))
	for _, row := range rows {
		result = append(result, fromAreaRow(row))
	}
	return result, nil
}

func (r *Repository) RenameArea(ctx context.Context, id pgtype.UUID, nombre string) (Area, error) {
	if !validID(id) {
		return Area{}, ErrInvalidID
	}
	visible, err := VisibleName(nombre)
	if err != nil {
		return Area{}, err
	}
	row, err := r.queries.UpdateAreaNombre(ctx, catalogdb.UpdateAreaNombreParams{ID: id, Nombre: visible})
	return fromAreaRow(row), storageError(err)
}

func (r *Repository) SetAreaActive(ctx context.Context, id pgtype.UUID, active bool) (Area, error) {
	if !validID(id) {
		return Area{}, ErrInvalidID
	}
	row, err := r.queries.SetAreaActivo(ctx, catalogdb.SetAreaActivoParams{ID: id, Activo: active})
	return fromAreaRow(row), storageError(err)
}

func validID(id pgtype.UUID) bool {
	return id.Valid && id.Bytes != [16]byte{}
}

func fromTipoRow(row catalogdb.HorasSocialesTipoProyecto) TipoProyecto {
	return TipoProyecto{ID: row.ID, Nombre: row.Nombre, Activo: row.Activo}
}

func fromAreaRow(row catalogdb.HorasSocialesArea) Area {
	return Area{ID: row.ID, Nombre: row.Nombre, Activo: row.Activo}
}

func storageError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	var postgres *pgconn.PgError
	if errors.As(err, &postgres) && postgres.Code == "23505" {
		switch postgres.ConstraintName {
		case "tipo_proyecto_nombre_canonico_unique", "area_nombre_canonico_unique":
			return ErrDuplicate
		}
	}
	return ErrStorage
}
