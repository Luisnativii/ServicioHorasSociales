package httpserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"horas-sociales/api/internal/modules/catalogos"
	"horas-sociales/api/internal/platform/auth"
)

func TestCatalogoTerritorialPostgreSQL(t *testing.T) {
	router, cookie := catalogoIntegracion(t, func(pool *pgxpool.Pool) func(chi.Router) {
		return CatalogoTerritorialRoutes(catalogos.NewTerritorioRepository(pool))
	})
	request := func(path, session string, status int) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/catalogos"+path, nil)
		if session != "" {
			req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: session})
		}
		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)
		if res.Code != status || res.Header().Get("Cache-Control") != "no-store" {
			t.Fatalf("%s: se esperaba %d sin caché, se recibió %d %s", path, status, res.Code, res.Body.String())
		}
		return res
	}
	listar := func(path string) ([]catalogos.LugarTerritorial, paginaCatalogo) {
		t.Helper()
		var respuesta struct {
			Lugares []catalogos.LugarTerritorial
			Pagina  paginaCatalogo
		}
		if err := json.Unmarshal(request(path, cookie, 200).Body.Bytes(), &respuesta); err != nil {
			t.Fatal(err)
		}
		return respuesta.Lugares, respuesta.Pagina
	}
	for _, path := range []string{"/departamentos", "/municipios", "/distritos"} {
		request(path, "", 401)
		request(path, "invalida", 401)
	}
	for _, path := range []string{
		"/departamentos?page_size=101", "/departamentos?municipio_id=1",
		"/municipios", "/municipios?departamento_id=00000000-0000-0000-0000-000000000000",
		"/distritos?municipio_id=invalido", "/distritos?municipio_id=x&municipio_id=y",
	} {
		request(path, cookie, 400)
	}
	departamentos, pagina := listar("/departamentos?page_size=1")
	if len(departamentos) != 1 || departamentos[0].Codigo != "01" || departamentos[0].PadreID != nil ||
		pagina != (paginaCatalogo{Page: 1, PageSize: 1, Total: 14}) {
		t.Fatalf("primera página inesperada: %+v %+v", departamentos, pagina)
	}
	vacios, pagina := listar("/departamentos?page_size=1&page=15")
	if vacios == nil || len(vacios) != 0 || pagina.Total != 14 {
		t.Fatalf("última página inesperada: %+v %+v", vacios, pagina)
	}
	departamento := departamentos[0].ID
	municipios, pagina := listar("/municipios?departamento_id=" + departamento)
	if len(municipios) != 3 || pagina.Total != 3 || municipios[2].Codigo != "03" {
		t.Fatalf("municipios esperados de Ahuachapán: %+v %+v", municipios, pagina)
	}
	for _, m := range municipios {
		if m.PadreID == nil || *m.PadreID != departamento {
			t.Fatalf("municipio ajeno: %+v", m)
		}
	}
	municipio := municipios[2].ID
	distritos, pagina := listar("/distritos?municipio_id=" + municipio)
	if len(distritos) != 4 || pagina.Total != 4 {
		t.Fatalf("distritos esperados de Ahuachapán Sur: %+v %+v", distritos, pagina)
	}
	found := false
	for _, d := range distritos {
		if d.PadreID == nil || *d.PadreID != municipio {
			t.Fatalf("distrito ajeno: %+v", d)
		}
		found = found || (d.Codigo == "08" && d.Nombre == "San Francisco Menéndez")
	}
	if !found {
		t.Fatal("no se encontró San Francisco Menéndez en Ahuachapán Sur")
	}
	vacios, pagina = listar("/distritos?municipio_id=" + departamento)
	if vacios == nil || len(vacios) != 0 || pagina.Total != 0 {
		t.Fatalf("se esperaban cero distritos para un UUID de otro nivel: %+v %+v", vacios, pagina)
	}
}
