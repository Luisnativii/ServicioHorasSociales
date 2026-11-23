package httpserver

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"horas-sociales/api/internal/modules/catalogos"
)

type coordinadorRespuesta struct {
	CarreraID string      `json:"carrera_id"`
	Nombre    pgtype.Text `json:"coordinador_nombre"`
}

type coordinadorOperacion func(http.ResponseWriter, *http.Request, context.Context, pgtype.UUID) (pgtype.Text, error)

func CoordinadorCarreraRoutes(repo *catalogos.CoordinadorRepository) func(chi.Router) {
	return func(r chi.Router) {
		const ruta = "/catalogos/carreras/{id}/coordinador"
		r.Get(ruta, coordinadorHandler(func(_ http.ResponseWriter, _ *http.Request, ctx context.Context, id pgtype.UUID) (pgtype.Text, error) {
			return repo.Find(ctx, id)
		}))
		r.Put(ruta, coordinadorHandler(func(w http.ResponseWriter, r *http.Request, ctx context.Context, id pgtype.UUID) (pgtype.Text, error) {
			var input struct {
				Nombre json.RawMessage `json:"coordinador_nombre"`
			}
			if err := ReadJSON(w, r, &input); err != nil {
				return pgtype.Text{}, err
			}
			var nombre pgtype.Text
			if err := json.Unmarshal(input.Nombre, &nombre); err != nil {
				return pgtype.Text{}, &InputError{Code: "invalid_json"}
			}
			return repo.Set(ctx, id, nombre)
		}))
	}
}

func coordinadorHandler(operar coordinadorOperacion) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := catalogoUUID(chi.URLParam(r, "id"))
		if err != nil {
			WriteInputError(w, r, err)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		nombre, err := operar(w, r, ctx, id)
		if err != nil {
			errorClasificacion(w, r, err)
			return
		}
		WriteJSON(w, r, http.StatusOK, coordinadorRespuesta{CarreraID: id.String(), Nombre: nombre})
	}
}
