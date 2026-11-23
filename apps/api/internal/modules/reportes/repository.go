package reportes

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const ExportLimit = 5000

type Filter struct {
	From, To, State, Department, Municipality, District, Career, Type string
	Areas                                                             []string
}

type Summary struct {
	Students       int64 `json:"estudiantes"`
	Participations int64 `json:"participaciones"`
	Projects       int64 `json:"proyectos"`
	Internal       int64 `json:"internas"`
	External       int64 `json:"externas"`
	Completed      int64 `json:"cumplimientos_conocidos"`
}

type Row struct {
	Carnet     string  `json:"carnet"`
	Student    string  `json:"estudiante"`
	Project    string  `json:"proyecto"`
	Date       string  `json:"fecha_horas"`
	Hours      int32   `json:"horas"`
	Nature     string  `json:"naturaleza"`
	Faculty    *string `json:"facultad"`
	Career     *string `json:"carrera"`
	Type       *string `json:"tipo"`
	Area       *string `json:"area"`
	LegacyArea *string `json:"area_css"`
	District   *string `json:"distrito"`
	State      string  `json:"estado"`
}

type Report struct {
	Summary Summary `json:"resumen"`
	Rows    []Row   `json:"movimientos"`
	Total   int64   `json:"total"`
}

type Repository struct{ pool *pgxpool.Pool }

