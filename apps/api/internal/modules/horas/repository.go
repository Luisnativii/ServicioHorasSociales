package horas

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrClosed   = errors.New("proyecto o participación cerrados")
	ErrDate     = errors.New("fecha fuera del período permitido")
	ErrConflict = errors.New("solicitud usada con otros datos")
)

const Goal = 300

type Movement struct {
	ID          string    `json:"id"`
	ProjectID   string    `json:"proyecto_id"`
	Project     string    `json:"proyecto"`
	Hours       int32     `json:"horas"`
	Nature      string    `json:"naturaleza"`
	Date        string    `json:"fecha_horas"`
	Recorded    time.Time `json:"registrado_en"`
	Origin      string    `json:"origen"`
	Concept     *string   `json:"concepto"`
	PortalActor *string   `json:"administrador_id"`
	CoreActor   *string   `json:"usuario_core_id"`
}

type Progress struct {
	Internal        int64 `json:"internas"`
	External        int64 `json:"externas"`
	InternalMissing int64 `json:"internas_pendientes"`
	ExternalMissing int64 `json:"externas_pendientes"`
	Goal            int   `json:"meta_por_categoria"`
}

type Credit struct {
	Hours   int32
	Date    string
	Concept string
	Request pgtype.UUID
}

type Receipt struct {
	ID       string `json:"id"`
	Replayed bool   `json:"repetida"`
}

type Repository struct{ pool *pgxpool.Pool }

func New(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

func (r *Repository) List(ctx context.Context, student pgtype.UUID, limit int, offset int64) ([]Movement, Progress, int64, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, Progress{}, 0, err
	}
	defer tx.Rollback(ctx)
	var exists bool
	if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT FROM public.students WHERE id=$1)`, student).Scan(&exists); err != nil {
		return nil, Progress{}, 0, err
	}
	if !exists {
		return nil, Progress{}, 0, pgx.ErrNoRows
	}

	var progress Progress
	var count int64
	err = tx.QueryRow(ctx, `SELECT coalesce(sum(h.horas) FILTER (WHERE h.naturaleza='INTERNA'),0),
		coalesce(sum(h.horas) FILTER (WHERE h.naturaleza='EXTERNA'),0),count(*)
		FROM public.hour_movements h JOIN public.project_enrollments e ON e.id=h.enrollment_id
		WHERE e.student_id=$1`, student).Scan(&progress.Internal, &progress.External, &count)
	if err != nil {
		return nil, Progress{}, 0, err
	}
	progress.Goal = Goal
	progress.InternalMissing = max(Goal-progress.Internal, 0)
	progress.ExternalMissing = max(Goal-progress.External, 0)

	rows, err := tx.Query(ctx, `SELECT h.id::text,p.id::text,p.titulo,h.horas,h.naturaleza,h.fecha_horas::text,h.registrado_en,
		h.origen,coalesce(h.concepto,h.descripcion),a.administrador_id::text,h.registrado_por::text
		FROM public.hour_movements h JOIN public.project_enrollments e ON e.id=h.enrollment_id
		JOIN public.projects p ON p.id=e.project_id
		LEFT JOIN horas_sociales.autor_movimiento a ON a.movimiento_id=h.id
		WHERE e.student_id=$1 ORDER BY h.fecha_horas DESC,h.registrado_en DESC,h.id LIMIT $2 OFFSET $3`, student, limit, offset)
	if err != nil {
		return nil, Progress{}, 0, err
	}
	items := make([]Movement, 0)
	for rows.Next() {
		var item Movement
		err = rows.Scan(&item.ID, &item.ProjectID, &item.Project, &item.Hours, &item.Nature, &item.Date, &item.Recorded,
			&item.Origin, &item.Concept, &item.PortalActor, &item.CoreActor)
		if err != nil {
			rows.Close()
			return nil, Progress{}, 0, err
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, Progress{}, 0, err
	}
	return items, progress, count, tx.Commit(ctx)
}

func (r *Repository) Register(ctx context.Context, project, enrollment, actor pgtype.UUID, credit Credit) (Receipt, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Receipt{}, err
	}
	defer tx.Rollback(ctx)
	// Assignments and closures lock the same project row first.
	var state, nature string
	err = tx.QueryRow(ctx, `SELECT estado,naturaleza FROM public.projects WHERE id=$1 FOR UPDATE`, project).Scan(&state, &nature)
	if err != nil {
		return Receipt{}, err
	}
	receipt, err := previous(ctx, tx, project, enrollment, actor, credit)
	if !errors.Is(err, pgx.ErrNoRows) {
		return receipt, err
	}
	if state != "ACTIVO" {
		return Receipt{}, ErrClosed
	}

	var entryState string
	var validDate bool
	err = tx.QueryRow(ctx, `SELECT estado,$3::text::date BETWEEN
		(asignada_en AT TIME ZONE 'America/El_Salvador')::date
		AND (clock_timestamp() AT TIME ZONE 'America/El_Salvador')::date
		FROM public.project_enrollments WHERE id=$2 AND project_id=$1 FOR UPDATE`,
		project, enrollment, credit.Date).Scan(&entryState, &validDate)
	if err != nil {
		return Receipt{}, err
	}
	if entryState != "ASIGNADA" && entryState != "EN_CURSO" {
		return Receipt{}, ErrClosed
	}
	if !validDate {
		return Receipt{}, ErrDate
	}
	if _, err = tx.Exec(ctx, `SELECT set_config('app.actor',$1,true)`, "admin:"+actor.String()); err != nil {
		return Receipt{}, err
	}

	var id pgtype.UUID
	err = tx.QueryRow(ctx, `INSERT INTO public.hour_movements(enrollment_id,horas,naturaleza,fecha_horas,solicitud_id,origen,concepto)
		VALUES($1,$2,$3,$4::text::date,$5,'MANUAL',$6) ON CONFLICT ON CONSTRAINT uq_hour_movements_solicitud DO NOTHING
		RETURNING id`, enrollment, credit.Hours, nature, credit.Date, credit.Request, credit.Concept).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		// Another project may have claimed the key while this request was waiting.
		receipt, err = previous(ctx, tx, project, enrollment, actor, credit)
		if errors.Is(err, pgx.ErrNoRows) {
			return Receipt{}, ErrConflict
		}
		return receipt, err
	}
	if err != nil {
		return Receipt{}, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO horas_sociales.autor_movimiento(movimiento_id,administrador_id)
		VALUES($1,$2)`, id, actor); err != nil {
		return Receipt{}, err
	}
	return Receipt{ID: id.String()}, tx.Commit(ctx)
}

func previous(ctx context.Context, tx pgx.Tx, project, enrollment, actor pgtype.UUID, credit Credit) (Receipt, error) {
	var result Receipt
	var same bool
	err := tx.QueryRow(ctx, `SELECT h.id::text,coalesce(e.project_id=$2 AND h.enrollment_id=$3 AND a.administrador_id=$4
		AND h.horas=$5 AND h.fecha_horas=$6::text::date AND h.concepto=$7,false)
		FROM public.hour_movements h JOIN public.project_enrollments e ON e.id=h.enrollment_id
		LEFT JOIN horas_sociales.autor_movimiento a ON a.movimiento_id=h.id WHERE h.solicitud_id=$1`,
		credit.Request, project, enrollment, actor, credit.Hours, credit.Date, credit.Concept).Scan(&result.ID, &same)
	if err != nil {
		return Receipt{}, err
	}
	if !same {
		return Receipt{}, ErrConflict
	}
	result.Replayed = true
	return result, nil
}
