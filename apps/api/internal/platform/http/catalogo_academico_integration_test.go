package httpserver

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"horas-sociales/api/internal/modules/catalogos"
	"horas-sociales/api/internal/platform/auth"
)

func TestCatalogoAcademicoPostgreSQL(t *testing.T) {
	if os.Getenv("HS_TEST_CORE_CATALOG") != "1" {
		t.Skip("requiere PostgreSQL aislado con los datos de scripts/check-core-readiness.sh")
	}
	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		t.Fatal("no se pudo preparar PostgreSQL de pruebas")
	}
	defer pool.Close()

	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatal(err)
	}
	tokens, err := auth.New(key)
	if err != nil {
		t.Fatal(err)
	}
	browser, err := auth.NewBrowser(tokens, "https://portal.test")
	if err != nil {
		t.Fatal(err)
	}
	cookie, err := tokens.Issue("00000000-0000-4000-8000-000000000035")
	if err != nil {
		t.Fatal(err)
	}
	router := Router(pool, Options{
		Tokens: tokens, Browser: browser,
		RegisterPrivate: CatalogoAcademicoRoutes(catalogos.NewAcademiaRepository(pool)),
	})
	const base = "/api/v1/catalogos"
	const facultad = "00000000-0000-4000-8000-000000035001"
	const carrera = "00000000-0000-4000-8000-000000035002"
	const facultadInactiva = "00000000-0000-4000-8000-000000035003"
	const carreraHistorica = "00000000-0000-4000-8000-000000035004"
	request := func(method, route, session string, status int) []byte {
		t.Helper()
		req := httptest.NewRequest(method, "https://portal.test"+base+route, nil)
		if session != "" {
			req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: session})
		}
		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)
		if res.Code != status || res.Header().Get("Cache-Control") != "no-store" {
			t.Fatalf("%s %s: se esperaba %d sin caché, se recibió %d %s", method, route, status, res.Code, res.Body.String())
		}
		return res.Body.Bytes()
	}
	unmarshal := func(data []byte, dst any) {
		t.Helper()
		if err := json.Unmarshal(data, dst); err != nil {
			t.Fatal(err)
		}
	}

	for _, path := range []string{"/facultades", "/carreras", "/facultades/" + facultad, "/carreras/" + carrera} {
		request(http.MethodGet, path, "", http.StatusUnauthorized)
		request(http.MethodGet, path, "invalida", http.StatusUnauthorized)
		request(http.MethodPost, path, cookie, http.StatusMethodNotAllowed)
	}
	for _, path := range []string{
		"/facultades?page_size=101", "/facultades?page=1&page=2", "/facultades?order=desc",
		"/facultades?incluir_inactivas=1", "/carreras?facultad_id=invalido", "/carreras?q=a&q=b",
		"/facultades?facultad_id=" + facultad, "/facultades?q=%00",
		"/carreras/00000000-0000-0000-0000-000000000000",
	} {
		request(http.MethodGet, path, cookie, http.StatusBadRequest)
	}
	request(http.MethodGet, "/carreras/00000000-0000-4000-8000-000000035099", cookie, http.StatusNotFound)

	var facultades struct {
		Facultades []facultadRespuesta
		Pagina     paginaCatalogo
	}
	unmarshal(request(http.MethodGet, "/facultades?q=FACULTAD+ACADEMICA&incluir_inactivas=true&page_size=1&page=2", cookie, http.StatusOK), &facultades)
	if len(facultades.Facultades) != 1 || facultades.Facultades[0].ID != facultadInactiva ||
		facultades.Pagina != (paginaCatalogo{Page: 2, PageSize: 1, Total: 2}) {
		t.Fatalf("paginación de facultades: %+v", facultades)
	}

	var carreras struct {
		Carreras []carreraRespuesta
		Pagina   paginaCatalogo
	}
	for _, test := range []struct {
		Query string
		ID    string
	}{
		{"?q=CARRERA+ACADEMICA", carrera},
		{"?q=CARRERA+ACADEMICA&facultad_id=" + facultad, carrera},
		{"?q=CARRERA+ACADEMICA&facultad_id=" + facultadInactiva, ""},
		{"?q=CARRERA+ACADEMICA&facultad_id=" + facultadInactiva + "&incluir_inactivas=true", carreraHistorica},
	} {
		unmarshal(request(http.MethodGet, "/carreras"+test.Query, cookie, http.StatusOK), &carreras)
		if test.ID == "" {
			if carreras.Carreras == nil || len(carreras.Carreras) != 0 || carreras.Pagina.Total != 0 {
				t.Fatalf("listado esperado vacío para %s: %+v", test.Query, carreras)
			}
		} else if len(carreras.Carreras) != 1 || carreras.Carreras[0].ID != test.ID || carreras.Pagina.Total != 1 {
			t.Fatalf("listado filtrado para %s: %+v", test.Query, carreras)
		}
	}

	var f facultadRespuesta
	unmarshal(request(http.MethodGet, "/facultades/"+facultadInactiva, cookie, http.StatusOK), &f)
	if f.ID != facultadInactiva || f.Activo {
		t.Fatalf("facultad histórica: %+v", f)
	}
	var c carreraRespuesta
	unmarshal(request(http.MethodGet, "/carreras/"+carreraHistorica, cookie, http.StatusOK), &c)
	if c.ID != carreraHistorica || c.FacultadID != facultadInactiva || c.FacultadActiva || !c.Activo {
		t.Fatalf("relación histórica: %+v", c)
	}
}
