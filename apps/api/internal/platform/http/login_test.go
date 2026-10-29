package httpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"horas-sociales/api/internal/modules/admin"
	"horas-sociales/api/internal/platform/auth"
	"horas-sociales/api/internal/platform/password"
	"horas-sociales/api/internal/platform/ratelimit"
)

type buscadorFalso struct {
	cuenta admin.Administrator
	err    error
}

func (f buscadorFalso) FindByEmail(context.Context, string) (admin.Administrator, error) {
	return f.cuenta, f.err
}

func (f buscadorFalso) FindByID(context.Context, pgtype.UUID) (admin.Administrator, error) {
	return f.cuenta, f.err
}

func (f buscadorFalso) List(context.Context, int32, int64) ([]admin.Administrator, error) {
	if f.err != nil {
		return nil, f.err
	}
	return []admin.Administrator{f.cuenta}, nil
}

func (f buscadorFalso) Count(context.Context) (int32, error) {
	if f.err != nil {
		return 0, f.err
	}
	return 1, nil
}

func (f buscadorFalso) SetActive(_ context.Context, id pgtype.UUID, activar bool) (admin.Administrator, error) {
	if f.err != nil {
		return admin.Administrator{}, f.err
	}
	if f.cuenta.ID != id {
		return admin.Administrator{}, admin.ErrNotFound
	}
	f.cuenta.Activo = activar
	return f.cuenta, nil
}

func idDePrueba() pgtype.UUID {
	return pgtype.UUID{
		Bytes: [16]byte{0xb1, 0x4f, 0x03, 0x25, 0x16, 0x40, 0x40, 0x14, 0xa0, 0x0e, 0x15, 0xb2, 0x49, 0xb7, 0x88, 0x66},
		Valid: true,
	}
}

func cuentaDePrueba(activo bool) admin.Administrator {
	return admin.Administrator{ID: idDePrueba(), Correo: "admin@example.test", Activo: activo}
}

func routerDePrueba(t *testing.T, buscar administradorBuscador, limiter *ratelimit.Limiter) (*auth.Tokens, *auth.Browser, http.Handler) {
	tokens, err := auth.New(bytes.Repeat([]byte{19}, 32))
	if err != nil {
		t.Fatal(err)
	}
	browser, err := auth.NewBrowser(tokens, "http://localhost")
	if err != nil {
		t.Fatal(err)
	}
	if limiter == nil {
		limiter = ratelimit.NewLimiter(5, time.Minute, 64)
	}
	return tokens, browser, Router(&probe{}, Options{Tokens: tokens, Browser: browser, Buscador: buscar, Limiter: limiter})
}

func prepararCSRF(t *testing.T, browser *auth.Browser, router http.Handler) string {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "http://localhost/api/v1/auth/csrf", nil)
	req.Host = "localhost"
	req.Header.Set("Origin", "http://localhost")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("csrf inesperado: %d %s", w.Code, w.Body.String())
	}
	for _, c := range w.Result().Cookies() {
		if c.Name == auth.CSRFCookie {
			return c.Value
		}
	}
	t.Fatal("cookie CSRF no preparada")
	return ""
}

func inicioDeSesion(t *testing.T, router http.Handler, csrf, cuerpo string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "http://localhost/api/v1/auth/login", strings.NewReader(cuerpo))
	req.Host = "localhost"
	req.Header.Set("Origin", "http://localhost")
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.CSRFCookie, Value: csrf})
	req.Header.Set("X-CSRF-Token", csrf)
	router.ServeHTTP(w, req)
	return w
}

func cookieDeRespuestaRecorder(w *httptest.ResponseRecorder, nombre string) string {
	for _, c := range w.Result().Cookies() {
		if c.Name == nombre {
			return c.Value
		}
	}
	return ""
}