func New(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

// All local calendar days are evaluated as half-open intervals. The legacy
// fields remain separate from canonically attributed dimensions.
const population = `
WITH filters AS (
	SELECT nullif($1,'')::date date_from,nullif($2,'')::date date_to,$3::text state,
		nullif($4,'')::uuid department_id,nullif($5,'')::uuid municipality_id,nullif($6,'')::uuid district_id,
		nullif($7,'')::uuid career_id,nullif($8,'')::uuid type_id,$9::text[] areas
), movement AS (
	SELECT h.id,h.enrollment_id,e.student_id,e.project_id,h.horas,h.naturaleza,h.fecha_horas,
		h.fecha_horas::timestamp AT TIME ZONE 'America/El_Salvador' day_start,
		(h.fecha_horas+1)::timestamp AT TIME ZONE 'America/El_Salvador' day_end
	FROM public.hour_movements h JOIN public.project_enrollments e ON e.id=h.enrollment_id
	CROSS JOIN filters f
	WHERE (f.date_from IS NULL OR h.fecha_horas>=f.date_from) AND (f.date_to IS NULL OR h.fecha_horas<=f.date_to)
), filtered AS (
	SELECT m.*,s.carnet,s.nombre_completo,p.titulo,p.estado,p.atributos->>'area_css' legacy_area,
		ca.nombre career,fa.nombre faculty,t.nombre project_type,a.nombre area,d.nombre district
	FROM movement m JOIN public.students s ON s.id=m.student_id
	JOIN public.projects p ON p.id=m.project_id
	LEFT JOIN public.geo_districts d ON d.id=p.district_id
	LEFT JOIN public.geo_municipalities mu ON mu.id=d.municipality_id
	CROSS JOIN filters f
	LEFT JOIN LATERAL (
		SELECT c.career_id,c.faculty_id FROM public.student_career_periods c
		WHERE c.student_id=m.student_id AND c.valid_from<=m.day_start
		  AND (c.valid_until IS NULL OR c.valid_until>=m.day_end)
		  AND NOT EXISTS (
			SELECT FROM public.student_career_periods other
			WHERE other.student_id=c.student_id AND other.valid_from IS DISTINCT FROM c.valid_from
			  AND (other.valid_from IS NULL OR other.valid_from<m.day_end)
			  AND (other.valid_until IS NULL OR other.valid_until>m.day_start))
		ORDER BY c.valid_from DESC LIMIT 1
	) ah ON true
	LEFT JOIN public.careers ca ON ca.id=ah.career_id
	LEFT JOIN public.faculties fa ON fa.id=ah.faculty_id
	LEFT JOIN LATERAL (
		SELECT c.tipo_id,c.area_id FROM horas_sociales.proyecto_clasificacion c
		WHERE c.proyecto_id=m.project_id AND c.vigente_desde<=m.day_start
		  AND NOT EXISTS (SELECT FROM horas_sociales.proyecto_clasificacion other
			WHERE other.proyecto_id=m.project_id AND other.vigente_desde>m.day_start AND other.vigente_desde<m.day_end)
		ORDER BY c.vigente_desde DESC LIMIT 1
	) ph ON true
	LEFT JOIN horas_sociales.tipo_proyecto t ON t.id=ph.tipo_id
	LEFT JOIN horas_sociales.area a ON a.id=ph.area_id
	WHERE (f.state='' OR p.estado=f.state)
	  AND (f.department_id IS NULL OR mu.department_id=f.department_id)
	  AND (f.municipality_id IS NULL OR d.municipality_id=f.municipality_id)
	  AND (f.district_id IS NULL OR d.id=f.district_id)
	  AND (f.career_id IS NULL OR ah.career_id=f.career_id)
	  AND (f.type_id IS NULL OR ph.tipo_id=f.type_id)
	  AND (cardinality(f.areas)=0 OR ph.area_id::text=ANY(f.areas))
)`

const summarySQL = population + `
, cohort AS (SELECT DISTINCT student_id FROM filtered),
daily AS (
	SELECT c.student_id,h.fecha_horas,
		coalesce(sum(h.horas) FILTER (WHERE h.naturaleza='INTERNA'),0) internal,
		coalesce(sum(h.horas) FILTER (WHERE h.naturaleza='EXTERNA'),0) external
	FROM cohort c JOIN public.project_enrollments e ON e.student_id=c.student_id
	JOIN public.hour_movements h ON h.enrollment_id=e.id
	GROUP BY c.student_id,h.fecha_horas
), cumulative AS (
	SELECT student_id,fecha_horas,sum(internal) OVER days internal,sum(external) OVER days external
	FROM daily WINDOW days AS (PARTITION BY student_id ORDER BY fecha_horas ROWS UNBOUNDED PRECEDING)
), first_completion AS (
	SELECT student_id,min(fecha_horas) completed_at
	FROM cumulative WHERE internal>=300 AND external>=300 GROUP BY student_id
)
SELECT count(*),count(DISTINCT student_id),count(DISTINCT enrollment_id),count(DISTINCT project_id),
	coalesce(sum(horas) FILTER (WHERE naturaleza='INTERNA'),0),
	coalesce(sum(horas) FILTER (WHERE naturaleza='EXTERNA'),0),
	(SELECT count(*) FROM first_completion c CROSS JOIN filters f
	  WHERE (f.date_from IS NULL OR c.completed_at>=f.date_from)
	    AND (f.date_to IS NULL OR c.completed_at<=f.date_to))
FROM filtered`

func (r *Repository) Query(ctx context.Context, f Filter, limit int, offset int64, export bool) (Report, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return Report{}, err
	}
	defer tx.Rollback(ctx)
	if f.Areas == nil {
		f.Areas = []string{}
	}
	args := []any{f.From, f.To, f.State, f.Department, f.Municipality, f.District, f.Career, f.Type, f.Areas}
	report := Report{Rows: make([]Row, 0)}
	err = tx.QueryRow(ctx, summarySQL, args...).Scan(&report.Total, &report.Summary.Students, &report.Summary.Participations,
		&report.Summary.Projects, &report.Summary.Internal, &report.Summary.External, &report.Summary.Completed)
	if err != nil {
		return Report{}, err
	}
	if export && report.Total > ExportLimit {
		return report, tx.Commit(ctx)
	}
	rows, err := tx.Query(ctx, population+`
	SELECT carnet,nombre_completo,titulo,fecha_horas::text,horas,naturaleza,faculty,career,
		project_type,area,legacy_area,district,estado FROM filtered
	ORDER BY fecha_horas DESC,id LIMIT $10 OFFSET $11`, append(args, limit, offset)...)
	if err != nil {
		return Report{}, err
	}
	for rows.Next() {
		var row Row
		err = rows.Scan(&row.Carnet, &row.Student, &row.Project, &row.Date, &row.Hours, &row.Nature, &row.Faculty,
			&row.Career, &row.Type, &row.Area, &row.LegacyArea, &row.District, &row.State)
		if err != nil {
			rows.Close()
			return Report{}, err
		}
		report.Rows = append(report.Rows, row)
	}
	if err = rows.Err(); err != nil {
		return Report{}, err
	}
	return report, tx.Commit(ctx)
}
