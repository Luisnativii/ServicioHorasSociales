package httpserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"horas-sociales/api/internal/modules/admin"
	"horas-sociales/api/internal/platform/auth"
)

func cuentaConID(t *testing.T, id pgtype.UUID, activo bool) admin.Administrator {
	t.Helper()
	return admin.Administrator{ID: id, Correo: "admin@example.test", Activo: activo}
}

func tokenValido(t *testing.T, tokens *auth.Tokens, id pgtype.UUID) string {
	t.Helper()
	raw, err := tokens.Issue(id.String())
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func prepararCSRFConAcceso(t *testing.T, browser *auth.Browser, router http.Handler, acceso string) string {
	t.Helper()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "http://localhost/api/v1/auth/csrf", nil)
	req.Host = "localhost"
	req.Header.Set("Origin", "http://localhost")
	req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: acceso})
	router.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("csrf con acceso: %d %s", w.Code, w.Body.String())
	}
	for _, c := range w.Result().Cookies() {
		if c.Name == auth.CSRFCookie {
			return c.Value
		}
	}
	t.Fatal("cookie CSRF no preparada")
	return ""
}

func peticionCuentas(t *testing.T, router http.Handler, metodo, ruta string, acceso, csrf string) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(metodo, "http://localhost"+ruta, nil)
	req.Host = "localhost"
	req.Header.Set("Origin", "http://localhost")
	if acceso != "" {
		req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: acceso})
	}
	if csrf != "" {
		req.AddCookie(&http.Cookie{Name: auth.CSRFCookie, Value: csrf})
		req.Header.Set("X-CSRF-Token", csrf)
	}
	router.ServeHTTP(w, req)
	return w
}

func TestListarCuentasRequiereJWT(t *testing.T) {
	_, _, router := routerDePrueba(t, buscadorFalso{cuenta: cuentaDePrueba(true)}, nil)
	w := peticionCuentas(t, router, "GET", "/api/v1/admin/cuentas", "", "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("listar sin JWT: %d", w.Code)
	}
}

func TestListarCuentas(t *testing.T) {
	id1 := idDePrueba()
	id2 := pgtype.UUID{
		Bytes: [16]byte{0xc2, 0x4f, 0x03, 0x25, 0x16, 0x40, 0x40, 0x14, 0xa0, 0x0e, 0x15, 0xb2, 0x49, 0xb7, 0x88, 0x67},
		Valid: true,
	}
	buscar := buscadorFalso{cuenta: cuentaConID(t, id1, true)}
	buscar.cuenta = cuentaConID(t, id1, true)
	// buscadorFalso.List ignora el correo y devuelve la cuenta guardada; usar dos cuentas no es posible con este falso.
	_ = id2
	tokens, _, router := routerDePrueba(t, buscar, nil)
	acceso := tokenValido(t, tokens, id1)

	w := peticionCuentas(t, router, "GET", "/api/v1/admin/cuentas", acceso, "")
	if w.Code != http.StatusOK {
		t.Fatalf("listar: %d %s", w.Code, w.Body.String())
	}
	if w.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("falta no-store")
	}
	var respuesta listarCuentasRespuesta
	if err := parseJSON(w.Body.Bytes(), &respuesta); err != nil {
		t.Fatal(err)
	}
	if len(respuesta.Cuentas) != 1 || respuesta.Cuentas[0].ID != id1.String() || respuesta.Cuentas[0].Correo != "admin@example.test" || !respuesta.Cuentas[0].Activo {
		t.Fatalf("cuentas inesperadas: %+v", respuesta.Cuentas)
	}
	if respuesta.Pagina.Pagina != 1 || respuesta.Pagina.PageSize != 25 || respuesta.Pagina.Total != 1 {
		t.Fatalf("paginación inesperada: %+v", respuesta.Pagina)
	}
}

