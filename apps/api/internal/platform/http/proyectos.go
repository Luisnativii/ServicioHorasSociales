package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"horas-sociales/api/internal/modules/proyectos"
	"horas-sociales/api/internal/platform/auth"
)

func ProyectosRoutes(repo *proyectos.Repository, writes bool) func(chi.Router) {
	return func(r chi.Router) {
		r.Get("/proyectos", listProjects(repo))
		r.Get("/proyectos/{id}", func(w http.ResponseWriter, r *http.Request) {
			id, err := catalogoUUID(chi.URLParam(r, "id"))
			if err != nil {
				WriteInputError(w, r, err)
				return
			}
			ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
			defer cancel()
			item, err := repo.Get(ctx, id.String())
			writeProject(w, r, http.StatusOK, item, err)
		})
		r.Post("/proyectos", saveProject(repo, writes, true))
		r.Put("/proyectos/{id}", saveProject(repo, writes, false))
	}
}

func listProjects(repo *proyectos.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			WriteError(w, r, "bad_request")
			return
		}
		areas := make([]string, 0)
		for key, values := range q {
			switch key {
			case "page", "page_size":
				continue
			case "q", "estado":
				if len(values) == 1 && textoClasificacionValido(values[0]) {
					continue
				}
			case "institucion_id":
				if len(values) == 1 {
					if _, err := catalogoUUID(values[0]); err == nil {
						continue
					}
				}
			case "area_id":
				if len(values) <= 50 {
					for _, value := range values {
						id, err := catalogoUUID(value)
						if err != nil {
							writeFieldError(w, r, "bad_request", key)
							return
						}
						areas = append(areas, id.String())
					}
					continue
				}
			}
			writeFieldError(w, r, "bad_request", key)
			return
		}
		p, err := ParsePagination(q, map[string]string{"titulo": "titulo"}, "titulo")
		if err != nil {
			WriteInputError(w, r, err)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		results, total, err := repo.List(ctx, strings.TrimSpace(q.Get("q")), strings.ToLower(q.Get("institucion_id")), q.Get("estado"), areas, p.PageSize, p.Offset)
		writeProject(w, r, http.StatusOK, struct {
			Proyectos []json.RawMessage `json:"proyectos"`
			Pagina    paginaCatalogo    `json:"pagina"`
		}{results, paginaCatalogo{p.Page, p.PageSize, total}}, err)
	}
}

func saveProject(repo *proyectos.Repository, writes, creating bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !writes {
			writeFieldError(w, r, "service_unavailable", "escritura_core")
			return
		}
		actor, ok := auth.FromContext(r.Context())
		if !ok {
			WriteError(w, r, "unauthorized")
			return
		}
		actorID, err := catalogoUUID(actor.ID)
		if err != nil {
			WriteError(w, r, "unauthorized")
			return
		}
		var id string
		if !creating {
			item, err := catalogoUUID(chi.URLParam(r, "id"))
			if err != nil {
				WriteInputError(w, r, err)
				return
			}
			id = item.String()
		}
		var input proyectos.Input
		if err := ReadJSON(w, r, &input); err != nil {
			WriteInputError(w, r, err)
			return
		}
		if field := validateProject(&input); field != "" {
			writeFieldError(w, r, "bad_request", field)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		id, err = repo.Save(ctx, id, actorID.String(), input)
		if err != nil {
			writeProject(w, r, 0, nil, err)
			return
		}
		item, err := repo.Get(ctx, id)
		status := http.StatusOK
		if creating {
			status = http.StatusCreated
		}
		writeProject(w, r, status, item, err)
	}
}

func validateProject(input *proyectos.Input) string {
	input.Titulo = strings.TrimSpace(input.Titulo)
	input.Descripcion = strings.TrimSpace(input.Descripcion)
	if input.Titulo == "" || len([]rune(input.Titulo)) > 300 || strings.ContainsRune(input.Titulo, 0) {
		return "titulo"
	}
	if len([]rune(input.Descripcion)) > 10000 || strings.ContainsRune(input.Descripcion, 0) {
		return "descripcion"
	}
	for _, entry := range []struct {
		field string
		value *string
	}{
		{"institucion_id", &input.Institution}, {"distrito_id", &input.District},
		{"tipo_id", &input.Tipo}, {"area_id", &input.Area}, {"modalidad_id", &input.Modalidad},
	} {
		id, err := catalogoUUID(*entry.value)
		if err != nil {
			return entry.field
		}
		*entry.value = id.String()
	}
	if input.Naturaleza == "" {
		input.Naturaleza = "EXTERNA"
	}
	if input.Naturaleza != "EXTERNA" && input.Naturaleza != "INTERNA" {
		return "naturaleza"
	}
	if input.Cupo != nil && *input.Cupo <= 0 {
		return "cupo"
	}
	if input.Inicio != nil {
		if _, err := time.Parse(time.DateOnly, *input.Inicio); err != nil {
			return "fecha_inicio"
		}
	}
	c := &input.Responsable
	c.Nombre = strings.TrimSpace(c.Nombre)
	contact := contactoInstitucion{Nombre: &c.Nombre, Correo: c.Correo, Telefono: c.Telefono, Extension: c.Extension}
	if invalid := contact.campoInvalido(); invalid != "" {
		return "responsable." + strings.TrimPrefix(invalid, "contacto.")
	}
	return ""
}

func writeProject(w http.ResponseWriter, r *http.Request, status int, item any, err error) {
	var field proyectos.FieldError
	var pgErr *pgconn.PgError
	switch {
	case err == nil:
		WriteJSON(w, r, status, item)
	case errors.Is(err, pgx.ErrNoRows):
		WriteError(w, r, "not_found")
	case errors.As(err, &field):
		writeFieldError(w, r, "bad_request", string(field))
	case errors.As(err, &pgErr) && pgErr.Code == "23505":
		writeFieldError(w, r, "conflict", "titulo")
	case pgErr != nil && pgErr.ConstraintName == "ck_projects_fechas":
		writeFieldError(w, r, "bad_request", "fecha_inicio")
	default:
		WriteError(w, r, "service_unavailable")
	}
}
