package httpserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"horas-sociales/api/internal/modules/estudiantes"
	"horas-sociales/api/internal/platform/auth"
)

func TestEstudiantesAPIPostgreSQL(t *testing.T) {
	router, session := catalogoIntegracion(t, func(pool *pgxpool.Pool) func(chi.Router) {
		return EstudiantesRoutes(estudiantes.NewRepository(pool))
	})
	const origin = "https://portal.test"
	const base = "/api/v1/estudiantes"
	const first = "00000000-0000-4000-8000-000000035002"
	const second = "00000000-0000-4000-8000-000000055002"
	var csrf *http.Cookie
	request := func(method, path, body string, authenticated, proof bool, status int, target any) {
		t.Helper()
		req := httptest.NewRequest(method, origin+path, strings.NewReader(body))
		req.Header.Set("Origin", origin)
		req.Header.Set("Content-Type", "application/json")
		if authenticated {
			req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: session})
		}
		if proof {
			req.AddCookie(csrf)
			req.Header.Set(auth.CSRFHeader, csrf.Value)
		}
		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)
		if res.Code != status || res.Header().Get("Cache-Control") != "no-store" {
			t.Fatalf("%s %s: esperado %d, recibido %d %s", method, path, status, res.Code, res.Body.String())
		}
		for _, cookie := range res.Result().Cookies() {
			if cookie.Name == auth.CSRFCookie {
				csrf = cookie
			}
		}
		if strings.Contains(res.Body.String(), "dui") {
			t.Fatal("dato excluido del DTO")
		}
		if target != nil {
			if err := json.Unmarshal(res.Body.Bytes(), target); err != nil {
				t.Fatal(err)
			}
		}
	}
	request("GET", base, "", false, false, 401, nil)
	request("POST", base, "{}", true, false, 403, nil)
	request("GET", "/api/v1/auth/csrf", "", true, false, 204, nil)
	if csrf == nil {
		t.Fatal("falta cookie CSRF")
	}
	var page struct {
		Estudiantes []estudiantes.Student
		Pagina      paginaCatalogo
	}
	request("GET", base+"?q=00550001&carrera_id="+first, "", true, false, 200, &page)
	if len(page.Estudiantes) != 1 || page.Estudiantes[0].NombreCompleto != "ARCHIVO, ESTUDIANTE" || page.Estudiantes[0].Nombres != nil {
		t.Fatalf("histórico inesperado: %+v", page)
	}
	historical := page.Estudiantes[0]
	itemPath := base + "/" + historical.ID
	body := `{"carnet":"00550001","correo_institucional":"archivo55@uca.edu.sv","carrera_id":"` + second + `"`
	request("PUT", itemPath, body+"}", true, true, 400, nil)
	request("PUT", itemPath, body+`,"motivo_cambio_carrera":"Traslado confirmado"}`, true, true, 200, nil)
	var detail estudiantes.Detail
	request("GET", itemPath, "", true, false, 200, &detail)
	if detail.Carrera.ID != second || detail.NombreCompleto != historical.NombreCompleto || !detail.HistorialDisponible ||
		len(detail.Vigencias) != 2 || detail.Vigencias[0].Motivo == nil || *detail.Vigencias[0].Motivo != "Traslado confirmado" ||
		detail.Vigencias[1].Carrera.ID != first || detail.Vigencias[1].Desde != nil ||
		detail.Vigencias[1].Hasta == nil || detail.Vigencias[0].Desde == nil ||
		!detail.Vigencias[1].Hasta.Equal(*detail.Vigencias[0].Desde) {
		t.Fatalf("vigencias inesperadas: %+v", detail)
	}
	input := `{"carnet":"00550002","nombres":"Estudiante","apellidos":"Sintético","correo_institucional":"estudiante55@uca.edu.sv","carrera_id":"` + first + `"}`
	var item estudiantes.Student
	request("POST", base, input, true, true, 201, &item)
	if item.Carnet != "00550002" || item.Facultad.ID != historical.Facultad.ID {
		t.Fatalf("alta inesperada: %+v", item)
	}
	var duplicate struct{ Error errorDetail }
	request("POST", base, strings.Replace(input, "estudiante55@", "otro55@", 1), true, true, 409, &duplicate)
	if duplicate.Error.Field != "carnet" {
		t.Fatalf("campo del conflicto inesperado: %+v", duplicate)
	}
	request("GET", base+"?q=005500&facultad_id="+item.Facultad.ID+"&page_size=1&page=2", "", true, false, 200, &page)
	if page.Pagina.Total != 2 || len(page.Estudiantes) != 1 || page.Estudiantes[0].ID != item.ID {
		t.Fatalf("página inesperada: %+v", page)
	}
	request("GET", base+"/00000000-0000-4000-8000-000000055404", "", true, false, 404, nil)
}
