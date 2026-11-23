package httpserver

import (
	"context"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"horas-sociales/api/internal/modules/participaciones"
	"horas-sociales/api/internal/platform/auth"
)

func CierresRoutes(repo *participaciones.Repository, enabled bool) func(chi.Router) {
	return func(r chi.Router) {
		r.Post("/proyectos/{id}/cierre", closeEntity(repo, enabled, false))
		r.Post("/proyectos/{id}/participaciones/{participacion}/cierre", closeEntity(repo, enabled, true))
	}
}

func closeEntity(repo *participaciones.Repository, enabled, individual bool) http.HandlerFunc {
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
		var enrollment pgtype.UUID
		if individual {
			enrollment, err = catalogoUUID(chi.URLParam(r, "participacion"))
			if err != nil {
				WriteInputError(w, r, err)
				return
			}
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
			Action string `json:"accion"`
			Date   string `json:"fecha_cierre"`
			Reason string `json:"motivo"`
		}
		if err := ReadJSON(w, r, &body); err != nil {
			WriteInputError(w, r, err)
			return
		}
		alternative := "cancelar"
		if individual {
			alternative = "retirar"
		}
		if body.Action != "finalizar" && body.Action != alternative {
			writeFieldError(w, r, "bad_request", "accion")
			return
		}
		if date, err := time.Parse(time.DateOnly, body.Date); err != nil || date.Year() < 1 {
			writeFieldError(w, r, "bad_request", "fecha_cierre")
			return
		}
		text := strings.TrimSpace(body.Reason)
		if (body.Action == alternative && text == "") || strings.ContainsRune(text, 0) || utf8.RuneCountInString(text) > 500 {
			writeFieldError(w, r, "bad_request", "motivo")
			return
		}
		var reason *string
		if text != "" {
			reason = &text
		}
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		result, err := repo.Close(ctx, project, enrollment, actor, participaciones.CloseInput{
			Action: body.Action, Date: body.Date, Reason: reason,
		})
		writeParticipation(w, r, http.StatusOK, result, err)
	}
}
