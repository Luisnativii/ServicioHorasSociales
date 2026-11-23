package estudiantes

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
	ErrCareer  = errors.New("carrera no disponible")
	ErrReason  = errors.New("falta motivo del cambio de carrera")
	ErrHistory = errors.New("historial académico no disponible")
)

type Reference struct {
	ID     string `json:"id"`
	Nombre string `json:"nombre"`
}

type Student struct {
	ID             string    `json:"id"`
	Carnet         string    `json:"carnet"`
	NombreCompleto string    `json:"nombre_completo"`
	Nombres        *string   `json:"nombres"`
	Apellidos      *string   `json:"apellidos"`
	Correo         string    `json:"correo_institucional"`
	Carrera        Reference `json:"carrera"`
	Facultad       Reference `json:"facultad"`
	Activo         bool      `json:"activo"`
}

type Period struct {
	Carrera  Reference  `json:"carrera"`
	Facultad Reference  `json:"facultad"`
	Desde    *time.Time `json:"desde"`
	Hasta    *time.Time `json:"hasta"`
	Motivo   *string    `json:"motivo"`
}

type Detail struct {
	Student
	Vigencias           []Period `json:"vigencias"`
	HistorialDisponible bool     `json:"historial_disponible"`
}

type Input struct {
	Carnet, Correo, Motivo string
	Nombres, Apellidos     *string
	Carrera                pgtype.UUID
}

type Repository struct{ pool *pgxpool.Pool }

func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

const columns = `s.id::text, s.carnet, s.nombre_completo, s.nombres, s.apellidos,
	s.correo_institucional::text, c.id::text, c.nombre, f.id::text, f.nombre, s.activo`
const joins = `JOIN public.careers c ON c.id = s.career_id JOIN public.faculties f ON f.id = c.faculty_id`

type scanner interface{ Scan(...any) error }

func scanStudent(row scanner) (Student, error) {
	var s Student
	err := row.Scan(&s.ID, &s.Carnet, &s.NombreCompleto, &s.Nombres, &s.Apellidos,
		&s.Correo, &s.Carrera.ID, &s.Carrera.Nombre, &s.Facultad.ID, &s.Facultad.Nombre, &s.Activo)
	return s, err
}

func (r *Repository) List(ctx context.Context, search string, faculty, career pgtype.UUID, limit int, offset int64) ([]Student, int64, error) {
	const filter = `WHERE ($1 = '' OR strpos(lower(s.nombre_completo), lower($1)) > 0 OR strpos(s.carnet, $1) > 0)
		AND ($2::uuid IS NULL OR f.id = $2) AND ($3::uuid IS NULL OR c.id = $3)`
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, 0, err
	}
	defer tx.Rollback(ctx)
	var total int64
	err = tx.QueryRow(ctx, `SELECT count(*) FROM public.students s `+joins+" "+filter, search, faculty, career).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	rows, err := tx.Query(ctx, `SELECT `+columns+` FROM public.students s `+joins+" "+filter+
		` ORDER BY lower(s.nombre_completo), s.id LIMIT $4 OFFSET $5`, search, faculty, career, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	list := make([]Student, 0)
	for rows.Next() {
		s, err := scanStudent(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, s)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return list, total, tx.Commit(ctx)
}

func (r *Repository) Get(ctx context.Context, id pgtype.UUID) (Detail, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return Detail{}, err
	}
	defer tx.Rollback(ctx)
	s, err := scanStudent(tx.QueryRow(ctx, `SELECT `+columns+` FROM public.students s `+joins+` WHERE s.id = $1`, id))
	if err != nil {
		return Detail{}, err
	}
	detail := Detail{Student: s, Vigencias: make([]Period, 0)}
	detail.HistorialDisponible, err = historyAvailable(ctx, tx)
	if err != nil {
		return Detail{}, err
	}
	if detail.HistorialDisponible {
		rows, err := tx.Query(ctx, `SELECT c.id::text, c.nombre, f.id::text, f.nombre, h.valid_from, h.valid_until, h.reason
			FROM public.student_career_periods h JOIN public.careers c ON c.id = h.career_id
			JOIN public.faculties f ON f.id = h.faculty_id WHERE h.student_id = $1
			ORDER BY h.valid_from DESC NULLS LAST`, id)
		if err != nil {
			return Detail{}, err
		}
		defer rows.Close()
		for rows.Next() {
			var p Period
			if err := rows.Scan(&p.Carrera.ID, &p.Carrera.Nombre, &p.Facultad.ID, &p.Facultad.Nombre, &p.Desde, &p.Hasta, &p.Motivo); err != nil {
				return Detail{}, err
			}
			detail.Vigencias = append(detail.Vigencias, p)
		}
		if err := rows.Err(); err != nil {
			return Detail{}, err
		}
	}
	return detail, tx.Commit(ctx)
}

