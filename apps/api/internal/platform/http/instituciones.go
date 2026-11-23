package httpserver

import (
	"context"
	"errors"
	"net/http"
	"net/mail"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"horas-sociales/api/internal/modules/instituciones"
	"horas-sociales/api/internal/platform/auth"
)

type sedeInstitucion struct {
	DistritoID string `json:"distrito_id"`
	Nombre     string `json:"nombre"`
}

type contactoInstitucion struct {
	Nombre    *string `json:"nombre"`
	Correo    *string `json:"correo"`
	Telefono  *string `json:"telefono"`
	Extension *string `json:"extension"`
}

var correoInstitucional = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)
var telefonoInstitucional = regexp.MustCompile(`^[0-9+\-[:space:]]{8,20}$`)

func (c contactoInstitucion) campoInvalido() string {
	if c.Nombre == nil || strings.TrimSpace(*c.Nombre) == "" || !textoClasificacionValido(*c.Nombre) {
		return "contacto.nombre"
	}
	if c.Correo != nil {
		parsed, err := mail.ParseAddress(*c.Correo)
		if err != nil || parsed.Address != *c.Correo || !correoInstitucional.MatchString(*c.Correo) {
			return "contacto.correo"
		}
	}
	if c.Telefono != nil && (!telefonoInstitucional.MatchString(*c.Telefono) || !strings.ContainsAny(*c.Telefono, "0123456789")) {
		return "contacto.telefono"
	}
	if c.Extension != nil && (c.Telefono == nil || strings.TrimSpace(*c.Extension) == "" || !textoClasificacionValido(*c.Extension)) {
		return "contacto.extension"
	}
	return ""
}

type institucionRespuesta struct {
	Contacto          *contactoInstitucion `json:"contacto"`
	ID                string               `json:"id"`
	Nombre            string               `json:"nombre"`
	CodigoCSS         *string              `json:"codigo_css"`
	Activo            bool                 `json:"activo"`
	Sede              *sedeInstitucion     `json:"sede"`
	UbicacionHeredada *string              `json:"ubicacion_heredada"`
}

func InstitucionesRoutes(repo *instituciones.Repository) func(chi.Router) {
	return func(r chi.Router) {
		const base = "/instituciones"
		r.Get(base, listarInstituciones(repo))
		r.Get(base+"/{id}", func(w http.ResponseWriter, r *http.Request) {
			id, err := catalogoUUID(chi.URLParam(r, "id"))
			if err != nil {
				WriteInputError(w, r, err)
				return
			}
			ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
			defer cancel()
			item, err := repo.FindByID(ctx, id)
			escribirInstitucion(w, r, http.StatusOK, item, err)
		})
		r.Post(base, cambiarInstitucion(repo, http.StatusCreated, false, false))
		r.Put(base+"/{id}", cambiarInstitucion(repo, http.StatusOK, true, false))
		r.Post(base+"/{id}/desactivar", cambiarInstitucion(repo, http.StatusOK, true, true))
	}
}

func listarInstituciones(repo *instituciones.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			WriteError(w, r, "bad_request")
			return
		}
		var filtros instituciones.Filtros
		for key, values := range query {
			switch key {
			case "page", "page_size":
				continue
			case "q":
				if len(values) == 1 && textoClasificacionValido(values[0]) {
					filtros.Busqueda = strings.TrimSpace(values[0])
					continue
				}
			case "incluir_inactivas":
				if len(values) == 1 && (values[0] == "true" || values[0] == "false") {
					filtros.IncluirInactivas = values[0] == "true"
					continue
				}
			}
			WriteError(w, r, "bad_request")
			return
		}
		p, err := ParsePagination(query, map[string]string{"nombre": "nombre"}, "nombre")
		if err != nil {
			WriteInputError(w, r, err)
			return
		}
		filtros.Limite, filtros.Desplazamiento = int32(p.PageSize), p.Offset
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		items, total, err := repo.List(ctx, filtros)
		if err != nil {
			WriteError(w, r, "service_unavailable")
			return
		}
		resultado := make([]institucionRespuesta, 0, len(items))
		for _, item := range items {
			resultado = append(resultado, institucionDesdeModelo(item))
		}
		WriteJSON(w, r, http.StatusOK, struct {
			Instituciones []institucionRespuesta `json:"instituciones"`
			Pagina        paginaCatalogo         `json:"pagina"`
		}{resultado, paginaCatalogo{p.Page, p.PageSize, total}})
	}
}

