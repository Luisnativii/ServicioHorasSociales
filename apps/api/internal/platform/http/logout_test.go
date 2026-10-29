package httpserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"horas-sociales/api/internal/platform/auth"
	"horas-sociales/api/internal/platform/password"
)

func peticionLogout(t *testing.T, router http.Handler, csrf, origen string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", origen+"/api/v1/auth/logout", nil)
	req.Host = strings.TrimPrefix(origen, "http://")
	req.Header.Set("Origin", origen)
	req.AddCookie(&http.Cookie{Name: auth.CSRFCookie, Value: csrf})
	req.Header.Set("X-CSRF-Token", csrf)
	router.ServeHTTP(w, req)
	return w
}

func cookiesDeRespuesta(w *httptest.ResponseRecorder) map[string]string {
	result := map[string]string{}
	for _, linea := range w.Result().Header["Set-Cookie"] {
		partes := strings.SplitN(linea, ";", 2)
		nombreValor := strings.TrimSpace(partes[0])
		if i := strings.Index(nombreValor, "="); i >= 0 {
			result[strings.TrimSpace(nombreValor[:i])] = strings.TrimSpace(nombreValor[i+1:])
		}
	}
	return result
}

func atributosDeCookie(w *httptest.ResponseRecorder, nombre string) string {
	for _, linea := range w.Result().Header["Set-Cookie"] {
		if strings.HasPrefix(strings.ToLower(linea), strings.ToLower(nombre)+"=") {
			return linea
		}
	}
	return ""
}

func hayAtributo(atributos, clave string) bool {
	for _, parte := range strings.Split(atributos, ";") {
		if strings.EqualFold(strings.TrimSpace(parte), clave) {
			return true
		}
	}
	return false
}

