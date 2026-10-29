package httpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"horas-sociales/api/internal/modules/admin"
	"horas-sociales/api/internal/platform/auth"
	"horas-sociales/api/internal/platform/password"
)

func cuentasRouterPG(t *testing.T, pool *pgxpool.Pool) (*auth.Tokens, *auth.Browser, string) {
	clave, err := os.CreateTemp("", "hs-cuentas-test-key-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(clave.Name())
	if _, err := clave.Write(bytes.Repeat([]byte{71}, 32)); err != nil {
		t.Fatal(err)
	}
	if err := clave.Chmod(0600); err != nil {
		t.Fatal(err)
	}
	clave.Close()
	tokens, err := auth.LoadKey(clave.Name())
	if err != nil {
		t.Fatal(err)
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	srv := &http.Server{Handler: http.NotFoundHandler()}
	go srv.Serve(ln)
	t.Cleanup(func() { srv.Close() })
	origen := "http://" + ln.Addr().String()
	browser, err := auth.NewBrowser(tokens, origen)
	if err != nil {
		t.Fatal(err)
	}
	buscar := admin.NewRepository(pool)
	srv.Handler = Router(pool, Options{Tokens: tokens, Browser: browser, Buscador: buscar})
	return tokens, browser, origen
}

func csrfCookiePGConAcceso(t *testing.T, origen, acceso string) string {
	req, err := http.NewRequest("GET", origen+"/api/v1/auth/csrf", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Host = strings.TrimPrefix(origen, "http://")
	req.Header.Set("Origin", origen)
	req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: acceso})
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("csrf con acceso: %d", res.StatusCode)
	}
	for _, c := range res.Cookies() {
		if c.Name == auth.CSRFCookie {
			return c.Value
		}
	}
	t.Fatal("cookie CSRF no preparada con acceso")
	return ""
}

func doCuentas(t *testing.T, origen, metodo, ruta string, acceso, csrf string) (*http.Response, []byte) {
	req, err := http.NewRequest(metodo, origen+ruta, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Host = strings.TrimPrefix(origen, "http://")
	req.Header.Set("Origin", origen)
	if acceso != "" {
		req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: acceso})
	}
	if csrf != "" {
		req.AddCookie(&http.Cookie{Name: auth.CSRFCookie, Value: csrf})
		req.Header.Set("X-CSRF-Token", csrf)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	cuerpoRes, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	return res, cuerpoRes
}

func idInexistente(t *testing.T) pgtype.UUID {
	t.Helper()
	var id pgtype.UUID
	copy(id.Bytes[:], []byte{0xde, 0xad, 0xbe, 0xef, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1})
	id.Valid = true
	return id
}