func TestListarCuentasPaginacionInvalida(t *testing.T) {
	tokens, _, router := routerDePrueba(t, buscadorFalso{cuenta: cuentaDePrueba(true)}, nil)
	acceso := tokenValido(t, tokens, idDePrueba())
	w := peticionCuentas(t, router, "GET", "/api/v1/admin/cuentas?page=0", acceso, "")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("página inválida: %d", w.Code)
	}
}

func TestActivarCuenta(t *testing.T) {
	id := idDePrueba()
	buscar := buscadorFalso{cuenta: cuentaConID(t, id, false)}
	tokens, browser, router := routerDePrueba(t, buscar, nil)
	acceso := tokenValido(t, tokens, id)
	csrf := prepararCSRFConAcceso(t, browser, router, acceso)

	w := peticionCuentas(t, router, "POST", "/api/v1/admin/cuentas/"+id.String()+"/activar", acceso, csrf)
	if w.Code != http.StatusOK {
		t.Fatalf("activar: %d %s", w.Code, w.Body.String())
	}
	var cuenta cuentaRespuesta
	if err := parseJSON(w.Body.Bytes(), &cuenta); err != nil {
		t.Fatal(err)
	}
	if cuenta.ID != id.String() || !cuenta.Activo || cuenta.Correo != "admin@example.test" {
		t.Fatalf("cuenta activada inesperada: %+v", cuenta)
	}
}

func TestDesactivarCuenta(t *testing.T) {
	id := idDePrueba()
	buscar := buscadorFalso{cuenta: cuentaConID(t, id, true)}
	tokens, browser, router := routerDePrueba(t, buscar, nil)
	acceso := tokenValido(t, tokens, id)
	csrf := prepararCSRFConAcceso(t, browser, router, acceso)

	w := peticionCuentas(t, router, "POST", "/api/v1/admin/cuentas/"+id.String()+"/desactivar", acceso, csrf)
	if w.Code != http.StatusOK {
		t.Fatalf("desactivar: %d %s", w.Code, w.Body.String())
	}
	var cuenta cuentaRespuesta
	if err := parseJSON(w.Body.Bytes(), &cuenta); err != nil {
		t.Fatal(err)
	}
	if cuenta.ID != id.String() || cuenta.Activo {
		t.Fatalf("cuenta desactivada inesperada: %+v", cuenta)
	}
}

func TestActivarCuentaSinCSRF(t *testing.T) {
	id := idDePrueba()
	tokens, _, router := routerDePrueba(t, buscadorFalso{cuenta: cuentaConID(t, id, true)}, nil)
	acceso := tokenValido(t, tokens, id)
	w := peticionCuentas(t, router, "POST", "/api/v1/admin/cuentas/"+id.String()+"/activar", acceso, "")
	if w.Code != http.StatusForbidden {
		t.Fatalf("activar sin CSRF: %d", w.Code)
	}
}

func TestActivarCuentaInexistente(t *testing.T) {
	id := idDePrueba()
	buscar := buscadorFalso{err: admin.ErrNotFound}
	tokens, browser, router := routerDePrueba(t, buscar, nil)
	acceso := tokenValido(t, tokens, id)
	csrf := prepararCSRFConAcceso(t, browser, router, acceso)
	w := peticionCuentas(t, router, "POST", "/api/v1/admin/cuentas/"+id.String()+"/desactivar", acceso, csrf)
	if w.Code != http.StatusNotFound {
		t.Fatalf("cuenta inexistente: %d", w.Code)
	}
}

func TestActivarCuentaIDInvalido(t *testing.T) {
	tokens, browser, router := routerDePrueba(t, buscadorFalso{cuenta: cuentaDePrueba(true)}, nil)
	acceso := tokenValido(t, tokens, idDePrueba())
	csrf := prepararCSRFConAcceso(t, browser, router, acceso)
	w := peticionCuentas(t, router, "POST", "/api/v1/admin/cuentas/no-es-un-uuid/activar", acceso, csrf)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("ID inválido: %d", w.Code)
	}
}

func parseJSON(cuerpo []byte, dst any) error {
	return json.Unmarshal(cuerpo, dst)
}
