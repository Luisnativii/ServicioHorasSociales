package httpserver

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"horas-sociales/api/internal/modules/catalogos"
)

type listarTerritorio func(context.Context, int, int64, pgtype.UUID) ([]catalogos.LugarTerritorial, int64, error)

func CatalogoTerritorialRoutes(repo *catalogos.TerritorioRepository) func(chi.Router) {
	return func(r chi.Router) {
		r.Get("/catalogos/departamentos", listadoTerritorial("", repo.ListDepartamentos))
		r.Get("/catalogos/municipios", listadoTerritorial("departamento_id", repo.ListMunicipios))
		r.Get("/catalogos/distritos", listadoTerritorial("municipio_id", repo.ListDistritos))
	}
}

func listadoTerritorial(parametroPadre string, listar listarTerritorio) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			WriteError(w, r, "bad_request")
			return
		}
		for key, values := range query {
			if len(values) != 1 || (key != "page" && key != "page_size" && key != parametroPadre) {
				WriteError(w, r, "bad_request")
				return
			}
		}
		p, err := ParsePagination(query, map[string]string{"codigo": "codigo"}, "codigo")
		if err != nil {
			WriteInputError(w, r, err)
			return
		}
		var padre pgtype.UUID
		if parametroPadre != "" {
			padre, err = catalogoUUID(query.Get(parametroPadre))
			if err != nil {
				WriteInputError(w, r, err)
				return
			}
		}

		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		lugares, total, err := listar(ctx, p.PageSize, p.Offset, padre)
		if err != nil {
			errorCatalogo(w, r, err)
			return
		}
		WriteJSON(w, r, http.StatusOK, struct {
			Lugares []catalogos.LugarTerritorial `json:"lugares"`
			Pagina  paginaCatalogo               `json:"pagina"`
		}{lugares, paginaCatalogo{p.Page, p.PageSize, total}})
	}
}
