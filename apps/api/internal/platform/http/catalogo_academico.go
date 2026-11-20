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
	"github.com/jackc/pgx/v5/pgtype"

	"horas-sociales/api/internal/modules/catalogos"
)

type paginaCatalogo struct {
	Page     int64 `json:"page"`
	PageSize int   `json:"page_size"`
	Total    int64 `json:"total"`
}

type facultadRespuesta struct {
	ID     string `json:"id"`
	Nombre string `json:"nombre"`
	Activo bool   `json:"activo"`
}

type carreraRespuesta struct {
	ID             string `json:"id"`
	FacultadID     string `json:"facultad_id"`
	Nombre         string `json:"nombre"`
	Activo         bool   `json:"activo"`
	FacultadActiva bool   `json:"facultad_activa"`
}

func CatalogoAcademicoRoutes(repo *catalogos.AcademiaRepository) func(chi.Router) {
	return func(r chi.Router) {
		r.Get("/catalogos/facultades", func(w http.ResponseWriter, r *http.Request) {
			p, filtros, _, err := filtrosAcademicos(r.URL.RawQuery, false)
			if err != nil {
				WriteInputError(w, r, err)
				return
			}
			ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
			defer cancel()
			facultades, total, err := repo.ListFacultades(ctx, filtros)
			if err != nil {
				errorCatalogo(w, r, err)
				return
			}
			resultado := make([]facultadRespuesta, 0, len(facultades))
			for _, f := range facultades {
				resultado = append(resultado, facultadDesdeModelo(f))
			}
			WriteJSON(w, r, http.StatusOK, struct {
				Facultades []facultadRespuesta `json:"facultades"`
				Pagina     paginaCatalogo      `json:"pagina"`
			}{resultado, paginaCatalogo{p.Page, p.PageSize, total}})
		})
		r.Get("/catalogos/carreras", func(w http.ResponseWriter, r *http.Request) {
			p, filtros, facultadID, err := filtrosAcademicos(r.URL.RawQuery, true)
			if err != nil {
				WriteInputError(w, r, err)
				return
			}
			ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
			defer cancel()
			carreras, total, err := repo.ListCarreras(ctx, filtros, facultadID)
			if err != nil {
				errorCatalogo(w, r, err)
				return
			}
			resultado := make([]carreraRespuesta, 0, len(carreras))
			for _, c := range carreras {
				resultado = append(resultado, carreraDesdeModelo(c))
			}
			WriteJSON(w, r, http.StatusOK, struct {
				Carreras []carreraRespuesta `json:"carreras"`
				Pagina   paginaCatalogo     `json:"pagina"`
			}{resultado, paginaCatalogo{p.Page, p.PageSize, total}})
		})
		r.Get("/catalogos/facultades/{id}", func(w http.ResponseWriter, r *http.Request) {
			id, err := catalogoUUID(chi.URLParam(r, "id"))
			if err != nil {
				WriteInputError(w, r, err)
				return
			}
			ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
			defer cancel()
			f, err := repo.FindFacultad(ctx, id)
			if err != nil {
				errorCatalogo(w, r, err)
				return
			}
			WriteJSON(w, r, http.StatusOK, facultadDesdeModelo(f))
		})
		r.Get("/catalogos/carreras/{id}", func(w http.ResponseWriter, r *http.Request) {
			id, err := catalogoUUID(chi.URLParam(r, "id"))
			if err != nil {
				WriteInputError(w, r, err)
				return
			}
			ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
			defer cancel()
			c, err := repo.FindCarrera(ctx, id)
			if err != nil {
				errorCatalogo(w, r, err)
				return
			}
			WriteJSON(w, r, http.StatusOK, carreraDesdeModelo(c))
		})
	}
}

func facultadDesdeModelo(f catalogos.Facultad) facultadRespuesta {
	return facultadRespuesta{ID: f.ID.String(), Nombre: f.Nombre, Activo: f.Activo}
}

func carreraDesdeModelo(c catalogos.Carrera) carreraRespuesta {
	return carreraRespuesta{
		ID: c.ID.String(), FacultadID: c.FacultyID.String(), Nombre: c.Nombre,
		Activo: c.Activo, FacultadActiva: c.FacultadActiva,
	}
}

func filtrosAcademicos(rawQuery string, filtrarFacultad bool) (Pagination, catalogos.FiltrosAcademicos, pgtype.UUID, error) {
	var facultadID pgtype.UUID
	var filtros catalogos.FiltrosAcademicos
	query, err := url.ParseQuery(rawQuery)
	if err != nil {
		return Pagination{}, filtros, facultadID, &InputError{Code: "bad_request"}
	}
	p, err := ParsePagination(query, map[string]string{"nombre": "nombre_normalizado"}, "nombre")
	if err != nil {
		return Pagination{}, filtros, facultadID, err
	}
	// El repositorio ordena por nombre normalizado e ID en ambos listados.
	if p.Descending {
		return Pagination{}, filtros, facultadID, &InputError{Code: "invalid_pagination"}
	}
	filtros.Limite, filtros.Desplazamiento = int32(p.PageSize), p.Offset
	for key, values := range query {
		switch key {
		case "page", "page_size", "sort", "order":
			continue
		}
		if len(values) != 1 {
			return Pagination{}, filtros, facultadID, &InputError{Code: "bad_request"}
		}
		value := values[0]
		switch key {
		case "q":
			if !utf8.ValidString(value) || strings.ContainsRune(value, 0) {
				return Pagination{}, filtros, facultadID, &InputError{Code: "bad_request"}
			}
			filtros.Busqueda = value
		case "incluir_inactivas":
			if value != "true" && value != "false" {
				return Pagination{}, filtros, facultadID, &InputError{Code: "bad_request"}
			}
			filtros.IncluirInactivas = value == "true"
		case "facultad_id":
			if !filtrarFacultad {
				return Pagination{}, filtros, facultadID, &InputError{Code: "bad_request"}
			}
			facultadID, err = catalogoUUID(value)
			if err != nil {
				return Pagination{}, filtros, facultadID, err
			}
		default:
			return Pagination{}, filtros, facultadID, &InputError{Code: "bad_request"}
		}
	}
	return p, filtros, facultadID, nil
}

func catalogoUUID(value string) (pgtype.UUID, error) {
	var id pgtype.UUID
	if len(value) != 36 || id.Scan(value) != nil || id.Bytes == [16]byte{} || !strings.EqualFold(id.String(), value) {
		return pgtype.UUID{}, &InputError{Code: "bad_request"}
	}
	return id, nil
}

func errorCatalogo(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, catalogos.ErrNotFound):
		WriteError(w, r, "not_found")
	case errors.Is(err, catalogos.ErrInvalidID):
		WriteError(w, r, "bad_request")
	case errors.Is(err, catalogos.ErrInvalidPage):
		WriteError(w, r, "invalid_pagination")
	default:
		WriteError(w, r, "service_unavailable")
	}
}