func historyAvailable(ctx context.Context, tx pgx.Tx) (bool, error) {
	var available bool
	err := tx.QueryRow(ctx, `SELECT EXISTS (
		SELECT 1 FROM pg_catalog.pg_trigger WHERE tgrelid = 'public.students'::regclass
		  AND tgfoid = to_regprocedure('public.fn_student_career_periods()')
		  AND tgname = 'trg_student_career_periods' AND tgenabled IN ('O', 'A')
	) AND coalesce(has_table_privilege(to_regclass('public.student_career_periods'), 'SELECT'), false)`).Scan(&available)
	return available, err
}

func (r *Repository) Save(ctx context.Context, id, actor pgtype.UUID, input Input) (Student, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Student{}, err
	}
	defer tx.Rollback(ctx)
	var create = !id.Valid
	var current pgtype.UUID
	if !create {
		if err := tx.QueryRow(ctx, `SELECT career_id FROM public.students WHERE id = $1 FOR UPDATE`, id).Scan(&current); err != nil {
			return Student{}, err
		}
	}
	careerChanged := create || current != input.Carrera
	if careerChanged {
		available, err := historyAvailable(ctx, tx)
		if err != nil {
			return Student{}, err
		}
		if !available {
			return Student{}, ErrHistory
		}
		if !create && input.Motivo == "" {
			return Student{}, ErrReason
		}
		var valid bool
		err = tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM public.careers c
			JOIN public.faculties f ON f.id = c.faculty_id WHERE c.id = $1 AND c.activo AND f.activo)`, input.Carrera).Scan(&valid)
		if err != nil {
			return Student{}, err
		}
		if !valid {
			return Student{}, ErrCareer
		}
	}
	if _, err := tx.Exec(ctx, `SELECT set_config('app.actor', $1, true), set_config('app.career_change_reason', $2, true)`,
		"admin:"+actor.String(), input.Motivo); err != nil {
		return Student{}, err
	}
	if create {
		err = tx.QueryRow(ctx, `INSERT INTO public.students (carnet, nombres, apellidos, correo_institucional, career_id)
			VALUES ($1, $2, $3, $4, $5) RETURNING id`, input.Carnet, input.Nombres, input.Apellidos, input.Correo, input.Carrera).Scan(&id)
	} else {
		_, err = tx.Exec(ctx, `UPDATE public.students SET carnet = $2, nombres = coalesce($3, nombres), apellidos = coalesce($4, apellidos),
			correo_institucional = $5, career_id = $6 WHERE id = $1`,
			id, input.Carnet, input.Nombres, input.Apellidos, input.Correo, input.Carrera)
	}
	if err != nil {
		return Student{}, err
	}
	item, err := scanStudent(tx.QueryRow(ctx, `SELECT `+columns+` FROM public.students s `+joins+` WHERE s.id = $1`, id))
	if err != nil {
		return Student{}, err
	}
	return item, tx.Commit(ctx)
}

func DuplicateField(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		switch pgErr.ConstraintName {
		case "uq_students_carnet":
			return "carnet"
		case "uq_students_correo_institucional":
			return "correo_institucional"
		}
	}
	return ""
}
