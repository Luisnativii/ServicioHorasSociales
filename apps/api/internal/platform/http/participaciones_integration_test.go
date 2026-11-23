package httpserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"horas-sociales/api/internal/modules/participaciones"
	"horas-sociales/api/internal/platform/auth"
)

func TestParticipacionesAPIPostgreSQL(t *testing.T) {
	router, session := catalogoIntegracion(t, func(pool *pgxpool.Pool) func(chi.Router) {
		return ParticipacionesRoutes(participaciones.New(pool), true)
	})
	const open = "/api/v1/proyectos/00000000-0000-4000-8000-000000059001/participaciones"
	const closed = "/api/v1/proyectos/00000000-0000-4000-8000-000000059002/participaciones"
	const first = "00000000-0000-4000-8000-000000059011"
	const second = "00000000-0000-4000-8000-000000059012"
	const inactive = "00000000-0000-4000-8000-000000059013"
	const origin = "https://portal.test"
	var csrf *http.Cookie
	request := func(method, path, body string, authenticated, proof bool) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, origin+path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Origin", origin)
		if authenticated {
			req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: session})
		}
		if proof {
			req.AddCookie(csrf)
			req.Header.Set(auth.CSRFHeader, csrf.Value)
		}
		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)
		return res
	}
	token := request("GET", "/api/v1/auth/csrf", "", true, false)
	for _, cookie := range token.Result().Cookies() {
		if cookie.Name == auth.CSRFCookie {
			csrf = cookie
		}
	}
	if csrf == nil {
		t.Fatal("falta prueba CSRF")
	}
	assert := func(res *httptest.ResponseRecorder, status int, field string) []byte {
		t.Helper()
		if res.Code != status {
			t.Fatalf("esperado %d, recibido %d %s", status, res.Code, res.Body.Bytes())
		}
		if field != "" {
			var problem struct{ Error errorDetail }
			if err := json.Unmarshal(res.Body.Bytes(), &problem); err != nil || problem.Error.Field != field {
				t.Fatalf("esperaba %q: %s", field, res.Body.Bytes())
			}
		}
		return res.Body.Bytes()
	}
	type page struct {
		Participaciones []participaciones.Item
		Pagina          paginaCatalogo
	}
	read := func(path string) page {
		t.Helper()
		var result page
		if err := json.Unmarshal(assert(request("GET", path, "", true, false), 200, ""), &result); err != nil {
			t.Fatal(err)
		}
		return result
	}
	body := func(id string) string { return `{"estudiante_id":"` + id + `"}` }
	assert(request("GET", open, "", false, false), 401, "")
	assert(request("POST", open, body(first), true, false), 403, "")
	assert(request("POST", closed, body(second), true, true), 409, "estado_proyecto")
	assert(request("POST", open, body(inactive), true, true), 409, "estado_estudiante")
	history := read(closed)
	if len(history.Participaciones) != 1 || history.Participaciones[0].Cerrada == nil || history.Participaciones[0].Estado != "FINALIZADA" {
		t.Fatalf("historial inesperado: %+v", history)
	}
	type attempt struct {
		id  string
		res *httptest.ResponseRecorder
	}
	outcomes := make(chan attempt, 2)
	for _, id := range []string{first, second} {
		go func() { outcomes <- attempt{id, request("POST", open, body(id), true, true)} }()
	}
	var winner, waiting string
	for range 2 {
		result := <-outcomes
		if result.res.Code == 201 {
			winner = result.id
			continue
		}
		assert(result.res, 409, "cupo")
		waiting = result.id
	}
	if winner == "" || waiting == "" {
		t.Fatalf("resultado concurrente inesperado: %q %q", winner, waiting)
	}
	one := read(open)
	if len(one.Participaciones) != 1 || one.Participaciones[0].StudentID != winner {
		t.Fatalf("se excedió el cupo sin confirmar: %+v", one)
	}
	assert(request("POST", open, `{"estudiante_id":"`+waiting+`","confirmar_exceso_cupo":true}`, true, true), 201, "")
	assert(request("POST", open, body(winner), true, true), 409, "participacion")
	two := read(open + "?page_size=1")
	if two.Pagina.Total != 2 || len(two.Participaciones) != 1 {
		t.Fatalf("paginación inesperada: %+v", two)
	}
	student := read("/api/v1/estudiantes/" + two.Participaciones[0].StudentID + "/participaciones")
	if len(student.Participaciones) == 0 || student.Participaciones[0].ID != two.Participaciones[0].ID {
		t.Fatalf("las dos vistas difieren: %+v %+v", two, student)
	}
}
