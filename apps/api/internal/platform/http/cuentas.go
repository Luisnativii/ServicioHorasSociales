package httpserver

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"horas-sociales/api/internal/modules/admin"
	"horas-sociales/api/internal/platform/auth"
)

type cuentaRespuesta struct {
	ID     string `json:"id"`
	Correo string `json:"correo"`
	Activo bool   `json:"activo"`
}

type listarCuentasRespuesta struct {
	Cuentas []cuentaRespuesta `json:"cuentas"`
	Pagina  paginaRespuesta   `json:"pagina"`
}

type paginaRespuesta struct {
	Pagina   int64 `json:"pagina"`
	PageSize int   `json:"page_size"`
	Total    int64 `json:"total"`
}

func cuentaDesdeModelo(a admin.Administrator) cuentaRespuesta {
	return cuentaRespuesta{ID: a.ID.String(), Correo: a.Correo, Activo: a.Activo}
}

// listarCuentasHandler implementa GET /api/v1/admin/cuentas.
// Requiere JWT; no exige CSRF porque es una consulta de solo lectura.
func listarCuentasHandler(buscar administradorBuscador) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		paginacion, err := ParsePagination(r.URL.Query(), map[string]string{"creado_en": "creado_en"}, "creado_en")
		if err != nil {
			WriteInputError(w, r, err)
			return
		}

		cuentas, err := buscar.List(ctx, int32(paginacion.PageSize), paginacion.Offset)
		if err != nil {
			WriteError(w, r, "internal_error")
			return
		}

		total, err := buscar.Count(ctx)
		if err != nil {
			WriteError(w, r, "internal_error")
			return
		}

		resultado := make([]cuentaRespuesta, 0, len(cuentas))
		for _, c := range cuentas {
			resultado = append(resultado, cuentaDesdeModelo(c))
		}

		w.Header().Set("Cache-Control", "no-store")
		WriteJSON(w, r, http.StatusOK, listarCuentasRespuesta{
			Cuentas: resultado,
			Pagina: paginaRespuesta{
				Pagina:   paginacion.Page,
				PageSize: paginacion.PageSize,
				Total:    int64(total),
			},
		})
	}
}

// cambiarEstadoHandler implementa POST /api/v1/admin/cuentas/{id}/activar
// y desactivar. Requiere JWT, Origin y prueba CSRF firmada.
func cambiarEstadoHandler(buscar administradorBuscador, activar bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		identidad, ok := auth.FromContext(r.Context())
		if !ok {
			WriteError(w, r, "unauthorized")
			return
		}

		idStr := chi.URLParam(r, "id")
		if !auth.ValidSubject(idStr) {
			WriteError(w, r, "bad_request")
			return
		}

		var id pgtype.UUID
		if err := id.Scan(idStr); err != nil {
			WriteError(w, r, "bad_request")
			return
		}

		if _, err := buscar.FindByID(ctx, id); err != nil {
			WriteError(w, r, "not_found")
			return
		}

		actualizada, err := buscar.SetActive(ctx, id, activar)
		if err != nil {
			WriteError(w, r, "internal_error")
			return
		}

		// Registrar actor y resultado sin secretos ni hashes.
		slog.Info("cambio de estado de cuenta",
			"actor", identidad.ID,
			"cuenta", idStr,
			"activo", actualizada.Activo,
			"resultado", "ok")

		w.Header().Set("Cache-Control", "no-store")
		WriteJSON(w, r, http.StatusOK, cuentaDesdeModelo(actualizada))
	}
}