func cambiarInstitucion(repo *instituciones.Repository, status int, requiereID, desactivar bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
		var id pgtype.UUID
		if requiereID {
			id, err = catalogoUUID(chi.URLParam(r, "id"))
			if err != nil {
				WriteInputError(w, r, err)
				return
			}
		}
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		if desactivar {
			item, err := repo.Deactivate(ctx, id, actor)
			escribirInstitucion(w, r, status, item, err)
			return
		}
		var body struct {
			Nombre     string               `json:"nombre"`
			DistritoID *string              `json:"distrito_id"`
			Contacto   *contactoInstitucion `json:"contacto"`
		}
		if err := ReadJSON(w, r, &body); err != nil {
			WriteInputError(w, r, err)
			return
		}
		datos := instituciones.DatosInstitucion{Nombre: strings.TrimSpace(body.Nombre)}
		if datos.Nombre == "" || !textoClasificacionValido(datos.Nombre) {
			writeFieldError(w, r, "bad_request", "nombre")
			return
		}
		if body.DistritoID != nil {
			datos.DistritoID, err = catalogoUUID(*body.DistritoID)
			if err != nil {
				writeFieldError(w, r, "bad_request", "distrito_id")
				return
			}
		}
		if body.Contacto != nil {
			if field := body.Contacto.campoInvalido(); field != "" {
				writeFieldError(w, r, "bad_request", field)
				return
			}
			datos.Contacto = &instituciones.Contacto{
				Nombre: body.Contacto.Nombre, Correo: body.Contacto.Correo,
				Telefono: body.Contacto.Telefono, Extension: body.Contacto.Extension,
			}
		}
		var item instituciones.Institucion
		if requiereID {
			item, err = repo.Update(ctx, id, actor, datos)
		} else {
			item, err = repo.Create(ctx, actor, datos)
		}
		escribirInstitucion(w, r, status, item, err)
	}
}

func institucionDesdeModelo(item instituciones.Institucion) institucionRespuesta {
	res := institucionRespuesta{
		ID: item.ID.String(), Nombre: item.Nombre, CodigoCSS: item.CodigoCSS,
		Activo: item.Activo, UbicacionHeredada: item.UbicacionHeredada,
	}
	if item.Sede != nil {
		res.Sede = &sedeInstitucion{item.Sede.DistritoID.String(), item.Sede.Nombre}
	}
	contacto := item.Contacto
	if contacto.Nombre != nil || contacto.Correo != nil || contacto.Telefono != nil || contacto.Extension != nil {
		res.Contacto = &contactoInstitucion{contacto.Nombre, contacto.Correo, contacto.Telefono, contacto.Extension}
	}
	return res
}

func escribirInstitucion(w http.ResponseWriter, r *http.Request, status int, item instituciones.Institucion, err error) {
	switch {
	case err == nil:
		WriteJSON(w, r, status, institucionDesdeModelo(item))
	case errors.Is(err, instituciones.ErrDuplicate):
		writeFieldError(w, r, "conflict", "nombre")
	case errors.Is(err, instituciones.ErrInvalidSede):
		writeFieldError(w, r, "bad_request", "distrito_id")
	case errors.Is(err, instituciones.ErrNotFound):
		WriteError(w, r, "not_found")
	default:
		WriteError(w, r, "service_unavailable")
	}
}