func TestLogoutBorraAmbasCookiesConSesionValida(t *testing.T) {
	hash, err := password.Hash(context.Background(), []byte("contrasena1234"))
	if err != nil {
		t.Fatal(err)
	}
	cuentaValida := cuentaDePrueba(true)
	cuentaValida.ContrasenaHash = hash
	_, browser, router := routerDePrueba(t, buscadorFalso{cuenta: cuentaValida}, nil)

	csrf := prepararCSRF(t, browser, router)
	w := inicioDeSesion(t, router, csrf, `{"correo":"admin@example.test","contrasena":"contrasena1234"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("login previo: %d %s", w.Code, w.Body.String())
	}
	csrf = cookieDeRespuestaRecorder(w, auth.CSRFCookie)
	csrf = prepararCSRF(t, browser, router)

	resp := peticionLogout(t, router, csrf, "http://localhost")
	if resp.Code != http.StatusNoContent {
		t.Fatalf("logout: %d %s", resp.Code, resp.Body.String())
	}
	if resp.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("falta no-store")
	}

	cookies := cookiesDeRespuesta(resp)
	if _, existe := cookies["hs_access"]; !existe {
		t.Fatal("hs_access no se borró")
	}
	if _, existe := cookies["hs_csrf"]; !existe {
		t.Fatal("hs_csrf no se borró")
	}
	hsAccess := atributosDeCookie(resp, "hs_access")
	if !hayAtributo(hsAccess, "Path=/api") || !hayAtributo(hsAccess, "HttpOnly") {
		t.Fatalf("hs_access sin atributos esperados: %s", hsAccess)
	}
	if !hayAtributo(hsAccess, "Max-Age=0") && !hayAtributo(hsAccess, "Max-Age=-1") {
		t.Fatalf("hs_access sin expiración inmediata: %s", hsAccess)
	}
	if !strings.Contains(hsAccess, "Expires=") {
		t.Fatalf("hs_access sin fecha de expiración pasada: %s", hsAccess)
	}
	hsCsrf := atributosDeCookie(resp, "hs_csrf")
	if !hayAtributo(hsCsrf, "Path=/") || hayAtributo(hsCsrf, "HttpOnly") {
		t.Fatalf("hs_csrf con atributos incorrectos: %s", hsCsrf)
	}
	if !hayAtributo(hsCsrf, "Max-Age=0") && !hayAtributo(hsCsrf, "Max-Age=-1") {
		t.Fatalf("hs_csrf sin expiración inmediata: %s", hsCsrf)
	}
}

func TestLogoutSinJWTValido(t *testing.T) {
	_, browser, router := routerDePrueba(t, buscadorFalso{cuenta: cuentaDePrueba(true)}, nil)
	csrf := prepararCSRF(t, browser, router)

	resp := peticionLogout(t, router, csrf, "http://localhost")
	if resp.Code != http.StatusNoContent {
		t.Fatalf("logout sin JWT válido: %d %s", resp.Code, resp.Body.String())
	}
}

func TestLogoutRechazaSinCSRF(t *testing.T) {
	_, _, router := routerDePrueba(t, buscadorFalso{cuenta: cuentaDePrueba(true)}, nil)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "http://localhost/api/v1/auth/logout", nil)
	req.Host = "localhost"
	req.Header.Set("Origin", "http://localhost")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("sin CSRF: %d %s", w.Code, w.Body.String())
	}
}

func TestLogoutRechazaOrigenNoPermitido(t *testing.T) {
	_, browser, router := routerDePrueba(t, buscadorFalso{cuenta: cuentaDePrueba(true)}, nil)
	csrf := prepararCSRF(t, browser, router)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "http://localhost/api/v1/auth/logout", nil)
	req.Host = "localhost"
	req.Header.Set("Origin", "https://otro.ejemplo.test")
	req.AddCookie(&http.Cookie{Name: auth.CSRFCookie, Value: csrf})
	req.Header.Set("X-CSRF-Token", csrf)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("origen no permitido: %d %s", w.Code, w.Body.String())
	}
}

func TestLogoutRepetido(t *testing.T) {
	hash, err := password.Hash(context.Background(), []byte("contrasena1234"))
	if err != nil {
		t.Fatal(err)
	}
	cuentaValida := cuentaDePrueba(true)
	cuentaValida.ContrasenaHash = hash
	_, browser, router := routerDePrueba(t, buscadorFalso{cuenta: cuentaValida}, nil)

	csrf := prepararCSRF(t, browser, router)
	w := inicioDeSesion(t, router, csrf, `{"correo":"admin@example.test","contrasena":"contrasena1234"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("login previo: %d", w.Code)
	}
	csrf = cookieDeRespuestaRecorder(w, auth.CSRFCookie)

	// Primera vez con sesión válida.
	csrf = prepararCSRF(t, browser, router)
	resp := peticionLogout(t, router, csrf, "http://localhost")
	if resp.Code != http.StatusNoContent {
		t.Fatalf("primer logout: %d %s", resp.Code, resp.Body.String())
	}

	// Repetida con el mismo token CSRF: idempotente, 204 controlado (sin estado
	// de servidor el token CSRF sigue siendo válido hasta que se prepare otro).
	resp = peticionLogout(t, router, csrf, "http://localhost")
	if resp.Code != http.StatusNoContent {
		t.Fatalf("logout repetido con CSRF válido: %d %s", resp.Code, resp.Body.String())
	}

	// Tras preparar CSRF de nuevo, sigue funcionando.
	csrf = prepararCSRF(t, browser, router)
	resp = peticionLogout(t, router, csrf, "http://localhost")
	if resp.Code != http.StatusNoContent {
		t.Fatalf("logout tras nuevo CSRF: %d %s", resp.Code, resp.Body.String())
	}
}

func TestLogoutNoExponeDatosEnCuerpo(t *testing.T) {
	_, browser, router := routerDePrueba(t, buscadorFalso{cuenta: cuentaDePrueba(true)}, nil)
	csrf := prepararCSRF(t, browser, router)
	resp := peticionLogout(t, router, csrf, "http://localhost")
	if resp.Code != http.StatusNoContent {
		t.Fatalf("logout: %d", resp.Code)
	}
	if resp.Body.Len() != 0 {
		t.Fatal("logout no debe tener cuerpo")
	}
}
