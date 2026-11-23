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

	"horas-sociales/api/internal/modules/participaciones"
	"horas-sociales/api/internal/platform/auth"
)

func TestCierresAPIPostgreSQL(t *testing.T) {
	router, session := catalogoIntegracion(t, func(pool *pgxpool.Pool) func(chi.Router) {
		repo := participaciones.New(pool)
		return func(r chi.Router) { CierresRoutes(repo, true)(r); ParticipacionesRoutes(repo, true)(r) }
	})
	const base = "/api/v1/proyectos/00000000-0000-4000-8000-00000005300"
	const origin = "https://portal.test"
	var csrf *http.Cookie
	request := func(method, path, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, origin+path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Origin", origin)
		req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: session})
		if csrf != nil {
			req.AddCookie(csrf)
			req.Header.Set(auth.CSRFHeader, csrf.Value)
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		return w
	}
	for _, cookie := range request("GET", "/api/v1/auth/csrf", "").Result().Cookies() {
		if cookie.Name == auth.CSRFCookie {
			csrf = cookie
		}
	}
	if csrf == nil {
		t.Fatal("falta prueba CSRF")
	}
	loc, err := time.LoadLocation("America/El_Salvador")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().In(loc)
	today := now.Format(time.DateOnly)
	yesterday := now.AddDate(0, 0, -1).Format(time.DateOnly)
	body := func(action, date string) string {
		return `{"accion":"` + action + `","fecha_cierre":"` + date + `","motivo":"Cierre sintético"}`
	}
	expect := func(w *httptest.ResponseRecorder, status int, contains string) {
		t.Helper()
		if w.Code != status || (contains != "" && !strings.Contains(w.Body.String(), contains)) {
			t.Fatalf("esperado %d %q, recibido %d %s", status, contains, w.Code, w.Body.String())
		}
	}
	closeProject := func(project, action, date string) *httptest.ResponseRecorder {
		return request("POST", base+project+"/cierre", body(action, date))
	}
	closeEnrollment := func(project, row, action, date string) *httptest.ResponseRecorder {
		return request("POST", base+project+"/participaciones/00000000-0000-4000-8000-0000000530"+project+row+"/cierre", body(action, date))
	}
	list := func(project string) map[string]participaciones.Item {
		t.Helper()
		w := request("GET", base+project+"/participaciones", "")
		expect(w, 200, "")
		var payload struct {
			Items []participaciones.Item `json:"participaciones"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
			t.Fatal(err)
		}
		result := make(map[string]participaciones.Item)
		for _, item := range payload.Items {
			result[item.ID[len(item.ID)-1:]] = item
		}
		return result
	}
	expect(closeProject("1", "finalizar", today), 409, "participaciones_abiertas")
	expect(closeEnrollment("1", "1", "retirar", today), 200, `"estado_proyecto":"ACTIVO"`)
	expect(closeEnrollment("1", "2", "retirar", today), 200, "ABANDONADO")
	expect(closeProject("1", "cancelar", today), 409, "estado_proyecto")

	expect(closeEnrollment("2", "1", "retirar", today), 200, "RETIRADA")
	expect(closeEnrollment("2", "2", "finalizar", today), 200, `"estado_participacion":"FINALIZADA"`)
	expect(closeProject("2", "finalizar", today), 200, "FINALIZADO")
	if rows := list("2"); rows["1"].Estado != "RETIRADA" || rows["2"].Estado != "FINALIZADA" {
		t.Fatalf("se reescribió el historial: %+v", rows)
	}

	expect(closeEnrollment("3", "2", "retirar", today), 200, "RETIRADA")
	expect(closeEnrollment("3", "1", "retirar", yesterday), 400, "fecha_cierre")
	row := list("3")["1"]
	if row.Estado != "ASIGNADA" || row.Cerrada != nil || row.Actor != nil {
		t.Fatalf("el retiro inválido no hizo rollback: %+v", row)
	}
	expect(closeProject("3", "cancelar", today), 200, "CANCELADO")
	if rows := list("3"); rows["1"].Estado != "CANCELADA" || rows["2"].Estado != "RETIRADA" ||
		rows["1"].Actor == nil || rows["1"].FechaCierre == nil || *rows["1"].FechaCierre != today {
		t.Fatalf("cancelación inesperada: %+v", rows)
	}

	expect(closeProject("4", "finalizar", today), 409, "participaciones_abiertas")
	done := make(chan *httptest.ResponseRecorder, 1)
	go func() {
		done <- request("POST", base+"4/participaciones", `{"estudiante_id":"00000000-0000-4000-8000-000000059011"}`)
	}()
	expect(closeProject("4", "cancelar", today), 200, "CANCELADO")
	assignment := <-done
	if assignment.Code == 201 {
		for _, item := range list("4") {
			if item.Estado != "CANCELADA" {
				t.Fatalf("quedó participación abierta: %+v", item)
			}
		}
	} else {
		expect(assignment, 409, "estado_proyecto")
	}
}
