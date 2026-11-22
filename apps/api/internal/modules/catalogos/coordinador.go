package catalogos

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/jackc/pgx/v5/pgtype"

	"horas-sociales/api/internal/modules/catalogos/academiadb"
)

type CoordinadorRepository struct{ db academiadb.DBTX }

func NewCoordinadorRepository(db academiadb.DBTX) *CoordinadorRepository {
	return &CoordinadorRepository{db: db}
}

func (r *CoordinadorRepository) Find(ctx context.Context, id pgtype.UUID) (pgtype.Text, error) {
	if !validID(id) {
		return pgtype.Text{}, ErrInvalidID
	}
	var nombre pgtype.Text
	err := r.db.QueryRow(ctx, `
		SELECT coordinador.nombre
		FROM public.careers AS carrera
		LEFT JOIN horas_sociales.coordinador_carrera AS coordinador ON coordinador.carrera_id = carrera.id
		WHERE carrera.id = $1`, id).Scan(&nombre)
	return nombre, storageError(err)
}

func (r *CoordinadorRepository) Set(ctx context.Context, id pgtype.UUID, nombre pgtype.Text) (pgtype.Text, error) {
	if !validID(id) {
		return pgtype.Text{}, ErrInvalidID
	}
	if nombre.Valid {
		nombre.String = strings.TrimSpace(nombre.String)
		if nombre.String == "" || !utf8.ValidString(nombre.String) ||
			strings.ContainsRune(nombre.String, 0) || utf8.RuneCountInString(nombre.String) > 200 {
			return pgtype.Text{}, ErrInvalidName
		}
	}
	// Comprobar el UUID y escribir el dato del portal en una única sentencia.
	var guardado pgtype.Text
	err := r.db.QueryRow(ctx, `
		INSERT INTO horas_sociales.coordinador_carrera (carrera_id, nombre)
		SELECT id, $2 FROM public.careers WHERE id = $1
		ON CONFLICT (carrera_id) DO UPDATE SET nombre = EXCLUDED.nombre
		RETURNING nombre`, id, nombre).Scan(&guardado)
	return guardado, storageError(err)
}
