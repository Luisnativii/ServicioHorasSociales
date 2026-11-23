package httpserver

import (
	"bytes"
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/go-chi/chi/v5"

	"horas-sociales/api/internal/modules/reportes"
)

func ReportesRoutes(repo *reportes.Repository) func(chi.Router) {
	return func(r chi.Router) {
		r.Get("/reportes/actividad", activity(repo, false))
		r.Get("/reportes/actividad.xlsx", activity(repo, true))
	}
}

func activity(repo *reportes.Repository, export bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			WriteError(w, r, "bad_request")
			return
		}
		filter, invalid := parseActivityFilters(query, export)
		if invalid != "" {
			writeFieldError(w, r, "bad_request", invalid)
			return
		}
		page, err := ParsePagination(query, map[string]string{"fecha_horas": "fecha_horas"}, "fecha_horas")
		if err != nil {
			WriteInputError(w, r, err)
			return
		}
		limit := page.PageSize
		if export {
			limit = reportes.ExportLimit
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		result, err := repo.Query(ctx, filter, limit, page.Offset, export)
		if err != nil {
			WriteError(w, r, "service_unavailable")
			return
		}
		if !export {
			WriteJSON(w, r, http.StatusOK, struct {
				Summary reportes.Summary `json:"resumen"`
				Rows    []reportes.Row   `json:"movimientos"`
				Page    paginaCatalogo   `json:"pagina"`
			}{result.Summary, result.Rows, paginaCatalogo{page.Page, page.PageSize, result.Total}})
			return
		}
		if result.Total > reportes.ExportLimit {
			writeFieldError(w, r, "payload_too_large", "filtros")
			return
		}
		var content bytes.Buffer
		if err = reportes.WriteXLSX(&content, result, filter); err != nil {
			WriteError(w, r, "internal_error")
			return
		}
		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		w.Header().Set("Content-Disposition", `attachment; filename="actividad-horas-sociales.xlsx"`)
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write(content.Bytes())
	}
}

func parseActivityFilters(query url.Values, export bool) (reportes.Filter, string) {
	var f reportes.Filter
	for name, values := range query {
		if name == "area_id" {
			if len(values) > 50 {
				return f, name
			}
			for _, value := range values {
				id, err := catalogoUUID(value)
				if err != nil {
					return f, name
				}
				f.Areas = append(f.Areas, id.String())
			}
			continue
		}
		if len(values) != 1 || values[0] == "" {
			return f, name
		}
		value := values[0]
		switch name {
		case "page", "page_size":
			if export {
				return f, name
			}
		case "desde", "hasta":
			date, err := time.Parse(time.DateOnly, value)
			if err != nil || date.Year() < 1 {
				return f, name
			}
			if name == "desde" {
				f.From = value
			} else {
				f.To = value
			}
		case "estado":
			switch value {
			case "BORRADOR", "ACTIVO", "PAUSADO", "ABANDONADO", "CANCELADO", "FINALIZADO":
				f.State = value
			default:
				return f, name
			}
		case "departamento_id", "municipio_id", "distrito_id", "carrera_id", "tipo_id":
			id, err := catalogoUUID(value)
			if err != nil {
				return f, name
			}
			targets := map[string]*string{"departamento_id": &f.Department, "municipio_id": &f.Municipality,
				"distrito_id": &f.District, "carrera_id": &f.Career, "tipo_id": &f.Type}
			*targets[name] = id.String()
		default:
			return f, name
		}
	}
	if f.From != "" && f.To != "" && f.From > f.To {
		return f, "hasta"
	}
	return f, ""
}
