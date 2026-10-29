package httpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"horas-sociales/api/internal/modules/admin"
	"horas-sociales/api/internal/platform/auth"
	"horas-sociales/api/internal/platform/password"
	"horas-sociales/api/internal/platform/ratelimit"
)

func loginRouterPG(t *testing.T, pool *pgxpool.Pool) (*auth.Tokens, *auth.Browser, string) {
	clave, err := os.CreateTemp("", "hs-login-test-key-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(clave.Name())
	if _, err := clave.Write(bytes.Repeat([]byte{73}, 32)); err != nil {
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
	buscar := admin.NewRepository(pool)
	// Límite alto en integración: el control de 429 ya se verifica en
	// TestLoginLimiteDeIntentos; aquí se evita que fallos intencionales de
	// una subprueba bloqueen a las siguientes (IP fija 127.0.0.1 en contenedor).
	limiter := ratelimit.NewLimiter(100, time.Minute, 64)
	origen := "http://" + ln.Addr().String()
	browser, err := auth.NewBrowser(tokens, origen)
	if err != nil {
		t.Fatal(err)
	}
	srv.Handler = Router(pool, Options{Tokens: tokens, Browser: browser, Buscador: buscar, Limiter: limiter})
	return tokens, browser, origen
}

func doForm(t *testing.T, origen, metodo, ruta, cuerpo, csrf string) (*http.Response, []byte) {
	req, err := http.NewRequest(metodo, origen+ruta, strings.NewReader(cuerpo))
	if err != nil {
		t.Fatal(err)
	}
	req.Host = strings.TrimPrefix(origen, "http://")
	req.Header.Set("Origin", origen)
	if cuerpo != "" {
		req.Header.Set("Content-Type", "application/json")
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

func cookieDeRespuestaHTTP(res *http.Response, nombre string) string {
	for _, c := range res.Cookies() {
		if c.Name == nombre {
			return c.Value
		}
	}
	return ""
}

func csrfCookiePG(t *testing.T, origen string) string {
	res, err := http.DefaultClient.Get(origen + "/api/v1/auth/csrf")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("csrf: %d", res.StatusCode)
	}
	for _, c := range res.Cookies() {
		if c.Name == auth.CSRFCookie {
			return c.Value
		}
	}
	t.Fatal("cookie CSRF no preparada")
	return ""
}

func postLoginPG(t *testing.T, origen string, csrf, cuerpo string) (*http.Response, []byte) {
	return doForm(t, origen, "POST", "/api/v1/auth/login", cuerpo, csrf)
}

func getMePG(t *testing.T, origen string, acceso string) (*http.Response, []byte) {
	req, err := http.NewRequest("GET", origen+"/api/v1/auth/me", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Host = strings.TrimPrefix(origen, "http://")
	req.Header.Set("Origin", origen)
	if acceso != "" {
		req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: acceso})
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

func TestLoginPostgreSQL(t *testing.T) {
	if os.Getenv("HS_TEST_LOGIN") != "1" || os.Getenv("DATABASE_URL") == "" {
		t.Skip("integración explícita: HS_TEST_LOGIN=1 y DATABASE_URL hacia hs-test")
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

	hashUno, err := password.Hash(ctx, []byte("contrasena1234"))
	if err != nil {
		t.Fatal(err)
	}
	uno, err := buscar.Create(ctx, " UNO@EJEMPLO.TEST ", hashUno)
	if err != nil || !uno.ID.Valid || !uno.Activo {
		t.Fatalf("alta cuenta uno: %v", err)
	}
	dos, err := buscar.Create(ctx, "dos@ejemplo.test", hashUno)
	if err != nil || !dos.ID.Valid {
		t.Fatalf("alta cuenta dos: %v", err)
	}
	if _, err := buscar.SetActive(ctx, dos.ID, false); err != nil {
		t.Fatalf("desactivar cuenta dos: %v", err)
	}
	defer func() {
		pool.Exec(ctx, `DELETE FROM horas_sociales.administrador WHERE correo IN ('uno@ejemplo.test','dos@ejemplo.test')`)
	}()

	tokens, _, origen := loginRouterPG(t, pool)
	csrf := csrfCookiePG(t, origen)

	t.Run("dos cuentas válidas inician sesión", func(t *testing.T) {
		res, cuerpo := postLoginPG(t, origen, csrf, `{"correo":"uno@ejemplo.test","contrasena":"contrasena1234"}`)
		if res.StatusCode != http.StatusOK {
			t.Fatalf("login uno: %d %s", res.StatusCode, cuerpo)
		}
		var identidad identidadRespuesta
		if err := json.Unmarshal(cuerpo, &identidad); err != nil || identidad.ID != uno.ID.String() || identidad.Correo != "uno@ejemplo.test" {
			t.Fatalf("identidad uno: %s", cuerpo)
		}
		acceso := cookieDeRespuestaHTTP(res, auth.CookieName)
		if acceso == "" {
			t.Fatal("hs_access no emitida")
		}
		if _, err := tokens.Verify(acceso); err != nil {
			t.Fatal("hs_access no es un JWT válido")
		}
		if cookieDeRespuestaHTTP(res, auth.CSRFCookie) == "" {
			t.Fatal("hs_csrf rotada no emitida")
		}

		res, cuerpo = postLoginPG(t, origen, csrf, `{"correo":"dos@ejemplo.test","contrasena":"contrasena1234"}`)
		if res.StatusCode != http.StatusUnauthorized {
			t.Fatalf("dos inactiva: %d %s", res.StatusCode, cuerpo)
		}
	})

	t.Run("contraseña incorrecta e inexistente no entregan sesión", func(t *testing.T) {
		res, _ := postLoginPG(t, origen, csrf, `{"correo":"uno@ejemplo.test","contrasena":"incorrecta123"}`)
		if res.StatusCode != http.StatusUnauthorized {
			t.Fatalf("incorrecta: %d", res.StatusCode)
		}
		res, _ = postLoginPG(t, origen, csrf, `{"correo":"nadie@ejemplo.test","contrasena":"contrasena1234"}`)
		if res.StatusCode != http.StatusUnauthorized {
			t.Fatalf("inexistente: %d", res.StatusCode)
		}
		res, _ = getMePG(t, origen, "")
		if res.StatusCode != http.StatusUnauthorized {
			t.Fatalf("me sin sesión: %d", res.StatusCode)
		}
		for _, c := range res.Cookies() {
			if c.Name == auth.CookieName {
				t.Fatal("fallo no emite hs_access")
			}
		}
	})

	t.Run("me identifica al actor con sesión válida", func(t *testing.T) {
		res, _ := postLoginPG(t, origen, csrf, `{"correo":"uno@ejemplo.test","contrasena":"contrasena1234"}`)
		if res.StatusCode != http.StatusOK {
			t.Fatalf("re-login uno: %d", res.StatusCode)
		}
		acceso := cookieDeRespuestaHTTP(res, auth.CookieName)

		res, cuerpo := getMePG(t, origen, acceso)
		if res.StatusCode != http.StatusOK {
			t.Fatalf("me con sesión: %d %s", res.StatusCode, cuerpo)
		}
		var identidad identidadRespuesta
		if err := json.Unmarshal(cuerpo, &identidad); err != nil || identidad.ID != uno.ID.String() || identidad.Correo != "uno@ejemplo.test" {
			t.Fatalf("me identidad: %s", cuerpo)
		}

		res, _ = getMePG(t, origen, "no-es-un-jwt")
		if res.StatusCode != http.StatusUnauthorized {
			t.Fatalf("me token invalido: %d", res.StatusCode)
		}
	})
}
