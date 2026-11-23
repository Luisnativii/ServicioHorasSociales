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
	"unicode/utf8"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"horas-sociales/api/internal/modules/estudiantes"
	"horas-sociales/api/internal/platform/auth"
)

const studentPath = "/estudiantes"

var studentCarnet = regexp.MustCompile(`^[0-9]{8}$`)

func EstudiantesRoutes(repo *estudiantes.Repository) func(chi.Router) {
	return func(r chi.Router) {
		r.Get(studentPath, listStudents(repo))
		r.Get(studentPath+"/{id}", func(w http.ResponseWriter, r *http.Request) {
			id, err := catalogoUUID(chi.URLParam(r, "id"))
			if err != nil {
				WriteInputError(w, r, err)
				return
			}
			ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
			defer cancel()
			item, err := repo.Get(ctx, id)
			writeStudent(w, r, http.StatusOK, item, err)
		})
		r.Post(studentPath, saveStudent(repo, true))
		r.Put(studentPath+"/{id}", saveStudent(repo, false))
	}
}

func listStudents(repo *estudiantes.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			WriteError(w, r, "bad_request")
			return
		}
		var faculty, career pgtype.UUID
		var search string
		for key, values := range query {
			if len(values) != 1 {
				writeFieldError(w, r, "bad_request", key)
				return
			}
			switch key {
			case "page", "page_size":
				continue
			case "q":
				search = strings.TrimSpace(values[0])
				if utf8.RuneCountInString(search) > 100 || strings.ContainsRune(search, 0) {
					err = errors.New("búsqueda inválida")
				}
			case "facultad_id":
				faculty, err = catalogoUUID(values[0])
			case "carrera_id":
				career, err = catalogoUUID(values[0])
			default:
				WriteError(w, r, "bad_request")
				return
			}
			if err != nil {
				writeFieldError(w, r, "bad_request", key)
				return
			}
		}
		p, err := ParsePagination(query, map[string]string{"nombre": "nombre"}, "nombre")
		if err != nil {
			WriteInputError(w, r, err)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		items, total, err := repo.List(ctx, search, faculty, career, p.PageSize, p.Offset)
		writeStudent(w, r, http.StatusOK, struct {
			Estudiantes []estudiantes.Student `json:"estudiantes"`
			Pagina      paginaCatalogo        `json:"pagina"`
		}{items, paginaCatalogo{p.Page, p.PageSize, total}}, err)
	}
}

func saveStudent(repo *estudiantes.Repository, create bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		admin, ok := auth.FromContext(r.Context())
		if !ok {
			WriteError(w, r, "unauthorized")
			return
		}
		actor, err := catalogoUUID(admin.ID)
		if err != nil {
			WriteError(w, r, "unauthorized")
			return
		}
		var id pgtype.UUID
		status := http.StatusCreated
		if !create {
			status = http.StatusOK
			id, err = catalogoUUID(chi.URLParam(r, "id"))
			if err != nil {
				WriteInputError(w, r, err)
				return
			}
		}
		var body struct {
			Carnet    string  `json:"carnet"`
			Nombres   *string `json:"nombres"`
			Apellidos *string `json:"apellidos"`
			Correo    string  `json:"correo_institucional"`
			Carrera   string  `json:"carrera_id"`
			Motivo    string  `json:"motivo_cambio_carrera"`
		}
		if err := ReadJSON(w, r, &body); err != nil {
			WriteInputError(w, r, err)
			return
		}
		body.Carnet = strings.TrimSpace(body.Carnet)
		body.Correo = strings.ToLower(strings.TrimSpace(body.Correo))
		body.Motivo = strings.TrimSpace(body.Motivo)
		if !studentCarnet.MatchString(body.Carnet) {
			writeFieldError(w, r, "bad_request", "carnet")
			return
		}
		parsed, err := mail.ParseAddress(body.Correo)
		if err != nil || parsed.Address != body.Correo || !strings.HasSuffix(body.Correo, "@uca.edu.sv") {
			writeFieldError(w, r, "bad_request", "correo_institucional")
			return
		}
		fields := []struct {
			name  string
			value *string
		}{
			{"nombres", body.Nombres}, {"apellidos", body.Apellidos},
		}
		for _, field := range fields {
			if field.value == nil {
				if create || body.Nombres != nil || body.Apellidos != nil {
					writeFieldError(w, r, "bad_request", field.name)
					return
				}
				continue
			}
			*field.value = strings.TrimSpace(*field.value)
			if *field.value == "" || utf8.RuneCountInString(*field.value) > 120 || strings.ContainsRune(*field.value, 0) {
				writeFieldError(w, r, "bad_request", field.name)
				return
			}
		}
		career, err := catalogoUUID(body.Carrera)
		if err != nil {
			writeFieldError(w, r, "bad_request", "carrera_id")
			return
		}
		if utf8.RuneCountInString(body.Motivo) > 500 || strings.ContainsRune(body.Motivo, 0) {
			writeFieldError(w, r, "bad_request", "motivo_cambio_carrera")
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		item, err := repo.Save(ctx, id, actor, estudiantes.Input{
			Carnet: body.Carnet, Correo: body.Correo, Nombres: body.Nombres,
			Apellidos: body.Apellidos, Carrera: career, Motivo: body.Motivo,
		})
		writeStudent(w, r, status, item, err)
	}
}

func writeStudent(w http.ResponseWriter, r *http.Request, status int, item any, err error) {
	switch {
	case err == nil:
		WriteJSON(w, r, status, item)
	case errors.Is(err, pgx.ErrNoRows):
		WriteError(w, r, "not_found")
	case errors.Is(err, estudiantes.ErrCareer):
		writeFieldError(w, r, "bad_request", "carrera_id")
	case errors.Is(err, estudiantes.ErrReason):
		writeFieldError(w, r, "bad_request", "motivo_cambio_carrera")
	case errors.Is(err, estudiantes.ErrHistory):
		writeFieldError(w, r, "service_unavailable", "historial_academico")
	default:
		if field := estudiantes.DuplicateField(err); field != "" {
			writeFieldError(w, r, "conflict", field)
			return
		}
		WriteError(w, r, "service_unavailable")
	}
}
