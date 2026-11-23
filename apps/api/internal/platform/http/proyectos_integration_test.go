package httpserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"horas-sociales/api/internal/modules/proyectos"
	"horas-sociales/api/internal/platform/auth"
)

func TestProyectosAPIPostgreSQL(t *testing.T) {
	router, session := catalogoIntegracion(t, func(pool *pgxpool.Pool) func(chi.Router) {
		return ProyectosRoutes(proyectos.New(pool), true)
	})
	const origin = "https://portal.test"
	const base = "/api/v1/proyectos"
	var csrf *http.Cookie
	call := func(method, path, body string, proof bool, status int) []byte {
		t.Helper()
		req := httptest.NewRequest(method, origin+path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Origin", origin)
		req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: session})
		if proof {
			req.AddCookie(csrf)
			req.Header.Set(auth.CSRFHeader, csrf.Value)
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != status {
			t.Fatalf("%s %s: esperado %d, recibido %d %s", method, path, status, w.Code, w.Body.Bytes())
		}
		for _, cookie := range w.Result().Cookies() {
			if cookie.Name == auth.CSRFCookie {
				csrf = cookie
			}
		}
		return w.Body.Bytes()
	}
	type project struct {
		ID         string
		District   string `json:"distrito_id"`
		Estado     string
		Naturaleza string
		Sede       string
		Area       string `json:"area_id"`
		CSS        string `json:"correlativo_css"`
		AreaCSS    string `json:"area_css"`
	}
	read := func(raw []byte) project {
		t.Helper()
		var item project
		if err := json.Unmarshal(raw, &item); err != nil {
			t.Fatal(err)
		}
		return item
	}
	old := read(call("GET", base+"/00000000-0000-4000-8000-000000046003", "", false, 200))
	if old.District == "" || old.Sede == "" || old.CSS != "00050" || old.AreaCSS != "Área histórica literal" {
		t.Fatal("referencias de ejecución o sede ausentes")
	}
	const a = "00000000-0000-4000-8000-000000050002"
	const b = "00000000-0000-4000-8000-000000050003"
	body := `{"titulo":"Proyecto API #50","descripcion":"Verificación sintética","institucion_id":"00000000-0000-4000-8000-000000046001",
	"distrito_id":"` + old.District + `","tipo_id":"00000000-0000-4000-8000-000000050001","area_id":"` + a + `",
	"modalidad_id":"00000000-0000-4000-8000-000000050004","responsable":{"nombre":"Responsable de prueba","telefono":"2222-1111","extension":"004"}}`
	// Instituciones se desactiva en su propia prueba; restaura el permiso de alta con una institución independiente.
	body = strings.Replace(body, "00000000-0000-4000-8000-000000046001", "00000000-0000-4000-8000-000000046002", 1)
	call("GET", "/api/v1/auth/csrf", "", false, 204)
	call("POST", base, body, false, 403)
	item := read(call("POST", base, body, true, 201))
	if item.Estado != "ACTIVO" || item.Naturaleza != "EXTERNA" || item.District != old.District {
		t.Fatalf("alta inesperada: %+v", item)
	}
	call("POST", base, body, true, 409)
	item = read(call("PUT", base+"/"+item.ID, strings.Replace(body, a, b, 1), true, 200))
	if item.Area != b {
		t.Fatal("no cambió el área vigente")
	}
	var page struct {
		Proyectos []project
		Pagina    paginaCatalogo
	}
	if err := json.Unmarshal(call("GET", base+"?q=Proyecto+API&area_id="+a+"&area_id="+b+"&area_id="+b, "", false, 200), &page); err != nil {
		t.Fatal(err)
	}
	if page.Pagina.Total != 1 || len(page.Proyectos) != 1 || page.Proyectos[0].ID != item.ID {
		t.Fatalf("página inesperada: %+v", page)
	}
}
