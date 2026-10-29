// Package admin owns persistence for the approved single ADMIN role.
package admin

import (
	"context"
	"errors"
	"net/mail"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	admindb "horas-sociales/api/internal/modules/admin/db"
)

var (
	ErrInvalidEmail = errors.New("correo inválido")
	ErrInvalidHash  = errors.New("se requiere un hash Argon2id")
	ErrDuplicate    = errors.New("el correo ya existe")
	ErrNotFound     = errors.New("administrador no encontrado")
	ErrPagination   = errors.New("paginación inválida")
	ErrStorage      = errors.New("no se pudo completar la operación de identidad")
)

type Administrator struct {
	ID             pgtype.UUID
	Correo         string
	ContrasenaHash string `json:"-"`
	Activo         bool
	CreadoEn       time.Time
	ActualizadoEn  time.Time
}

type Repository struct{ queries *admindb.Queries }

func NewRepository(db admindb.DBTX) *Repository { return &Repository{admindb.New(db)} }

func CanonicalEmail(input string) (string, error) {
	canonical := strings.ToLower(strings.TrimSpace(input))
	address, err := mail.ParseAddress(canonical)
	if err != nil || address.Address != canonical || address.Name != "" {
		return "", ErrInvalidEmail
	}
	return canonical, nil
}

func (r *Repository) Create(ctx context.Context, correo, hash string) (Administrator, error) {
	canonical, err := CanonicalEmail(correo)
	if err != nil {
		return Administrator{}, err
	}
	if !strings.HasPrefix(hash, "$argon2id$") {
		return Administrator{}, ErrInvalidHash
	}
	row, err := r.queries.CreateAdministrador(ctx, admindb.CreateAdministradorParams{Correo: canonical, ContrasenaHash: hash})
	return fromRow(row), storageError(err)
}

func (r *Repository) FindByEmail(ctx context.Context, correo string) (Administrator, error) {
	canonical, err := CanonicalEmail(correo)
	if err != nil {
		return Administrator{}, err
	}
	row, err := r.queries.GetAdministradorPorCorreo(ctx, canonical)
	return fromRow(row), storageError(err)
}

func (r *Repository) FindByID(ctx context.Context, id pgtype.UUID) (Administrator, error) {
	row, err := r.queries.GetAdministradorPorID(ctx, id)
	return fromRow(row), storageError(err)
}

func (r *Repository) SetActive(ctx context.Context, id pgtype.UUID, active bool) (Administrator, error) {
	row, err := r.queries.SetAdministradorActivo(ctx, admindb.SetAdministradorActivoParams{ID: id, Activo: active})
	return fromRow(row), storageError(err)
}

func (r *Repository) List(ctx context.Context, limit int32, offset int64) ([]Administrator, error) {
	if limit < 1 || limit > 100 || offset < 0 {
		return nil, ErrPagination
	}
	rows, err := r.queries.ListAdministradores(ctx, admindb.ListAdministradoresParams{Limite: limit, Desplazamiento: offset})
	if err != nil {
		return nil, storageError(err)
	}
	result := make([]Administrator, 0, len(rows))
	for _, row := range rows {
		result = append(result, Administrator{ID: row.ID, Correo: row.Correo, Activo: row.Activo, CreadoEn: row.CreadoEn.Time, ActualizadoEn: row.ActualizadoEn.Time})
	}
	return result, nil
}

func (r *Repository) Count(ctx context.Context) (int32, error) {
	total, err := r.queries.CountAdministradores(ctx)
	return total, storageError(err)
}

func fromRow(row admindb.HorasSocialesAdministrador) Administrator {
	return Administrator{row.ID, row.Correo, row.ContrasenaHash, row.Activo, row.CreadoEn.Time, row.ActualizadoEn.Time}
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
	if errors.As(err, &postgres) && postgres.Code == "23505" && postgres.ConstraintName == "administrador_correo_unique" {
		return ErrDuplicate
	}
	return ErrStorage
}
