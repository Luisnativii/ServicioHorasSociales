package httpserver

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"math"
	"mime"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-chi/chi/v5/middleware"
)

const MaxJSONBytes int64 = 1 << 20

// InputError contains only a public code, never a decoder/database error.
type InputError struct{ Code string }

func (e *InputError) Error() string { return e.Code }

type errorDetail struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

var publicErrors = map[string]struct {
	status  int
	message string
}{
	"invalid_json":           {400, "El cuerpo debe ser un objeto JSON válido con los campos permitidos."},
	"invalid_pagination":     {400, "Revisa página, tamaño y ordenamiento."},
	"bad_request":            {400, "La solicitud no es válida."},
	"unauthorized":           {401, "Debes iniciar sesión."},
	"forbidden":              {403, "La operación no está permitida."},
	"not_found":              {404, "No se encontró el recurso."},
	"method_not_allowed":     {405, "El método no está permitido."},
	"conflict":               {409, "La operación entra en conflicto con el estado actual."},
	"payload_too_large":      {413, "El cuerpo supera el límite de 1 MiB."},
	"unsupported_media_type": {415, "Usa Content-Type application/json."},
	"internal_error":         {500, "Ocurrió un error interno."},
	"service_unavailable":    {503, "El servicio no está disponible."},
}

// WriteError uses a closed vocabulary so raw SQL/errors never reach clients.
func WriteError(w http.ResponseWriter, r *http.Request, code string) {
	definition, ok := publicErrors[code]
	if !ok {
		code = "internal_error"
		definition = publicErrors[code]
	}
	WriteJSON(w, r, definition.status, struct {
		Error errorDetail `json:"error"`
	}{errorDetail{code, definition.message, middleware.GetReqID(r.Context())}})
}

func WriteInputError(w http.ResponseWriter, r *http.Request, err error) {
	var input *InputError
	if errors.As(err, &input) {
		WriteError(w, r, input.Code)
		return
	}
	WriteError(w, r, "internal_error")
}

// Marshal before sending headers, so encoding failures cannot produce a 200
// followed by a partial JSON body. Error envelopes themselves are marshalable.
func WriteJSON(w http.ResponseWriter, r *http.Request, status int, value any) {
	body, err := json.Marshal(value)
	if err != nil {
		WriteError(w, r, "internal_error")
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_, _ = w.Write(append(body, '\n'))
}

// ReadJSON accepts one object, rejects unknown fields and limits the entire
// body (including trailing whitespace). dst is a pointer to a handler's DTO.
func ReadJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		return &InputError{"unsupported_media_type"}
	}
	r.Body = http.MaxBytesReader(w, r.Body, MaxJSONBytes)
	decoder := json.NewDecoder(r.Body)
	var raw json.RawMessage
	if err := decoder.Decode(&raw); err != nil {
		return jsonInputError(err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return &InputError{"invalid_json"}
		}
		return jsonInputError(err)
	}
	if len(raw) == 0 || raw[0] != '{' {
		return &InputError{"invalid_json"}
	}
	object := json.NewDecoder(bytes.NewReader(raw))
	object.DisallowUnknownFields()
	if err := object.Decode(dst); err != nil {
		return &InputError{"invalid_json"}
	}
	return nil
}

func jsonInputError(err error) error {
	var limit *http.MaxBytesError
	if errors.As(err, &limit) {
		return &InputError{"payload_too_large"}
	}
	return &InputError{"invalid_json"}
}

type Pagination struct {
	Page        int64
	PageSize    int
	Offset      int64
	OrderColumn string
	Descending  bool
}

// allowedSort maps public names to fixed, developer-owned SQL identifiers.
// Never populate it from request values. Values/filters still use SQL params.
func ParsePagination(query url.Values, allowedSort map[string]string, defaultSort string) (Pagination, error) {
	p := Pagination{Page: 1, PageSize: 25}
	values := map[string]string{"page": "1", "page_size": "25", "sort": defaultSort, "order": "asc"}
	for key := range values {
		if items, exists := query[key]; exists {
			if len(items) != 1 || items[0] == "" {
				return Pagination{}, &InputError{"invalid_pagination"}
			}
			values[key] = items[0]
		}
	}
	page, err := strconv.ParseInt(values["page"], 10, 64)
	if err != nil || page < 1 {
		return Pagination{}, &InputError{"invalid_pagination"}
	}
	size, err := strconv.Atoi(values["page_size"])
	if err != nil || size < 1 || size > 100 || page-1 > math.MaxInt64/int64(size) {
		return Pagination{}, &InputError{"invalid_pagination"}
	}
	column, ok := allowedSort[values["sort"]]
	if !ok || column == "" || (values["order"] != "asc" && values["order"] != "desc") {
		return Pagination{}, &InputError{"invalid_pagination"}
	}
	p.Page, p.PageSize, p.Offset = page, size, (page-1)*int64(size)
	p.OrderColumn, p.Descending = column, values["order"] == "desc"
	return p, nil
}
