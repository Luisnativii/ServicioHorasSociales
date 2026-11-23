package participaciones

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrProjectClosed   = errors.New("proyecto no activo")
	ErrStudentInactive = errors.New("estudiante no activo")
	ErrDuplicate       = errors.New("participación existente")
	ErrCapacity        = errors.New("requiere confirmar exceso de cupo")
)

type Item struct {
	ID        string     `json:"id"`
	ProjectID string     `json:"proyecto_id"`
	StudentID string     `json:"estudiante_id"`
	Project   string     `json:"proyecto"`
	Student   string     `json:"estudiante"`
	Carnet    string     `json:"carnet"`
	Estado    string     `json:"estado"`
	Asignada  time.Time  `json:"asignada_en"`
	Cerrada   *time.Time `json:"cerrada_en"`
}

type Repository struct{ pool *pgxpool.Pool }

func New(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

const participants = `
FROM public.project_enrollments e
JOIN public.projects p ON p.id=e.project_id
JOIN public.students s ON s.id=e.student_id
WHERE ($1::uuid IS NULL OR p.id=$1) AND ($2::uuid IS NULL OR s.id=$2)`

func (r *Repository) List(ctx context.Context, project, student pgtype.UUID, limit int, offset int64) ([]Item, int64, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, 0, err
	}
	defer tx.Rollback(ctx)
	var exists bool
	err = tx.QueryRow(ctx, `SELECT CASE WHEN $1::uuid IS NOT NULL
		THEN EXISTS(SELECT FROM public.projects WHERE id=$1)
		ELSE EXISTS(SELECT FROM public.students WHERE id=$2) END`, project, student).Scan(&exists)
	if err != nil {
		return nil, 0, err
	}
	if !exists {
		return nil, 0, pgx.ErrNoRows
	}
	rows, err := tx.Query(ctx, `SELECT e.id::text,p.id::text,s.id::text,p.titulo,s.nombre_completo,s.carnet,
		e.estado,e.asignada_en,e.cerrada_en `+participants+
		` ORDER BY e.asignada_en DESC,e.id LIMIT $3 OFFSET $4`, project, student, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	items := make([]Item, 0)
	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.ID, &item.ProjectID, &item.StudentID, &item.Project, &item.Student, &item.Carnet,
			&item.Estado, &item.Asignada, &item.Cerrada); err != nil {
			rows.Close()
			return nil, 0, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	var total int64
	if err := tx.QueryRow(ctx, `SELECT count(*) `+participants, project, student).Scan(&total); err != nil {
		return nil, 0, err
	}
	return items, total, tx.Commit(ctx)
}

func (r *Repository) Assign(ctx context.Context, project, student, actor pgtype.UUID, confirmed bool) (string, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)
	var state string
	var capacity *int32
	err = tx.QueryRow(ctx, `SELECT estado,cupos FROM public.projects WHERE id=$1 FOR UPDATE`, project).Scan(&state, &capacity)
	if err != nil {
		return "", err
	}
	if state != "ACTIVO" {
		return "", ErrProjectClosed
	}
	var active bool
	err = tx.QueryRow(ctx, `SELECT activo FROM public.students WHERE id=$1 FOR SHARE`, student).Scan(&active)
	if err != nil {
		return "", err
	}
	if !active {
		return "", ErrStudentInactive
	}
	var duplicate bool
	err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT FROM public.project_enrollments WHERE project_id=$1 AND student_id=$2)`, project, student).Scan(&duplicate)
	if err != nil {
		return "", err
	}
	if duplicate {
		return "", ErrDuplicate
	}
	if capacity != nil && !confirmed {
		// A separate statement sees commits made by previous holders of the project lock.
		var occupied int64
		err = tx.QueryRow(ctx, `SELECT count(*) FROM public.project_enrollments WHERE project_id=$1
			AND estado IN ('ASIGNADA','EN_CURSO')`, project).Scan(&occupied)
		if err != nil {
			return "", err
		}
		if occupied >= int64(*capacity) {
			return "", ErrCapacity
		}
	}
	if _, err = tx.Exec(ctx, `SELECT set_config('app.actor',$1,true)`, "admin:"+actor.String()); err != nil {
		return "", err
	}
	var id string
	err = tx.QueryRow(ctx, `INSERT INTO public.project_enrollments(project_id,student_id)
		VALUES($1,$2) ON CONFLICT ON CONSTRAINT uq_project_enrollments_student_project DO NOTHING
		RETURNING id::text`, project, student).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrDuplicate
	}
	if err != nil {
		return "", err
	}
	return id, tx.Commit(ctx)
}
