package proyectos

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrTerminal = errors.New("proyecto cerrado")

type FieldError string

func (e FieldError) Error() string { return string(e) }

type Contact struct {
	Nombre    string  `json:"nombre"`
	Correo    *string `json:"correo"`
	Telefono  *string `json:"telefono"`
	Extension *string `json:"extension"`
}

type Input struct {
	Titulo      string  `json:"titulo"`
	Descripcion string  `json:"descripcion"`
	Institution string  `json:"institucion_id"`
	District    string  `json:"distrito_id"`
	Naturaleza  string  `json:"naturaleza"`
	Inicio      *string `json:"fecha_inicio"`
	Cupo        *int32  `json:"cupo"`
	Tipo        string  `json:"tipo_id"`
	Area        string  `json:"area_id"`
	Modalidad   string  `json:"modalidad_id"`
	Responsable Contact `json:"responsable"`
}

type Repository struct{ pool *pgxpool.Pool }

func New(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

const selectProject = `
SELECT jsonb_build_object(
  'id',p.id,'titulo',p.titulo,'descripcion',p.descripcion,'correlativo_css',p.correlativo_css,
  'naturaleza',p.naturaleza,'estado',p.estado,'fecha_inicio',p.fecha_inicio,'fecha_cierre',p.fecha_cierre,
  'cerrado_en',c.cerrada_en,'cerrado_por',c.actor_id,'motivo_cierre',c.motivo,
  'cupo',p.cupos,'area_css',p.atributos->>'area_css','ubicacion_heredada',p.ubicacion,
  'institucion_id',i.id,'institucion',i.nombre,'sede',sede.nombre,'distrito_id',p.district_id,'distrito',d.nombre,
  'tipo_id',t.id,'tipo',t.nombre,'area_id',a.id,'area',a.nombre,'modalidad_id',m.id,'modalidad',m.nombre,
  'responsable',CASE WHEN r.nombre IS NOT NULL THEN jsonb_build_object(
    'nombre',r.nombre,'correo',r.correo,'telefono',r.telefono,'extension',r.extension) END)
FROM public.projects p JOIN public.institutions i ON i.id=p.institution_id
LEFT JOIN public.geo_districts sede ON sede.id=i.district_id
LEFT JOIN public.geo_districts d ON d.id=p.district_id
LEFT JOIN LATERAL (
  SELECT tipo_id, area_id, modalidad_id FROM horas_sociales.proyecto_clasificacion
  WHERE proyecto_id=p.id ORDER BY vigente_desde DESC LIMIT 1
) cfg ON true
LEFT JOIN horas_sociales.tipo_proyecto t ON t.id=cfg.tipo_id
LEFT JOIN horas_sociales.area a ON a.id=cfg.area_id
LEFT JOIN horas_sociales.modalidad m ON m.id=cfg.modalidad_id
LEFT JOIN horas_sociales.proyecto_responsable r ON r.proyecto_id=p.id
LEFT JOIN horas_sociales.cierre_evento c ON c.proyecto_id=p.id AND c.participacion_id IS NULL`

func (r *Repository) Get(ctx context.Context, id string) (json.RawMessage, error) {
	var result json.RawMessage
	err := r.pool.QueryRow(ctx, selectProject+` WHERE p.id=$1::text::uuid`, id).Scan(&result)
	return result, err
}

func (r *Repository) List(ctx context.Context, q, institution, state string, areas []string, limit int, offset int64) ([]json.RawMessage, int64, error) {
	const where = ` WHERE ($1='' OR strpos(lower(p.titulo),lower($1))>0 OR strpos(lower(i.nombre),lower($1))>0
	    OR strpos(coalesce(p.correlativo_css::text,''),$1)>0)
	  AND ($2='' OR i.id::text=$2) AND ($3='' OR p.estado=$3)
	  AND (cardinality($4::text[])=0 OR a.id::text=ANY($4::text[]))`
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, 0, err
	}
	defer tx.Rollback(ctx)
	items, err := tx.Query(ctx, selectProject+where+` ORDER BY lower(p.titulo),p.id LIMIT $5 OFFSET $6`, q, institution, state, areas, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	rows, err := pgx.CollectRows(items, pgx.RowTo[json.RawMessage])
	if err != nil {
		return nil, 0, err
	}
	var total int64
	err = tx.QueryRow(ctx, `SELECT count(*) FROM (`+selectProject+where+`) filtrados`, q, institution, state, areas).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	return rows, total, tx.Commit(ctx)
}

func (r *Repository) Save(ctx context.Context, id, actor string, input Input) (string, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)
	creating := id == ""
	var oldType, oldArea, oldMode string
	if !creating {
		var institution, state string
		err = tx.QueryRow(ctx, `SELECT institution_id::text, estado FROM public.projects WHERE id=$1::text::uuid FOR UPDATE`, id).Scan(&institution, &state)
		if err != nil {
			return "", err
		}
		if state == "FINALIZADO" || state == "CANCELADO" || state == "ABANDONADO" {
			return "", ErrTerminal
		}
		if institution != input.Institution {
			return "", FieldError("institucion_id")
		}
		err = tx.QueryRow(ctx, `SELECT tipo_id::text,area_id::text,modalidad_id::text
		  FROM horas_sociales.proyecto_clasificacion WHERE proyecto_id=$1::text::uuid ORDER BY vigente_desde DESC LIMIT 1`, id).
			Scan(&oldType, &oldArea, &oldMode)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return "", err
		}
	}
	if creating {
		var active bool
		err = tx.QueryRow(ctx, `SELECT activo FROM public.institutions WHERE id=$1::text::uuid FOR SHARE`, input.Institution).Scan(&active)
		if errors.Is(err, pgx.ErrNoRows) || (err == nil && !active) {
			return "", FieldError("institucion_id")
		}
		if err != nil {
			return "", err
		}
	}
	var district bool
	err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT FROM public.geo_districts WHERE id=$1::text::uuid)`, input.District).Scan(&district)
	if err != nil {
		return "", err
	}
	if !district {
		return "", FieldError("distrito_id")
	}

	for _, item := range []struct{ table, field, current, previous string }{
		{"tipo_proyecto", "tipo_id", input.Tipo, oldType}, {"area", "area_id", input.Area, oldArea}, {"modalidad", "modalidad_id", input.Modalidad, oldMode},
	} {
		if item.current == item.previous {
			continue
		}
		var active bool
		err = tx.QueryRow(ctx, `SELECT activo FROM horas_sociales.`+item.table+` WHERE id=$1::text::uuid FOR SHARE`, item.current).Scan(&active)
		if errors.Is(err, pgx.ErrNoRows) || (err == nil && !active) {
			return "", FieldError(item.field)
		}
		if err != nil {
			return "", err
		}
	}
	if oldMode != "" && input.Modalidad != oldMode {
		var hasHours bool
		err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT FROM public.hour_movements h JOIN public.project_enrollments e
		  ON e.id=h.enrollment_id WHERE e.project_id=$1::text::uuid)`, id).Scan(&hasHours)
		if err != nil {
			return "", err
		}
		if hasHours {
			return "", FieldError("modalidad_id")
		}
	}
	if _, err = tx.Exec(ctx, `SELECT set_config('app.actor',$1,true)`, "admin:"+actor); err != nil {
		return "", err
	}
	if creating {
		err = tx.QueryRow(ctx, `INSERT INTO public.projects
		  (titulo,descripcion,institution_id,district_id,naturaleza,fecha_inicio,cupos,estado)
		  VALUES ($1,$2,$3::text::uuid,$4::text::uuid,$5,$6::text::date,$7,'ACTIVO') RETURNING id::text`,
			input.Titulo, input.Descripcion, input.Institution, input.District, input.Naturaleza, input.Inicio, input.Cupo).Scan(&id)
	} else {
		_, err = tx.Exec(ctx, `UPDATE public.projects SET titulo=$2,descripcion=$3,district_id=$4::text::uuid,
		  naturaleza=$5,fecha_inicio=$6::text::date,cupos=$7 WHERE id=$1::text::uuid`, id, input.Titulo,
			input.Descripcion, input.District, input.Naturaleza, input.Inicio, input.Cupo)
	}
	if err != nil {
		return "", err
	}
	if input.Tipo != oldType || input.Area != oldArea || input.Modalidad != oldMode {
		_, err = tx.Exec(ctx, `INSERT INTO horas_sociales.proyecto_clasificacion(proyecto_id,tipo_id,area_id,modalidad_id)
		  VALUES($1::text::uuid,$2::text::uuid,$3::text::uuid,$4::text::uuid)`, id, input.Tipo, input.Area, input.Modalidad)
		if err != nil {
			return "", err
		}
	}
	c := input.Responsable
	_, err = tx.Exec(ctx, `INSERT INTO horas_sociales.proyecto_responsable(proyecto_id,nombre,correo,telefono,extension)
	  VALUES($1::text::uuid,$2,$3,$4,$5) ON CONFLICT(proyecto_id) DO UPDATE SET nombre=excluded.nombre,
	  correo=excluded.correo,telefono=excluded.telefono,extension=excluded.extension`, id, c.Nombre, c.Correo, c.Telefono, c.Extension)
	if err != nil {
		return "", err
	}
	return id, tx.Commit(ctx)
}
