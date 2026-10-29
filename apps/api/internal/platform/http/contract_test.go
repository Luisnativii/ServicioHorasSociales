package httpserver

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestJSONHandlerContract(t *testing.T) {
	for _, tc := range []struct {
		name, body, contentType string
		status                  int
	}{
		{"valid", `{"name":"prueba"}`, "application/json; charset=utf-8", 200},
		{"empty", "", "application/json", 400},
		{"malformed", `{"name":`, "application/json", 400},
		{"wrong type", `{"name":9}`, "application/json", 400},
		{"unknown field", `{"secret":"SELECT password FROM admin"}`, "application/json", 400},
		{"two objects", `{} {}`, "application/json", 400},
		{"trailing data", `{} invalid`, "application/json", 400},
		{"null", `null`, "application/json", 400},
		{"array", `[]`, "application/json", 400},
		{"media type", `{}`, "text/plain", 415},
		{"missing media type", `{}`, "", 415},
		{"large object", `{"name":"` + strings.Repeat("x", int(MaxJSONBytes)) + `"}`, "application/json", 413},
		{"large trailing whitespace", `{}` + strings.Repeat(" ", int(MaxJSONBytes)), "application/json", 413},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest("POST", "/", strings.NewReader(tc.body))
			r.Header.Set("Content-Type", tc.contentType)
			w := httptest.NewRecorder()
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var input struct {
					Name string `json:"name"`
				}
				if err := ReadJSON(w, r, &input); err != nil {
					WriteInputError(w, r, err)
					return
				}
				WriteJSON(w, r, 200, input)
			})
			handler.ServeHTTP(w, r)
			if w.Code != tc.status || !json.Valid(w.Body.Bytes()) {
				t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
			}
			if strings.Contains(w.Body.String(), "SELECT") || strings.Contains(w.Body.String(), "secret") {
				t.Fatal("se filtró contenido del error de entrada")
			}
		})
	}
}

func TestPaginationHandler(t *testing.T) {
	for _, query := range []string{
		"page=0", "page=-1", "page=x", "page=9223372036854775807&page_size=100",
		"page_size=0", "page_size=101", "page_size=1.5", "page_size=", "page=1&page=2",
		"sort=name%3BDROP+TABLE+students", "order=desc%3BDROP", "sort=", "order=DESC",
	} {
		t.Run(query, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/?"+query, nil)
			w := httptest.NewRecorder()
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, err := ParsePagination(r.URL.Query(), map[string]string{"name": "student_name"}, "name")
				if err == nil {
					t.Fatal("se aceptó paginación/orden inválido")
				}
				WriteInputError(w, r, err)
			}).ServeHTTP(w, r)
			if w.Code != 400 || strings.Contains(w.Body.String(), "DROP") {
				t.Fatalf("respuesta: %s", w.Body.String())
			}
		})
	}
	for _, tc := range []struct {
		query  string
		page   int64
		size   int
		offset int64
		desc   bool
	}{
		{"", 1, 25, 0, false},
		{"page=3&page_size=100&sort=name&order=desc", 3, 100, 200, true},
	} {
		q, _ := url.ParseQuery(tc.query)
		p, err := ParsePagination(q, map[string]string{"name": "student_name"}, "name")
		if err != nil || p.Page != tc.page || p.PageSize != tc.size || p.Offset != tc.offset || p.Descending != tc.desc || p.OrderColumn != "student_name" {
			t.Fatalf("pagination=%+v err=%v", p, err)
		}
	}
}

func TestErrorsAreGenericAndCorrelated(t *testing.T) {
	w := httptest.NewRecorder()
	Router(&probe{}).ServeHTTP(w, httptest.NewRequest("GET", "/api/v1/missing", nil))
	var body struct {
		Error errorDetail `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if w.Code != 401 || body.Error.Code != "unauthorized" || body.Error.RequestID == "" || body.Error.RequestID != w.Header().Get("X-Request-ID") {
		t.Fatalf("respuesta: %s", w.Body.String())
	}
	for _, mode := range []string{"unknown code", "internal error", "encoding error"} {
		w = httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		switch mode {
		case "unknown code":
			WriteError(w, r, "postgres://secret:password@db")
		case "internal error":
			WriteInputError(w, r, errors.New("SQL password secret"))
		case "encoding error":
			WriteJSON(w, r, 200, make(chan int))
		}
		if w.Code != 500 || !json.Valid(w.Body.Bytes()) || strings.Contains(w.Body.String(), "secret") || strings.Contains(w.Body.String(), "SQL") {
			t.Fatalf("respuesta: %s", w.Body.String())
		}
	}
}
