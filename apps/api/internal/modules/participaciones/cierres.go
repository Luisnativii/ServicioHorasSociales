package participaciones

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrClosedParticipation = errors.New("participación ya cerrada")
	ErrInvalidCloseDate    = errors.New("fecha anterior al inicio, asignación u horas acreditadas")
	ErrCannotFinish        = errors.New("el proyecto requiere al menos una participación y ninguna abierta")
)

type CloseInput struct {
	Action string
	Date   string
	Reason *string
}

type CloseResult struct {
	ProjectState       string `json:"estado_proyecto"`
	ParticipationState string `json:"estado_participacion,omitempty"`
}

func (r *Repository) Close(ctx context.Context, project, enrollment, actor pgtype.UUID, input CloseInput) (CloseResult, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return CloseResult{}, err
	}
	defer tx.Rollback(ctx)
	var state string
	var projectDateOK bool
	err = tx.QueryRow(ctx, `SELECT estado, fecha_inicio IS NULL OR fecha_inicio <= $2::text::date
		FROM public.projects WHERE id=$1 FOR UPDATE`, project, input.Date).Scan(&state, &projectDateOK)
	if err != nil {
		return CloseResult{}, err
	}
	if state != "ACTIVO" {
		return CloseResult{}, ErrProjectClosed
	}
	if !projectDateOK {
		return CloseResult{}, ErrInvalidCloseDate
	}

	var dateInvalid bool
	err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT FROM public.hour_movements h JOIN public.project_enrollments e
	  ON e.id=h.enrollment_id WHERE e.project_id=$1 AND ($2::uuid IS NULL OR e.id=$2)
	  AND h.fecha_horas > $3::text::date)`, project, enrollment, input.Date).Scan(&dateInvalid)
	if err != nil {
		return CloseResult{}, err
	}
	if dateInvalid {
		return CloseResult{}, ErrInvalidCloseDate
	}
	var instant time.Time
	err = tx.QueryRow(ctx, `SELECT clock_timestamp() WHERE $1::text::date <= (clock_timestamp() AT TIME ZONE 'America/El_Salvador')::date`,
		// A single timestamp is reused for all affected rows.
		input.Date).Scan(&instant)
	if errors.Is(err, pgx.ErrNoRows) {
		return CloseResult{}, ErrInvalidCloseDate
	}
	if err != nil {
		return CloseResult{}, err
	}
	if _, err = tx.Exec(ctx, `SELECT set_config('app.actor',$1,true)`, "admin:"+actor.String()); err != nil {
		return CloseResult{}, err
	}

	result := CloseResult{ProjectState: "ACTIVO"}
	if enrollment.Valid {
		var enrollmentState string
		var dateOK bool
		err = tx.QueryRow(ctx, `SELECT estado, (asignada_en AT TIME ZONE 'America/El_Salvador')::date <= $3::text::date
			FROM public.project_enrollments WHERE id=$1 AND project_id=$2 FOR UPDATE`, enrollment, project, input.Date).Scan(&enrollmentState, &dateOK)
		if err != nil {
			return CloseResult{}, err
		}
		if enrollmentState != "ASIGNADA" && enrollmentState != "EN_CURSO" {
			return CloseResult{}, ErrClosedParticipation
		}
		if !dateOK {
			return CloseResult{}, ErrInvalidCloseDate
		}
		result.ParticipationState = "FINALIZADA"
		if input.Action == "retirar" {
			result.ParticipationState = "RETIRADA"
		}
		if _, err = tx.Exec(ctx, `UPDATE public.project_enrollments SET estado=$2,cerrada_en=$3 WHERE id=$1`,
			enrollment, result.ParticipationState, instant); err != nil {
			return CloseResult{}, err
		}
		if err = recordClose(ctx, tx, project, enrollment, actor, input, instant, result.ParticipationState); err != nil {
			return CloseResult{}, err
		}
		if input.Action == "retirar" {
			var allWithdrawn bool
			err = tx.QueryRow(ctx, `SELECT NOT EXISTS(SELECT FROM public.project_enrollments WHERE project_id=$1 AND estado<>'RETIRADA')`, project).Scan(&allWithdrawn)
			if err != nil {
				return CloseResult{}, err
			}
			if allWithdrawn {
				// The last withdrawal also closes the project. Its date must cover all credited hours.
				err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT FROM public.hour_movements h JOIN public.project_enrollments e
					ON e.id=h.enrollment_id WHERE e.project_id=$1 AND h.fecha_horas > $2::text::date)`, project, input.Date).Scan(&dateInvalid)
				if err != nil {
					return CloseResult{}, err
				}
				if dateInvalid {
					return CloseResult{}, ErrInvalidCloseDate
				}
				result.ProjectState = "ABANDONADO"
			}
		}
	} else if input.Action == "cancelar" {
		err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT FROM public.project_enrollments WHERE project_id=$1 AND estado IN ('ASIGNADA','EN_CURSO')
			AND (asignada_en AT TIME ZONE 'America/El_Salvador')::date > $2::text::date)`, project, input.Date).Scan(&dateInvalid)
		if err != nil {
			return CloseResult{}, err
		}
		if dateInvalid {
			return CloseResult{}, ErrInvalidCloseDate
		}
		_, err = tx.Exec(ctx, `WITH affected AS (
			UPDATE public.project_enrollments SET estado='CANCELADA',cerrada_en=$2
			WHERE project_id=$1 AND estado IN ('ASIGNADA','EN_CURSO') RETURNING id
		) INSERT INTO horas_sociales.cierre_evento(proyecto_id,participacion_id,estado,fecha_cierre,cerrada_en,actor_id,motivo)
			SELECT $1,id,'CANCELADA',$3::text::date,$2,$4,$5 FROM affected`, project, instant, input.Date, actor, input.Reason)
		if err != nil {
			return CloseResult{}, err
		}
		result.ProjectState = "CANCELADO"
	} else {
		var eligible bool
		err = tx.QueryRow(ctx, `SELECT count(*)>0 AND count(*) FILTER(WHERE estado IN ('ASIGNADA','EN_CURSO'))=0
			FROM public.project_enrollments WHERE project_id=$1`, project).Scan(&eligible)
		if err != nil {
			return CloseResult{}, err
		}
		if !eligible {
			return CloseResult{}, ErrCannotFinish
		}
		result.ProjectState = "FINALIZADO"
	}
	if result.ProjectState != "ACTIVO" {
		if _, err = tx.Exec(ctx, `UPDATE public.projects SET estado=$2,fecha_cierre=$3::text::date WHERE id=$1`,
			project, result.ProjectState, input.Date); err != nil {
			return CloseResult{}, err
		}
		if err = recordClose(ctx, tx, project, pgtype.UUID{}, actor, input, instant, result.ProjectState); err != nil {
			return CloseResult{}, err
		}
	}
	return result, tx.Commit(ctx)
}

func recordClose(ctx context.Context, tx pgx.Tx, project, enrollment, actor pgtype.UUID, input CloseInput, instant time.Time, state string) error {
	_, err := tx.Exec(ctx, `INSERT INTO horas_sociales.cierre_evento(proyecto_id,participacion_id,estado,fecha_cierre,cerrada_en,actor_id,motivo)
		VALUES($1,$2,$3,$4::text::date,$5,$6,$7)`, project, enrollment, state, input.Date, instant, actor, input.Reason)
	return err
}
