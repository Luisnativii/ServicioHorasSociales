package catalogos

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"

	"horas-sociales/api/internal/modules/catalogos/academiadb"
)

var ErrInvalidPage = errors.New("paginación de catálogo inválida")

type Facultad struct {
	ID     pgtype.UUID
	Nombre string
	Activo bool
}

type Carrera struct {
	ID             pgtype.UUID
	FacultyID      pgtype.UUID
	Nombre         string
	Activo         bool
	FacultadActiva bool
}

type FiltrosAcademicos struct {
	Busqueda         string
	IncluirInactivas bool
	Limite           int32
	Desplazamiento   int64
}

type AcademiaRepository struct{ queries *academiadb.Queries }

func NewAcademiaRepository(db academiadb.DBTX) *AcademiaRepository {
	return &AcademiaRepository{queries: academiadb.New(db)}
}

func (r *AcademiaRepository) FindFacultad(ctx context.Context, id pgtype.UUID) (Facultad, error) {
	if !validID(id) {
		return Facultad{}, ErrInvalidID
	}
	row, err := r.queries.FindFacultad(ctx, id)
	return Facultad(row), storageError(err)
}

func (r *AcademiaRepository) FindCarrera(ctx context.Context, id pgtype.UUID) (Carrera, error) {
	if !validID(id) {
		return Carrera{}, ErrInvalidID
	}
	row, err := r.queries.FindCarrera(ctx, id)
	return Carrera(row), storageError(err)
}

func (r *AcademiaRepository) ListFacultades(ctx context.Context, f FiltrosAcademicos) ([]Facultad, int64, error) {
	if !f.validPage() {
		return nil, 0, ErrInvalidPage
	}
	search := strings.TrimSpace(f.Busqueda)
	rows, err := r.queries.ListFacultades(ctx, academiadb.ListFacultadesParams{
		IncluirInactivas: f.IncluirInactivas, Busqueda: search,
		Limite: f.Limite, Desplazamiento: f.Desplazamiento,
	})
	if err != nil {
		return nil, 0, storageError(err)
	}
	total, err := r.queries.CountFacultades(ctx, academiadb.CountFacultadesParams{
		IncluirInactivas: f.IncluirInactivas, Busqueda: search,
	})
	if err != nil {
		return nil, 0, storageError(err)
	}
	result := make([]Facultad, 0, len(rows))
	for _, row := range rows {
		result = append(result, Facultad(row))
	}
	return result, total, nil
}

func (r *AcademiaRepository) ListCarreras(ctx context.Context, f FiltrosAcademicos, facultadID pgtype.UUID) ([]Carrera, int64, error) {
	if !f.validPage() {
		return nil, 0, ErrInvalidPage
	}
	if facultadID.Valid && !validID(facultadID) {
		return nil, 0, ErrInvalidID
	}
	search := strings.TrimSpace(f.Busqueda)
	rows, err := r.queries.ListCarreras(ctx, academiadb.ListCarrerasParams{
		IncluirInactivas: f.IncluirInactivas, FacultadID: facultadID, Busqueda: search,
		Limite: f.Limite, Desplazamiento: f.Desplazamiento,
	})
	if err != nil {
		return nil, 0, storageError(err)
	}
	total, err := r.queries.CountCarreras(ctx, academiadb.CountCarrerasParams{
		IncluirInactivas: f.IncluirInactivas, FacultadID: facultadID, Busqueda: search,
	})
	if err != nil {
		return nil, 0, storageError(err)
	}
	result := make([]Carrera, 0, len(rows))
	for _, row := range rows {
		result = append(result, Carrera(row))
	}
	return result, total, nil
}

func (f FiltrosAcademicos) validPage() bool {
	return f.Limite >= 1 && f.Limite <= 100 && f.Desplazamiento >= 0
}
