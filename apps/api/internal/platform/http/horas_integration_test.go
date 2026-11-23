package httpserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	_ "time/tzdata"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"horas-sociales/api/internal/modules/horas"
	"horas-sociales/api/internal/modules/participaciones"
	"horas-sociales/api/internal/platform/auth"
)

func TestHorasAPIPostgreSQL(t *testing.T) {
	router, session := catalogoIntegracion(t, func(pool *pgxpool.Pool) func(chi.Router) {
		return func(r chi.Router) {
			HorasRoutes(horas.New(pool), true)(r)
			CierresRoutes(participaciones.New(pool), true)(r)
		}
	})
	const origin = "https://portal.test"
	const ledger = "/api/v1/estudiantes/00000000-0000-4000-8000-000000059011/horas"
	const project1 = "/api/v1/proyectos/00000000-0000-4000-8000-000000065001"
	const project2 = "/api/v1/proyectos/00000000-0000-4000-8000-000000065002"
	const endpoint1 = project1 + "/participaciones/00000000-0000-4000-8000-000000065011/horas"
	const endpoint2 = project2 + "/participaciones/00000000-0000-4000-8000-000000065012/horas"
	const key = "00000000-0000-4000-8000-000000065021"
	const second = "00000000-0000-4000-8000-000000065022"
	var csrf *http.Cookie
	call := func(method, path, body string, proof bool) *httptest.ResponseRecorder {
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
		return w
	}
	for _, c := range call("GET", "/api/v1/auth/csrf", "", false).Result().Cookies() {
		if c.Name == auth.CSRFCookie {
			csrf = c
		}
	}
	if csrf == nil {
		t.Fatal("sin prueba CSRF")
	}
	loc, err := time.LoadLocation("America/El_Salvador")
	if err != nil {
		t.Fatal(err)
	}
	today := time.Now().In(loc).Format(time.DateOnly)
	input := func(n, date, k string) string {
		return `{"horas":` + n + `,"fecha_horas":"` + date + `","concepto":"Acreditación sintética","solicitud_id":"` + k + `"}`
	}
	check := func(w *httptest.ResponseRecorder, status int, field string) {
		t.Helper()
		if w.Code != status || (field != "" && !strings.Contains(w.Body.String(), `"field":"`+field+`"`)) {
			t.Fatalf("esperado %d %s: recibido %d %s", status, field, w.Code, w.Body.String())
		}
	}
	check(call("POST", endpoint1, input("350", today, key), false), 403, "")
	check(call("POST", endpoint1, input("1.5", today, key), true), 400, "")
	check(call("POST", endpoint1, input("1", "2099-01-01", key), true), 400, "fecha_horas")
	check(call("POST", endpoint1, input("1", "2000-01-01", key), true), 400, "fecha_horas")
	outcomes := make(chan *httptest.ResponseRecorder, 2)
	for range 2 {
		go func() { outcomes <- call("POST", endpoint1, input("350", today, key), true) }()
	}
	ids := make(map[string]bool)
	for range 2 {
		w := <-outcomes
		var receipt horas.Receipt
		if err := json.Unmarshal(w.Body.Bytes(), &receipt); err != nil {
			t.Fatal(err)
		}
		if receipt.ID == "" || (w.Code != 201 && w.Code != 200) || receipt.Replayed != (w.Code == 200) {
			t.Fatalf("recibo inesperado: %d %s", w.Code, w.Body.String())
		}
		ids[receipt.ID] = receipt.Replayed
	}
	if len(ids) != 1 {
		t.Fatalf("movimientos duplicados: %+v", ids)
	}
	check(call("POST", endpoint1, input("351", today, key), true), 409, "solicitud_id")
	check(call("POST", endpoint2, input("350", today, key), true), 409, "solicitud_id")
	var page struct {
		Movimientos []horas.Movement
		Avance      horas.Progress
		Pagina      paginaCatalogo
	}
	read := func(url string) {
		t.Helper()
		w := call("GET", url, "", false)
		check(w, 200, "")
		if err := json.Unmarshal(w.Body.Bytes(), &page); err != nil {
			t.Fatal(err)
		}
	}
	read(ledger + "?page_size=1")
	if page.Pagina.Total != 2 || page.Avance.Internal != 350 || page.Avance.External != 20 ||
		page.Avance.InternalMissing != 0 || page.Avance.ExternalMissing != 280 ||
		len(page.Movimientos) != 1 || page.Movimientos[0].PortalActor == nil || page.Movimientos[0].CoreActor != nil ||
		page.Movimientos[0].Origin != "MANUAL" {
		t.Fatalf("libro inesperado: %+v", page)
	}
	read(ledger + "?page_size=1&page=2")
	if len(page.Movimientos) != 1 || page.Movimientos[0].Origin != "IMPORTACION" || page.Movimientos[0].Nature != "EXTERNA" {
		t.Fatalf("histórico inesperado: %+v", page)
	}
	closeBody := `{"accion":"cancelar","fecha_cierre":"` + today + `","motivo":"Cierre sintético"}`
	check(call("POST", project1+"/cierre", closeBody, true), 200, "")
	check(call("POST", endpoint1, input("350", today, key), true), 200, "")
	check(call("POST", endpoint1, input("1", today, second), true), 409, "estado_participacion")
	race := make(chan *httptest.ResponseRecorder, 1)
	go func() { race <- call("POST", endpoint2, input("10", today, second), true) }()
	check(call("POST", project2+"/cierre", closeBody, true), 200, "")
	w := <-race
	if w.Code != 201 {
		check(w, 409, "estado_participacion")
	}
	read(ledger)
	want := int64(20)
	if w.Code == 201 {
		want += 10
	}
	if page.Avance.External != want {
		t.Fatalf("saldo incoherente tras el cierre: %+v", page)
	}
}