func TestLoginExitosoEntregaCookiesYIdentidad(t *testing.T) {
	hash, err := password.Hash(context.Background(), []byte("contrasena1234"))
	if err != nil {
		t.Fatal(err)
	}
	cuentaValida := cuentaDePrueba(true)
	cuentaValida.ContrasenaHash = hash
	tokens, browser, router := routerDePrueba(t, buscadorFalso{cuenta: cuentaValida}, nil)
	csrf := prepararCSRF(t, browser, router)

	w := inicioDeSesion(t, router, csrf, `{"correo":"ADMIN@EXAMPLE.TEST ","contrasena":"contrasena1234"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("login: %d %s", w.Code, w.Body.String())
	}
	var identidad identidadRespuesta
	if err := json.Unmarshal(w.Body.Bytes(), &identidad); err != nil || identidad.ID != idDePrueba().String() || identidad.Correo != "admin@example.test" {
		t.Fatalf("identidad inesperada: %s", w.Body.String())
	}
	if w.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("falta no-store")
	}
	acceso := cookieDeRespuestaRecorder(w, auth.CookieName)
	if acceso == "" {
		t.Fatal("no se emitió cookie hs_access")
	}
	csrfNuevo := cookieDeRespuestaRecorder(w, auth.CSRFCookie)
	if csrfNuevo == "" || csrfNuevo == csrf {
		t.Fatal("no se rotó la cookie CSRF")
	}
	if identidad, err := tokens.Verify(acceso); err != nil || identidad.ID != idDePrueba().String() {
		t.Fatal("cookie hs_access no es un JWT válido para la cuenta")
	}
	for _, c := range w.Result().Cookies() {
		if c.Name == auth.CookieName && (!c.HttpOnly || c.Path != "/api") {
			t.Fatal("hs_access debe ser HttpOnly con Path=/api")
		}
	}
}

func TestLoginRechazaSinRevelarCuenta(t *testing.T) {
	hash, err := password.Hash(context.Background(), []byte("contrasena1234"))
	if err != nil {
		t.Fatal(err)
	}
	correoValido := "admin@example.test"

	casos := []struct {
		nombre     string
		buscar     buscadorFalso
		correo     string
		contrasena string
	}{
		{"contraseña incorrecta", buscadorFalso{cuenta: admin.Administrator{ID: idDePrueba(), Correo: correoValido, ContrasenaHash: hash, Activo: true}}, correoValido, "otra1234"},
		{"cuenta inexistente", buscadorFalso{err: admin.ErrNotFound}, "desconocido@example.test", "contrasena1234"},
		{"cuenta inactiva", buscadorFalso{cuenta: cuentaDePrueba(false)}, correoValido, "contrasena1234"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			_, browser, router := routerDePrueba(t, c.buscar, nil)
			csrf := prepararCSRF(t, browser, router)
			w := inicioDeSesion(t, router, csrf, `{"correo":"`+c.correo+`","contrasena":"`+c.contrasena+`"}`)
			if w.Code != http.StatusUnauthorized {
				t.Fatalf("%s: %d %s", c.nombre, w.Code, w.Body.String())
			}
			var errorRes struct {
				Error struct {
					Code string `json:"code"`
				} `json:"error"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &errorRes); err != nil || errorRes.Error.Code != "unauthorized" {
				t.Fatalf("error no genérico: %s", w.Body.String())
			}
			if cookieDeRespuestaRecorder(w, auth.CookieName) != "" {
				t.Fatal("fallo no debe emitir hs_access")
			}
		})
	}
}

func TestLoginEntradaInvalida(t *testing.T) {
	_, browser, router := routerDePrueba(t, buscadorFalso{cuenta: cuentaDePrueba(true)}, nil)
	csrf := prepararCSRF(t, browser, router)

	for _, c := range []struct {
		nombre   string
		cuerpo   string
		content  string
		esperado int
		codigo   string
	}{
		{"correo sin formato", `{"correo":"no-es-un-correo","contrasena":"contrasena1234"}`, "application/json", http.StatusBadRequest, "bad_request"},
		{"campo desconocido", `{"correo":"admin@example.test","contrasena":"contrasena1234","extra":1}`, "application/json", http.StatusBadRequest, "invalid_json"},
		{"sin tipo de contenido", `{"correo":"admin@example.test","contrasena":"contrasena1234"}`, "text/plain", http.StatusUnsupportedMediaType, "unsupported_media_type"},
		{"contraseña vacía", `{"correo":"admin@example.test","contrasena":""}`, "application/json", http.StatusBadRequest, "bad_request"},
		{"sin campos", `{}`, "application/json", http.StatusBadRequest, "bad_request"},
	} {
		t.Run(c.nombre, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "http://localhost/api/v1/auth/login", strings.NewReader(c.cuerpo))
			req.Host = "localhost"
			req.Header.Set("Origin", "http://localhost")
			req.Header.Set("Content-Type", c.content)
			req.AddCookie(&http.Cookie{Name: auth.CSRFCookie, Value: csrf})
			req.Header.Set("X-CSRF-Token", csrf)
			router.ServeHTTP(w, req)
			if w.Code != c.esperado {
				t.Fatalf("%s: %d %s", c.nombre, w.Code, w.Body.String())
			}
			if c.codigo != "" {
				var errorRes struct {
					Error struct {
						Code string `json:"code"`
					} `json:"error"`
				}
				if err := json.Unmarshal(w.Body.Bytes(), &errorRes); err != nil || errorRes.Error.Code != c.codigo {
					t.Fatalf("código inesperado: %s", w.Body.String())
				}
			}
		})
	}
}

