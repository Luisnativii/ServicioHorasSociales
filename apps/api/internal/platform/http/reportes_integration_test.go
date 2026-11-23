package httpserver

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"horas-sociales/api/internal/modules/reportes"
	"horas-sociales/api/internal/platform/auth"
)

func TestReportesAPIPostgreSQL(t *testing.T) {
	router, session := catalogoIntegracion(t, func(pool *pgxpool.Pool) func(chi.Router) {
		return ReportesRoutes(reportes.New(pool))
	})
	const path = "/api/v1/reportes/actividad"
	const filter = "?desde=2021-01-01&hasta=2021-02-03"
	request := func(url string, authorized bool) *httptest.ResponseRecorder {
		r := httptest.NewRequest("GET", "https://portal.test"+url, nil)
		if authorized {
			r.AddCookie(&http.Cookie{Name: auth.CookieName, Value: session})
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)
		return w
	}
	type report struct {
		Summary reportes.Summary `json:"resumen"`
		Rows    []reportes.Row   `json:"movimientos"`
		Page    paginaCatalogo   `json:"pagina"`
	}
	read := func(query string) report {
		t.Helper()
		var result report
		w := request(path+query, true)
		if w.Code != 200 {
			t.Fatalf("estado %d: %s", w.Code, w.Body.String())
		}
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		return result
	}
	result := read(filter)
	expected := reportes.Summary{Students: 1, Participations: 2, Projects: 2, Internal: 300, External: 300, Completed: 1}
	if result.Summary != expected || result.Page.Total != 3 || len(result.Rows) != 3 {
		t.Fatalf("resultado inesperado: %+v", result)
	}
	first, ambiguous, legacy := result.Rows[2], result.Rows[1], result.Rows[0]
	if first.Carnet != "00007401" || first.Career == nil || *first.Career != "Carrera Académica A" ||
		first.Area == nil || *first.Area != "Área sintética A #50" || ambiguous.Career != nil || ambiguous.Type != nil ||
		legacy.Area != nil || legacy.LegacyArea == nil || *legacy.LegacyArea != "Área original literal #74" {
		t.Fatalf("atribución inesperada: %+v", result.Rows)
	}
	const area = "00000000-0000-4000-8000-000000050002"
	narrow := read(filter + "&area_id=" + area + "&area_id=" + area + "&page_size=1")
	if narrow.Summary.Students != 1 || narrow.Summary.Internal != 300 || narrow.Summary.External != 0 ||
		narrow.Page.Total != 1 || len(narrow.Rows) != 1 {
		t.Fatalf("filtro produjo duplicados: %+v", narrow)
	}
	outside := read("?desde=2021-02-02&hasta=2021-02-02")
	if outside.Summary.Completed != 0 || outside.Summary.External != 250 {
		t.Fatalf("cumplimiento adelantado: %+v", outside)
	}
	for _, check := range []struct {
		query      string
		authorized bool
		status     int
	}{
		{path + ".xlsx" + filter, false, 401},
		{path + "?desde=2021-02-01&hasta=2021-01-01", true, 400},
		{path + ".xlsx" + filter + "&page=1", true, 400},
		{path + ".xlsx?desde=2022-02-01&hasta=2022-02-01", true, 413},
	} {
		if w := request(check.query, check.authorized); w.Code != check.status || w.Header().Get("Content-Disposition") != "" {
			t.Fatalf("respuesta inesperada %d en %s", w.Code, check.query)
		}
	}
	w := request(path+".xlsx"+filter, true)
	if w.Code != 200 || !strings.HasPrefix(w.Header().Get("Content-Disposition"), "attachment") ||
		w.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("descarga inesperada: %d", w.Code)
	}
	z, err := zip.NewReader(bytes.NewReader(w.Body.Bytes()), int64(w.Body.Len()))
	if err != nil {
		t.Fatal(err)
	}
	load := func(file string) string {
		t.Helper()
		part, err := z.Open(file)
		if err != nil {
			t.Fatal(err)
		}
		defer part.Close()
		data, err := io.ReadAll(part)
		if err != nil {
			t.Fatal(err)
		}
		return string(data)
	}
	detail, summary := load("xl/worksheets/sheet2.xml"), load("xl/worksheets/sheet1.xml")
	for _, fragment := range []string{`r="A2" t="inlineStr"><is><t xml:space="preserve">00007401`,
		`r="E2" t="n"><v>50</v>`, ">=No ejecutar</t>", ">=Proyecto sintético #74</t>", "Sin dato histórico"} {
		if !strings.Contains(detail, fragment) {
			t.Errorf("falta %q en el Excel", fragment)
		}
	}
	if strings.Contains(detail, "<f>") {
		t.Fatal("un valor de texto se escribió como fórmula")
	}
	for _, fragment := range []string{`r="B2" t="n"><v>1</v>`, `r="B3" t="n"><v>2</v>`,
		`r="B5" t="n"><v>300</v>`, `r="B6" t="n"><v>300</v>`, `r="B7" t="n"><v>1</v>`} {
		if !strings.Contains(summary, fragment) {
			t.Errorf("resumen de dashboard y Excel difieren: %s", fragment)
		}
	}
}
