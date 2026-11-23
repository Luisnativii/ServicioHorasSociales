package httpserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"horas-sociales/api/internal/modules/instituciones"
	"horas-sociales/api/internal/platform/auth"
)

func TestInstitucionesAPIPostgreSQL(t *testing.T) {
	router, session := catalogoIntegracion(t, func(pool *pgxpool.Pool) func(chi.Router) {
		return InstitucionesRoutes(instituciones.NewRepository(pool))
	})
	const origin = "https://portal.test"
	const base = "/api/v1/instituciones"
	const histórica = "00000000-0000-4000-8000-000000046001"
	var csrf *http.Cookie
	call := func(method, path, body, jwt string, proof bool, status int) []byte {
		t.Helper()
		req := httptest.NewRequest(method, origin+path, strings.NewReader(body))
		req.Header.Set("Origin", origin)
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
		if res.Code != status || res.Header().Get("Cache-Control") != "no-store" {
			t.Fatalf("%s %s: esperado %d, recibido %d %s", method, path, status, res.Code, res.Body.String())
		}
		if path == "/api/v1/auth/csrf" {
			for _, cookie := range res.Result().Cookies() {
				if cookie.Name == auth.CSRFCookie {
					csrf = cookie
				}
			}
		}
		return res.Body.Bytes()
	}
	decode := func(body []byte, target any) {
		t.Helper()
		if err := json.Unmarshal(body, target); err != nil {
			t.Fatal(err)
		}
	}
	assertField := func(body []byte, field string) {
		t.Helper()
		var result struct{ Error errorDetail }
		decode(body, &result)
		if result.Error.Field != field {
			t.Fatalf("esperaba error en %s: %s", field, body)
		}
	}
	call("GET", "/api/v1/auth/csrf", "", session, false, 204)
	if csrf == nil {
		t.Fatal("falta cookie de CSRF")
	}
	call("GET", base, "", "", false, 401)
	call("POST", base, `{"nombre":"Prueba"}`, session, false, 403)

	var existing institucionRespuesta
	decode(call("GET", base+"/"+histórica, "", session, false, 200), &existing)
	if existing.CodigoCSS == nil || *existing.CodigoCSS != "00046" || existing.Sede == nil {
		t.Fatalf("institución histórica inesperada: %+v", existing)
	}
	assertField(call("POST", base, `{"nombre":" "}`, session, true, 400), "nombre")
	assertField(call("POST", base, `{"nombre":"Prueba","distrito_id":"invalido"}`, session, true, 400), "distrito_id")
	assertField(call("POST", base, `{"nombre":"Prueba","distrito_id":"00000000-0000-4000-8000-000000047099"}`, session, true, 400), "distrito_id")
	call("POST", base, `{"nombre":"Prueba","codigo_css":"666"}`, session, true, 400)
	call("GET", base+"?page_size=101", "", session, false, 400)

	var item institucionRespuesta
	decode(call("POST", base, `{"nombre":"  Institución API prueba  ","distrito_id":"`+existing.Sede.DistritoID+`"}`,
		session, true, 201), &item)
	if item.ID == "" || item.Nombre != "Institución API prueba" || !item.Activo ||
		item.Sede == nil || *item.Sede != *existing.Sede || item.CodigoCSS != nil {
		t.Fatalf("alta inesperada: %+v", item)
	}
	assertField(call("POST", base, `{"nombre":"INSTITUCION API PRUEBA"}`, session, true, 409), "nombre")
	call("POST", base, `{"nombre":"Institución API sin sede"}`, session, true, 201)
	itemPath := base + "/" + item.ID
	call("PUT", itemPath, `{"nombre":"Institución API editada"}`, session, false, 403)
	decode(call("PUT", itemPath, `{"nombre":"Institución API editada","distrito_id":null}`, session, true, 200), &item)
	if item.Nombre != "Institución API editada" || item.Sede != nil || !item.Activo {
		t.Fatalf("edición inesperada: %+v", item)
	}
	page := func(query string) (items []institucionRespuesta, meta paginaCatalogo) {
		t.Helper()
		var result struct {
			Instituciones []institucionRespuesta
			Pagina        paginaCatalogo
		}
		decode(call("GET", base+query, "", session, false, 200), &result)
		return result.Instituciones, result.Pagina
	}
	items, meta := page("?q=api&page_size=1&page=2")
	if len(items) != 1 || items[0].Nombre != "Institución API sin sede" || meta != (paginaCatalogo{2, 1, 2}) {
		t.Fatalf("paginación inesperada: %+v %+v", items, meta)
	}
	items, meta = page("?q=api&page_size=1&page=3")
	if items == nil || len(items) != 0 || meta.Total != 2 {
		t.Fatalf("página vacía inesperada: %+v %+v", items, meta)
	}
	items, meta = page("?q=00046&incluir_inactivas=true")
	if len(items) != 1 || items[0].ID != histórica || meta.Total != 1 {
		t.Fatalf("búsqueda por código inesperada: %+v %+v", items, meta)
	}
	call("POST", itemPath+"/desactivar", "", session, false, 403)
	decode(call("POST", itemPath+"/desactivar", "", session, true, 200), &item)
	if item.Activo {
		t.Fatal("la institución sigue activa")
	}
	decode(call("GET", itemPath, "", session, false, 200), &item)
	if item.Activo || item.Nombre != "Institución API editada" {
		t.Fatalf("lectura histórica inesperada: %+v", item)
	}
	items, meta = page("?q=api")
	if len(items) != 1 || meta.Total != 1 {
		t.Fatalf("listado de activas inesperado: %+v %+v", items, meta)
	}
	items, meta = page("?q=api&incluir_inactivas=true")
	if len(items) != 2 || meta.Total != 2 {
		t.Fatalf("listado histórico inesperado: %+v %+v", items, meta)
	}
}
