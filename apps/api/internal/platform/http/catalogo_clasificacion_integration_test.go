package httpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"horas-sociales/api/internal/modules/catalogos"
	"horas-sociales/api/internal/platform/auth"
)

func TestCatalogoClasificacionPostgreSQL(t *testing.T) {
	if os.Getenv("HS_TEST_CATALOGS") != "1" {
		t.Skip("requiere la base efímera de scripts/check-catalogs.sh")
	}
	ctx := context.Background()
	config, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
	if err != nil || config.ConnConfig.Host != "db" || config.ConnConfig.User != "hs_app" ||
		!strings.HasPrefix(config.ConnConfig.Database, "hs_catalogs_probe_") {
		t.Fatal("se exige la base efímera de catálogos con cuenta hs_app")
	}
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatal("no se pudo preparar PostgreSQL de pruebas")
	}
	defer pool.Close()

	const origen = "https://portal.test"
	tokens, err := auth.New(bytes.Repeat([]byte{42}, 32))
	if err != nil {
		t.Fatal(err)
	}
	browser, err := auth.NewBrowser(tokens, origen)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := tokens.Issue("00000000-0000-4000-8000-000000000042")
	if err != nil {
		t.Fatal(err)
	}
	session := httptest.NewRecorder()
	browser.SetSession(session, raw)
	router := Router(pool, Options{Tokens: tokens, Browser: browser, RegisterPrivate: CatalogoClasificacionRoutes(catalogos.NewRepository(pool))})
	request := func(method, path, body string, jwt, csrf bool, status int) []byte {
		t.Helper()
		req := httptest.NewRequest(method, origen+"/api/v1/catalogos"+path, strings.NewReader(body))
		req.Header.Set("Origin", origen)
		req.Header.Set("Content-Type", "application/json")
		for _, cookie := range session.Result().Cookies() {
			switch {
			case cookie.Name == auth.CookieName && jwt:
				req.AddCookie(cookie)
			case cookie.Name == auth.CSRFCookie && csrf:
				req.AddCookie(cookie)
				req.Header.Set(auth.CSRFHeader, cookie.Value)
			}
		}
		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)
		if res.Code != status || res.Header().Get("Cache-Control") != "no-store" {
			t.Fatalf("%s %s: se esperaba %d sin caché, se recibió %d %s", method, path, status, res.Code, res.Body.String())
		}
		return res.Body.Bytes()
	}
	decode := func(body []byte, value any) {
		t.Helper()
		if err := json.Unmarshal(body, value); err != nil {
			t.Fatal(err)
		}
	}
	const tipos = "/tipos-proyecto"
	const areas = "/areas"
	ids := make(map[string]string)
	for _, path := range []string{tipos, areas} {
		request("GET", path, "", false, false, 401)
		request("POST", path, `{"nombre":"Apoyo"}`, false, false, 401)
		request("POST", path, `{"nombre":"Apoyo"}`, true, false, 403)
		request("POST", path, `{"nombre":"\u0000"}`, true, true, 400)
		request("POST", path, `{"nombre":"omitido","activo":false}`, true, true, 400)
		request("GET", path+"?incluir_inactivos=1", "", true, false, 400)
		request("GET", path+"?page_size=101", "", true, false, 400)
		request("GET", path+"/errado", "", true, false, 400)
		request("GET", path+"/00000000-0000-4000-8000-000000000099", "", true, false, 404)

		var original, segundo clasificacionRespuesta
		decode(request("POST", path, `{"nombre":"  Apoyo  "}`, true, true, 201), &original)
		decode(request("POST", path, `{"nombre":"Apoyo % Social"}`, true, true, 201), &segundo)
		if original.ID == "" || original.Nombre != "Apoyo" || !original.Activo {
			t.Fatalf("alta inesperada: %+v", original)
		}
		ids[path] = original.ID
		request("POST", path, `{"nombre":"aPOYO"}`, true, true, 409)
		request("PATCH", path+"/"+segundo.ID, `{"nombre":"APOYO"}`, true, true, 409)
		request("PATCH", path+"/"+segundo.ID, `{"nombre":"Apoyo % Comunitario"}`, true, false, 403)
		decode(request("PATCH", path+"/"+segundo.ID, `{"nombre":"Apoyo % Comunitario"}`, true, true, 200), &segundo)
		if segundo.Nombre != "Apoyo % Comunitario" {
			t.Fatalf("edición inesperada: %+v", segundo)
		}

		var lista struct {
			Opciones []clasificacionRespuesta
			Pagina   paginaCatalogo
		}
		decode(request("GET", path+"?q=APOYO&page_size=1&page=2", "", true, false, 200), &lista)
		if len(lista.Opciones) != 1 || lista.Opciones[0] != segundo || lista.Pagina != (paginaCatalogo{2, 1, 2}) {
			t.Fatalf("página inesperada: %+v", lista)
		}
		decode(request("GET", path+"?q=%25", "", true, false, 200), &lista)
		if len(lista.Opciones) != 1 || lista.Opciones[0] != segundo || lista.Pagina.Total != 1 {
			t.Fatalf("se esperaba búsqueda literal de %%: %+v", lista)
		}
		decode(request("GET", path+"?page_size=1&page=3", "", true, false, 200), &lista)
		if lista.Opciones == nil || len(lista.Opciones) != 0 || lista.Pagina.Total != 2 {
			t.Fatalf("se esperaba página vacía y total preservado: %+v", lista)
		}
	}
	if _, err := pool.Exec(ctx, `INSERT INTO public.hs_catalogs_reference (tipo_id, area_id) VALUES ($1, $2)`, ids[tipos], ids[areas]); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{tipos, areas} {
		id := ids[path]
		request("POST", path+"/"+id+"/desactivar", "", true, false, 403)
		request("DELETE", path+"/"+id, "", true, true, 405)
		var historical clasificacionRespuesta
		for range 2 {
			decode(request("POST", path+"/"+id+"/desactivar", "", true, true, 200), &historical)
			if historical.ID != id || historical.Activo || historical.Nombre != "Apoyo" {
				t.Fatalf("desactivación inesperada: %+v", historical)
			}
		}
		decode(request("GET", path+"/"+id, "", true, false, 200), &historical)
		if historical.ID != id || historical.Activo {
			t.Fatalf("detalle histórico inesperado: %+v", historical)
		}
		var lista struct {
			Opciones []clasificacionRespuesta
			Pagina   paginaCatalogo
		}
		decode(request("GET", path, "", true, false, 200), &lista)
		if len(lista.Opciones) != 1 || !lista.Opciones[0].Activo || lista.Pagina.Total != 1 {
			t.Fatalf("listado activo inesperado: %+v", lista)
		}
		decode(request("GET", path+"?incluir_inactivos=true", "", true, false, 200), &lista)
		if len(lista.Opciones) != 2 || lista.Opciones[0] != historical || lista.Pagina.Total != 2 {
			t.Fatalf("listado histórico inesperado: %+v", lista)
		}
	}
	var references int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM public.hs_catalogs_reference AS r
		JOIN horas_sociales.tipo_proyecto AS t ON t.id = r.tipo_id
		JOIN horas_sociales.area AS a ON a.id = r.area_id
		WHERE NOT t.activo AND NOT a.activo`).Scan(&references); err != nil || references != 1 {
		t.Fatalf("la desactivación debe conservar las referencias: %d, %v", references, err)
	}
}