func uuidString(id pgtype.UUID) string {
	b := id.Bytes
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func TestCuentasPostgreSQL(t *testing.T) {
	if os.Getenv("HS_TEST_CUENTAS") != "1" || os.Getenv("DATABASE_URL") == "" {
		t.Skip("integración explícita: HS_TEST_CUENTAS=1 y DATABASE_URL hacia hs-test")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		t.Fatal("no se pudo conectar a hs-test: ", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		t.Fatal("hs-test no responde: ", err)
	}
	buscar := admin.NewRepository(pool)

	hash, err := password.Hash(ctx, []byte("contrasena1234"))
	if err != nil {
		t.Fatal(err)
	}
	uno, err := buscar.Create(ctx, " CUENTAUNO@EJEMPLO.TEST ", hash)
	if err != nil || !uno.ID.Valid || !uno.Activo {
		t.Fatalf("alta cuenta uno: %v", err)
	}
	dos, err := buscar.Create(ctx, " cuentados@ejemplo.test ", hash)
	if err != nil || !dos.ID.Valid {
		t.Fatalf("alta cuenta dos: %v", err)
	}
	if _, err := buscar.SetActive(ctx, dos.ID, false); err != nil {
		t.Fatalf("desactivar cuenta dos: %v", err)
	}
	defer func() {
		pool.Exec(ctx, `DELETE FROM horas_sociales.administrador WHERE correo IN ('cuentauno@ejemplo.test','cuentados@ejemplo.test')`)
	}()

	tokens, _, origen := cuentasRouterPG(t, pool)
	actor := tokenValidoPG(t, tokens, uno.ID)
	csrf := csrfCookiePG(t, origen)

	t.Run("listar cuentas con paginación", func(t *testing.T) {
		res, cuerpo := doCuentas(t, origen, "GET", "/api/v1/admin/cuentas", actor, csrf)
		if res.StatusCode != http.StatusOK {
			t.Fatalf("listar: %d %s", res.StatusCode, cuerpo)
		}
		var respuesta listarCuentasRespuesta
		if err := json.Unmarshal(cuerpo, &respuesta); err != nil {
			t.Fatal(err)
		}
		if len(respuesta.Cuentas) != 2 || respuesta.Pagina.Total != 2 {
			t.Fatalf("listado inesperado: %+v", respuesta)
		}
		encontrados := map[string]cuentaRespuesta{}
		for _, c := range respuesta.Cuentas {
			encontrados[c.ID] = c
		}
		if c, ok := encontrados[uno.ID.String()]; !ok || !c.Activo || c.Correo != "cuentauno@ejemplo.test" {
			t.Fatalf("cuenta uno inesperada: %+v", c)
		}
		if c, ok := encontrados[dos.ID.String()]; !ok || c.Activo || c.Correo != "cuentados@ejemplo.test" {
			t.Fatalf("cuenta dos inesperada: %+v", c)
		}
	})

	t.Run("activar cuenta inactiva", func(t *testing.T) {
		csrfAcceso := csrfCookiePGConAcceso(t, origen, actor)
		ruta := "/api/v1/admin/cuentas/" + uuidString(dos.ID) + "/activar"
		res, cuerpo := doCuentas(t, origen, "POST", ruta, actor, csrfAcceso)
		if res.StatusCode != http.StatusOK {
			t.Fatalf("activar: %d %s", res.StatusCode, cuerpo)
		}
		var cuenta cuentaRespuesta
		if err := json.Unmarshal(cuerpo, &cuenta); err != nil {
			t.Fatal(err)
		}
		if cuenta.ID != dos.ID.String() || !cuenta.Activo {
			t.Fatalf("cuenta no activada: %+v", cuenta)
		}
	})

	t.Run("desactivar cuenta activa", func(t *testing.T) {
		csrfAcceso := csrfCookiePGConAcceso(t, origen, actor)
		ruta := "/api/v1/admin/cuentas/" + uuidString(uno.ID) + "/desactivar"
		res, cuerpo := doCuentas(t, origen, "POST", ruta, actor, csrfAcceso)
		if res.StatusCode != http.StatusOK {
			t.Fatalf("desactivar: %d %s", res.StatusCode, cuerpo)
		}
		var cuenta cuentaRespuesta
		if err := json.Unmarshal(cuerpo, &cuenta); err != nil {
			t.Fatal(err)
		}
		if cuenta.ID != uno.ID.String() || cuenta.Activo {
			t.Fatalf("cuenta no desactivada: %+v", cuenta)
		}
	})

	t.Run("cuenta inexistente devuelve 404", func(t *testing.T) {
		csrfAcceso := csrfCookiePGConAcceso(t, origen, actor)
		ruta := "/api/v1/admin/cuentas/" + uuidString(idInexistente(t)) + "/activar"
		res, _ := doCuentas(t, origen, "POST", ruta, actor, csrfAcceso)
		if res.StatusCode != http.StatusNotFound {
			t.Fatalf("inexistente: %d", res.StatusCode)
		}
	})

	t.Run("sin JWT devuelve 401", func(t *testing.T) {
		res, _ := doCuentas(t, origen, "GET", "/api/v1/admin/cuentas", "", "")
		if res.StatusCode != http.StatusUnauthorized {
			t.Fatalf("sin JWT: %d", res.StatusCode)
		}
	})

	t.Run("mutación sin CSRF devuelve 403", func(t *testing.T) {
		ruta := "/api/v1/admin/cuentas/" + uuidString(dos.ID) + "/desactivar"
		res, _ := doCuentas(t, origen, "POST", ruta, actor, "")
		if res.StatusCode != http.StatusForbidden {
			t.Fatalf("sin CSRF: %d", res.StatusCode)
		}
	})
}

func tokenValidoPG(t *testing.T, tokens *auth.Tokens, id pgtype.UUID) string {
	t.Helper()
	raw, err := tokens.Issue(id.String())
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
