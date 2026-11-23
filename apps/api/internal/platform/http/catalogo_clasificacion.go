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

type opcionClasificacion interface {
	catalogos.TipoProyecto | catalogos.Area | catalogos.Modalidad
}

type operacionesClasificacion[T opcionClasificacion] struct {
	pagina func(context.Context, catalogos.FiltrosClasificacion) ([]T, int64, error)
	buscar func(context.Context, pgtype.UUID) (T, error)
	crear  func(context.Context, string) (T, error)
	editar func(context.Context, pgtype.UUID, string) (T, error)
	estado func(context.Context, pgtype.UUID, bool) (T, error)
}

type clasificacionRespuesta struct {
	ID     string `json:"id"`
	Nombre string `json:"nombre"`
	Activo bool   `json:"activo"`
}

func CatalogoClasificacionRoutes(repo *catalogos.Repository) func(chi.Router) {
	return func(r chi.Router) {
		rutasClasificacion(r, "/catalogos/tipos-proyecto", operacionesClasificacion[catalogos.TipoProyecto]{
			pagina: repo.PageTipos, buscar: repo.FindTipoByID, crear: repo.CreateTipo,
			editar: repo.RenameTipo, estado: repo.SetTipoActive,
		})
		rutasClasificacion(r, "/catalogos/modalidades", operacionesClasificacion[catalogos.Modalidad]{
			pagina: repo.PageModalidades, buscar: repo.FindModalidadByID, crear: repo.CreateModalidad,
			editar: repo.RenameModalidad, estado: repo.SetModalidadActive,
		})
		rutasClasificacion(r, "/catalogos/areas", operacionesClasificacion[catalogos.Area]{
			pagina: repo.PageAreas, buscar: repo.FindAreaByID, crear: repo.CreateArea,
			editar: repo.RenameArea, estado: repo.SetAreaActive,
		})
	}
}

func rutasClasificacion[T opcionClasificacion](r chi.Router, ruta string, ops operacionesClasificacion[T]) {
	r.Get(ruta, func(w http.ResponseWriter, r *http.Request) {
		p, filtros, err := filtrosClasificacion(r.URL.RawQuery)
		if err != nil {
			WriteInputError(w, r, err)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		items, total, err := ops.pagina(ctx, filtros)
		if err != nil {
			errorClasificacion(w, r, err)
			return
		}
		opciones := make([]clasificacionRespuesta, 0, len(items))
		for _, item := range items {
			opciones = append(opciones, opcionDesdeModelo(item))
		}
		WriteJSON(w, r, http.StatusOK, struct {
			Opciones []clasificacionRespuesta `json:"opciones"`
			Pagina   paginaCatalogo           `json:"pagina"`
		}{opciones, paginaCatalogo{p.Page, p.PageSize, total}})
	})
	r.Post(ruta, operacionClasificacion(http.StatusCreated, false, func(w http.ResponseWriter, r *http.Request, ctx context.Context, _ pgtype.UUID) (T, error) {
		nombre, err := nombreClasificacion(w, r)
		if err != nil {
			var cero T
			return cero, err
		}
		return ops.crear(ctx, nombre)
	}))
	r.Get(ruta+"/{id}", operacionClasificacion(http.StatusOK, true, func(_ http.ResponseWriter, _ *http.Request, ctx context.Context, id pgtype.UUID) (T, error) {
		return ops.buscar(ctx, id)
	}))
	r.Patch(ruta+"/{id}", operacionClasificacion(http.StatusOK, true, func(w http.ResponseWriter, r *http.Request, ctx context.Context, id pgtype.UUID) (T, error) {
		nombre, err := nombreClasificacion(w, r)
		if err != nil {
			var cero T
			return cero, err
		}
		return ops.editar(ctx, id, nombre)
	}))
	r.Post(ruta+"/{id}/desactivar", operacionClasificacion(http.StatusOK, true, func(_ http.ResponseWriter, _ *http.Request, ctx context.Context, id pgtype.UUID) (T, error) {
		return ops.estado(ctx, id, false)
	}))
}

func operacionClasificacion[T opcionClasificacion](status int, requiereID bool, operar func(http.ResponseWriter, *http.Request, context.Context, pgtype.UUID) (T, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var id pgtype.UUID
		if requiereID {
			var err error
			id, err = catalogoUUID(chi.URLParam(r, "id"))
			if err != nil {
				WriteInputError(w, r, err)
				return
			}
		}
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		item, err := operar(w, r, ctx, id)
		if err != nil {
			errorClasificacion(w, r, err)
			return
		}
		WriteJSON(w, r, status, opcionDesdeModelo(item))
	}
}

func opcionDesdeModelo[T opcionClasificacion](item T) clasificacionRespuesta {
	opcion := catalogos.TipoProyecto(item)
	return clasificacionRespuesta{ID: opcion.ID.String(), Nombre: opcion.Nombre, Activo: opcion.Activo}
}

const maxNombreClasificacion = 200

func nombreClasificacion(w http.ResponseWriter, r *http.Request) (string, error) {
	var input struct {
		Nombre string `json:"nombre"`
	}
	if err := ReadJSON(w, r, &input); err != nil {
		return "", err
	}
	nombre, err := catalogos.VisibleName(input.Nombre)
	if err != nil || !textoClasificacionValido(nombre) {
		return "", catalogos.ErrInvalidName
	}
	return nombre, nil
}

func textoClasificacionValido(value string) bool {
	return utf8.ValidString(value) && !strings.ContainsRune(value, 0) && utf8.RuneCountInString(value) <= maxNombreClasificacion
}

func filtrosClasificacion(rawQuery string) (Pagination, catalogos.FiltrosClasificacion, error) {
	var filtros catalogos.FiltrosClasificacion
	query, err := url.ParseQuery(rawQuery)
	if err != nil {
		return Pagination{}, filtros, &InputError{Code: "bad_request"}
	}
	for key, values := range query {
		switch key {
		case "page", "page_size":
			continue
		case "q":
			if len(values) == 1 && textoClasificacionValido(values[0]) {
				filtros.Busqueda = strings.TrimSpace(values[0])
				continue
			}
		case "incluir_inactivos":
			if len(values) == 1 && (values[0] == "true" || values[0] == "false") {
				filtros.IncluirInactivos = values[0] == "true"
				continue
			}
		}
		return Pagination{}, filtros, &InputError{Code: "bad_request"}
	}
	p, err := ParsePagination(query, map[string]string{"nombre": "nombre"}, "nombre")
	filtros.Limite, filtros.Desplazamiento = int32(p.PageSize), p.Offset
	return p, filtros, err
}

func errorClasificacion(w http.ResponseWriter, r *http.Request, err error) {
	var input *InputError
	switch {
	case errors.As(err, &input):
		WriteInputError(w, r, input)
	case errors.Is(err, catalogos.ErrInvalidName):
		WriteError(w, r, "bad_request")
	case errors.Is(err, catalogos.ErrDuplicate):
		WriteError(w, r, "conflict")
	default:
		errorCatalogo(w, r, err)
	}
}
