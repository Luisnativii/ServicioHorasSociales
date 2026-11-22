package httpserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"horas-sociales/api/internal/modules/catalogos"
	"horas-sociales/api/internal/platform/auth"
)

func TestCoordinadorCarreraPostgreSQL(t *testing.T) {
	router, session := catalogoIntegracion(t, func(pool *pgxpool.Pool) func(chi.Router) {
		return CoordinadorCarreraRoutes(catalogos.NewCoordinadorRepository(pool))
	})
	const host = "https://portal.test"
	const carreraA = "00000000-0000-4000-8000-000000035002"
	const carreraB = "00000000-0000-4000-8000-000000035004"
	const carreraInactiva = "00000000-0000-4000-8000-000000035005"
	const base = "/api/v1/catalogos/carreras/"
	var csrf *http.Cookie
	request := func(method, path, body, jwt string, proof bool, expected int) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(method, host+path, strings.NewReader(body))
		req.Header.Set("Origin", host)
		req.Header.Set("Content-Type", "application/json")
		if jwt != "" {
			req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: jwt})
		}
		if proof {
			req.AddCookie(csrf)
			req.Header.Set(auth.CSRFHeader, csrf.Value)
		}
		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)
		if res.Code != expected || res.Header().Get("Cache-Control") != "no-store" {
			t.Fatalf("%s %s: se esperaba %d, se recibió %d %s", method, path, expected, res.Code, res.Body.String())
		}
		return res
	}
	prepared := request("GET", "/api/v1/auth/csrf", "", session, false, 204)
	for _, cookie := range prepared.Result().Cookies() {
		if cookie.Name == auth.CSRFCookie {
			csrf = cookie
		}
	}
	if csrf == nil {
		t.Fatal("falta la cookie CSRF")
	}
	path := func(id string) string { return base + id + "/coordinador" }
	assertName := func(res *httptest.ResponseRecorder, id string, expected *string) {
		t.Helper()
		var found struct {
			ID     string  `json:"carrera_id"`
			Nombre *string `json:"coordinador_nombre"`
		}
		if err := json.Unmarshal(res.Body.Bytes(), &found); err != nil || found.ID != id {
			t.Fatalf("respuesta inesperada: %s", res.Body.String())
		}
		if (found.Nombre == nil) != (expected == nil) || (expected != nil && *found.Nombre != *expected) {
			t.Fatalf("nombre inesperado: %s", res.Body.String())
		}
	}
	const name = `{"coordinador_nombre":"  Nombre de prueba  "}`
	const clear = `{"coordinador_nombre":null}`
	request("GET", path(carreraA), "", "", false, 401)
	request("PUT", path(carreraA), name, session, false, 403)
	request("PUT", path("malformado"), name, session, true, 400)
	request("PUT", path("00000000-0000-4000-8000-000000035099"), name, session, true, 404)
	request("GET", path("00000000-0000-4000-8000-000000035099"), "", session, false, 404)
	for _, invalid := range []string{`{}`, `{"coordinador_nombre":false}`, `{"coordinador_nombre":"   "}`} {
		request("PUT", path(carreraA), invalid, session, true, 400)
	}
	assertName(request("GET", path(carreraInactiva), "", session, false, 200), carreraInactiva, nil)
	want := "Nombre de prueba"
	for _, id := range []string{carreraA, carreraB} {
		assertName(request("PUT", path(id), clear, session, true, 200), id, nil)
		assertName(request("PUT", path(id), name, session, true, 200), id, &want)
		assertName(request("GET", path(id), "", session, false, 200), id, &want)
	}
	assertName(request("PUT", path(carreraA), clear, session, true, 200), carreraA, nil)
	assertName(request("GET", path(carreraB), "", session, false, 200), carreraB, &want)
	assertName(request("PUT", path(carreraB), clear, session, true, 200), carreraB, nil)
}
