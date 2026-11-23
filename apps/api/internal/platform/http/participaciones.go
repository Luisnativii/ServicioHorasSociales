package httpserver

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"horas-sociales/api/internal/modules/participaciones"
	"horas-sociales/api/internal/platform/auth"
)

func ParticipacionesRoutes(repo *participaciones.Repository, writes bool) func(chi.Router) {
	return func(r chi.Router) {
		r.Get("/proyectos/{id}/participaciones", listParticipations(repo, true))
		r.Get("/estudiantes/{id}/participaciones", listParticipations(repo, false))
		r.Post("/proyectos/{id}/participaciones", func(w http.ResponseWriter, r *http.Request) {
			if !writes {
				writeFieldError(w, r, "service_unavailable", "escritura_core")
				return
			}
			project, err := catalogoUUID(chi.URLParam(r, "id"))
			if err != nil {
				WriteInputError(w, r, err)
				return
			}
			identity, ok := auth.FromContext(r.Context())
			if !ok {
				WriteError(w, r, "unauthorized")
				return
			}
			actor, err := catalogoUUID(identity.ID)
			if err != nil {
				WriteError(w, r, "unauthorized")
				return
			}
			var body struct {
				Student   string `json:"estudiante_id"`
				Confirmed bool   `json:"confirmar_exceso_cupo"`
			}
			if err := ReadJSON(w, r, &body); err != nil {
				WriteInputError(w, r, err)
				return
			}
			student, err := catalogoUUID(body.Student)
			if err != nil {
				writeFieldError(w, r, "bad_request", "estudiante_id")
				return
			}
			ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
			defer cancel()
			id, err := repo.Assign(ctx, project, student, actor, body.Confirmed)
			writeParticipation(w, r, http.StatusCreated, struct {
				ID string `json:"id"`
			}{id}, err)
		})
	}
}

func listParticipations(repo *participaciones.Repository, byProject bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := catalogoUUID(chi.URLParam(r, "id"))
		if err != nil {
			WriteInputError(w, r, err)
			return
		}
		var project, student pgtype.UUID
		if byProject {
			project = id
		} else {
			student = id
		}
		query, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			WriteError(w, r, "bad_request")
			return
		}
		for key := range query {
			if key != "page" && key != "page_size" {
				WriteError(w, r, "bad_request")
				return
			}
		}
		page, err := ParsePagination(query, map[string]string{"asignada_en": "asignada_en"}, "asignada_en")
		if err != nil {
			WriteInputError(w, r, err)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		items, total, err := repo.List(ctx, project, student, page.PageSize, page.Offset)
		writeParticipation(w, r, http.StatusOK, struct {
			Items  []participaciones.Item `json:"participaciones"`
			Pagina paginaCatalogo         `json:"pagina"`
		}{items, paginaCatalogo{page.Page, page.PageSize, total}}, err)
	}
}

func writeParticipation(w http.ResponseWriter, r *http.Request, status int, item any, err error) {
	switch {
	case err == nil:
		WriteJSON(w, r, status, item)
	case errors.Is(err, pgx.ErrNoRows):
		WriteError(w, r, "not_found")
	case errors.Is(err, participaciones.ErrProjectClosed):
		writeFieldError(w, r, "conflict", "estado_proyecto")
	case errors.Is(err, participaciones.ErrStudentInactive):
		writeFieldError(w, r, "conflict", "estado_estudiante")
	case errors.Is(err, participaciones.ErrDuplicate):
		writeFieldError(w, r, "conflict", "participacion")
	case errors.Is(err, participaciones.ErrClosedParticipation):
		writeFieldError(w, r, "conflict", "estado_participacion")
	case errors.Is(err, participaciones.ErrInvalidCloseDate):
		writeFieldError(w, r, "bad_request", "fecha_cierre")
	case errors.Is(err, participaciones.ErrCannotFinish):
		writeFieldError(w, r, "conflict", "participaciones_abiertas")
	case errors.Is(err, participaciones.ErrCapacity):
		writeFieldError(w, r, "conflict", "cupo")
	default:
		WriteError(w, r, "service_unavailable")
	}
}