func TestLoginRechazaSinCSRF(t *testing.T) {
	_, _, router := routerDePrueba(t, buscadorFalso{cuenta: cuentaDePrueba(true)}, nil)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "http://localhost/api/v1/auth/login", strings.NewReader(`{"correo":"admin@example.test","contrasena":"contrasena1234"}`))
	req.Host = "localhost"
	req.Header.Set("Origin", "http://localhost")
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("sin CSRF: %d %s", w.Code, w.Body.String())
	}
}

func TestLoginLimiteDeIntentos(t *testing.T) {
	_, browser, _ := routerDePrueba(t, buscadorFalso{cuenta: cuentaDePrueba(true)}, nil)
	limiter := ratelimit.NewLimiter(2, time.Minute, 16)
	tokens, _, router := routerDePrueba(t, buscadorFalso{cuenta: cuentaDePrueba(true)}, limiter)
	csrf := prepararCSRF(t, browser, router)

	for i := 0; i < 2; i++ {
		w := inicioDeSesion(t, router, csrf, `{"correo":"admin@example.test","contrasena":"incorrecta123"}`)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("intento %d: %d", i, w.Code)
		}
	}
	w := inicioDeSesion(t, router, csrf, `{"correo":"admin@example.test","contrasena":"incorrecta123"}`)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("límite: %d %s", w.Code, w.Body.String())
	}
	if w.Header().Get("Retry-After") == "" {
		t.Fatal("429 debe indicar Retry-After")
	}
	var errorRes struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &errorRes); err != nil || errorRes.Error.Code != "too_many_requests" {
		t.Fatalf("código esperado too_many_requests: %s", w.Body.String())
	}
	_ = tokens
}

func TestMeSinSesionResponde401(t *testing.T) {
	_, _, router := routerDePrueba(t, buscadorFalso{cuenta: cuentaDePrueba(true)}, nil)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "http://localhost/api/v1/auth/me", nil)
	req.Host = "localhost"
	req.Header.Set("Origin", "http://localhost")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("me sin sesión: %d", w.Code)
	}
}

func TestMeConSesionIdentificaActor(t *testing.T) {
	tokens, _, router := routerDePrueba(t, buscadorFalso{cuenta: cuentaDePrueba(true)}, nil)
	raw, err := tokens.Issue(idDePrueba().String())
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "http://localhost/api/v1/auth/me", nil)
	req.Host = "localhost"
	req.Header.Set("Origin", "http://localhost")
	req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: raw})
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("me: %d %s", w.Code, w.Body.String())
	}
	var identidad identidadRespuesta
	if err := json.Unmarshal(w.Body.Bytes(), &identidad); err != nil || identidad.ID != idDePrueba().String() || identidad.Correo != "admin@example.test" {
		t.Fatalf("identidad inesperada: %s", w.Body.String())
	}
	if w.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("falta no-store")
	}
}

func TestMeRechazaTokenInvalidoYCookiesAmbiguas(t *testing.T) {
	_, _, router := routerDePrueba(t, buscadorFalso{cuenta: cuentaDePrueba(true)}, nil)
	for _, c := range []struct {
		nombre  string
		cookies []*http.Cookie
	}{
		{"token alterado", []*http.Cookie{{Name: auth.CookieName, Value: "no-es-un-jwt"}}},
		{"cookies duplicadas", []*http.Cookie{{Name: auth.CookieName, Value: "uno"}, {Name: auth.CookieName, Value: "dos"}}},
	} {
		t.Run(c.nombre, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "http://localhost/api/v1/auth/me", nil)
			req.Host = "localhost"
			req.Header.Set("Origin", "http://localhost")
			for _, cookie := range c.cookies {
				req.AddCookie(cookie)
			}
			router.ServeHTTP(w, req)
			if w.Code != http.StatusUnauthorized {
				t.Fatalf("%s: %d", c.nombre, w.Code)
			}
		})
	}
}
