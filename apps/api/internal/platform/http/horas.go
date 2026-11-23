package httpserver

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	"horas-sociales/api/internal/modules/horas"
	"horas-sociales/api/internal/platform/auth"
)

func HorasRoutes(repo *horas.Repository, enabled bool) func(chi.Router) {
	return func(r chi.Router) {
		r.Get("/estudiantes/{id}/horas", listHours(repo))
		r.Post("/proyectos/{id}/participaciones/{participacion}/horas", registerHours(repo, enabled))
	}
}

func listHours(repo *horas.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		student, err := catalogoUUID(chi.URLParam(r, "id"))
		if err != nil {
			WriteInputError(w, r, err)
			return
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
		page, err := ParsePagination(query, map[string]string{"fecha_horas": "fecha_horas"}, "fecha_horas")
		if err != nil {
			WriteInputError(w, r, err)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		items, progress, total, err := repo.List(ctx, student, page.PageSize, page.Offset)
		writeHours(w, r, http.StatusOK, struct {
			Items    []horas.Movement `json:"movimientos"`
			Progress horas.Progress   `json:"avance"`
			Page     paginaCatalogo   `json:"pagina"`
		}{items, progress, paginaCatalogo{page.Page, page.PageSize, total}}, err)
	}
}

func registerHours(repo *horas.Repository, enabled bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !enabled {
			writeFieldError(w, r, "service_unavailable", "escritura_core")
			return
		}
		project, err := catalogoUUID(chi.URLParam(r, "id"))
		if err != nil {
			WriteInputError(w, r, err)
			return
		}
		enrollment, err := catalogoUUID(chi.URLParam(r, "participacion"))
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
			Hours   int32  `json:"horas"`
			Date    string `json:"fecha_horas"`
			Concept string `json:"concepto"`
			Request string `json:"solicitud_id"`
		}
		if err := ReadJSON(w, r, &body); err != nil {
			WriteInputError(w, r, err)
			return
		}
		if body.Hours <= 0 {
			writeFieldError(w, r, "bad_request", "horas")
			return
		}
		if date, err := time.Parse(time.DateOnly, body.Date); err != nil || date.Year() < 1 {
			writeFieldError(w, r, "bad_request", "fecha_horas")
			return
		}
		body.Concept = strings.TrimSpace(body.Concept)
		if body.Concept == "" || utf8.RuneCountInString(body.Concept) > 500 || strings.ContainsRune(body.Concept, 0) {
			writeFieldError(w, r, "bad_request", "concepto")
			return
		}
		key, err := catalogoUUID(body.Request)
		if err != nil {
			writeFieldError(w, r, "bad_request", "solicitud_id")
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		item, err := repo.Register(ctx, project, enrollment, actor, horas.Credit{Hours: body.Hours, Date: body.Date, Concept: body.Concept, Request: key})
		status := http.StatusCreated
		if item.Replayed {
			status = http.StatusOK
		}
		writeHours(w, r, status, item, err)
	}
}

func writeHours(w http.ResponseWriter, r *http.Request, status int, result any, err error) {
	switch {
	case err == nil:
		WriteJSON(w, r, status, result)
	case errors.Is(err, pgx.ErrNoRows):
		WriteError(w, r, "not_found")
	case errors.Is(err, horas.ErrClosed):
		writeFieldError(w, r, "conflict", "estado_participacion")
	case errors.Is(err, horas.ErrDate):
		writeFieldError(w, r, "bad_request", "fecha_horas")
	case errors.Is(err, horas.ErrConflict):
		writeFieldError(w, r, "conflict", "solicitud_id")
	default:
		WriteError(w, r, "service_unavailable")
	}
}
