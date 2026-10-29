package auth

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestBrowserExchange(t *testing.T) {
	tokens, _ := New(bytes.Repeat([]byte{8}, 32))
	now := time.Now().UTC().Truncate(time.Second)
	tokens.now = func() time.Time { return now }
	browser, err := NewBrowser(tokens, "https://portal.example.test")
	if err != nil {
		t.Fatal(err)
	}
	request := func(csrf, jwt string) *http.Request {
		r := httptest.NewRequest("POST", "https://portal.example.test/api/v1/fixture", nil)
		r.Header.Set("Origin", "https://portal.example.test")
		if csrf != "" {
			r.AddCookie(&http.Cookie{Name: CSRFCookie, Value: csrf})
			r.Header.Set(CSRFHeader, csrf)
		}
		if jwt != "" {
			r.AddCookie(&http.Cookie{Name: CookieName, Value: jwt})
		}
		return r
	}
	prepare := httptest.NewRecorder()
	if err := browser.PrepareCSRF(prepare, httptest.NewRequest("GET", "https://portal.example.test/api/v1/auth/csrf", nil)); err != nil {
		t.Fatal(err)
	}
	prelogin := prepare.Result().Cookies()[0]
	if prelogin.HttpOnly || !prelogin.Secure || prelogin.Path != "/" || prelogin.SameSite != http.SameSiteStrictMode || prelogin.Domain != "" {
		t.Fatal("cookie CSRF")
	}
	if err := browser.Check(request(prelogin.Value, "")); err != nil {
		t.Fatal("prelogin rechazado")
	}
	raw, _ := tokens.Issue("b14f0325-1640-4014-a00e-15b249b78866")
	login := httptest.NewRecorder()
	if err := browser.SetSession(login, raw); err != nil {
		t.Fatal(err)
	}
	cookies := login.Result().Cookies()
	access, csrf := cookies[0], cookies[1]
	if !access.HttpOnly || !access.Secure || access.Path != "/api" || access.Domain != "" || access.MaxAge != 7200 || access.SameSite != http.SameSiteStrictMode {
		t.Fatal("cookie privada")
	}
	if err := browser.Check(request(csrf.Value, raw)); err != nil {
		t.Fatal("mutación autenticada")
	}
	if err := browser.Check(request(prelogin.Value, raw)); err == nil {
		t.Fatal("token de prelogin aceptado con JWT")
	}
	other, _ := tokens.Issue("613699dd-13dc-47f2-943d-3e099e1f313e")
	if err := browser.Check(request(csrf.Value, other)); err == nil {
		t.Fatal("CSRF reutilizado con otra identidad")
	}
	for _, bad := range []string{"origin", "missing_origin", "host", "header", "missing_cookie", "duplicate_cookie", "duplicate_header", "forged", "duplicate_access"} {
		t.Run(bad, func(t *testing.T) {
			r := request(csrf.Value, raw)
			switch bad {
			case "origin":
				r.Header.Set("Origin", "https://portal.example.test.attacker.test")
			case "missing_origin":
				r.Header.Del("Origin")
				r.Header.Set("X-Forwarded-Host", "portal.example.test")
			case "host":
				r.Host = "attacker.test"
			case "header":
				r.Header.Set(CSRFHeader, "alterado")
			case "missing_cookie":
				r.Header.Set("Cookie", CookieName+"="+raw)
			case "duplicate_cookie":
				r.AddCookie(csrf)
			case "duplicate_header":
				r.Header.Add(CSRFHeader, csrf.Value)
			case "forged":
				r = request("v1.1."+strings.Repeat("A", 43)+"."+strings.Repeat("A", 43), raw)
			case "duplicate_access":
				r.AddCookie(access)
			}
			if err := browser.Check(r); err == nil {
				t.Fatal("petición insegura aceptada")
			}
		})
	}
	// A current-looking nonce still requires the server MAC, even if cookie/header match.
	forged := strings.Split(csrf.Value, ".")
	forged[3] = strings.Repeat("A", 43)
	if err := browser.Check(request(strings.Join(forged, "."), raw)); err == nil {
		t.Fatal("firma falsa")
	}
	tokens.now = func() time.Time { return now.Add(Lifetime + ClockSkew + time.Second) }
	if err := browser.Check(request(prelogin.Value, "")); err == nil {
		t.Fatal("CSRF vencido")
	}
	if err := browser.Check(request(csrf.Value, raw)); err == nil {
		t.Fatal("JWT vencido conserva vinculación")
	}
	refresh := httptest.NewRecorder()
	get := httptest.NewRequest("GET", "https://portal.example.test/api/v1/auth/csrf", nil)
	get.AddCookie(access)
	if err := browser.PrepareCSRF(refresh, get); err != nil {
		t.Fatal(err)
	}
	if err := browser.Check(request(refresh.Result().Cookies()[0].Value, raw)); err != nil {
		t.Fatal("preparación para logout vencido")
	}
	clear := httptest.NewRecorder()
	browser.ClearSession(clear)
	for i, c := range clear.Result().Cookies() {
		if c.MaxAge != -1 || c.Path != cookies[i].Path || c.Secure != cookies[i].Secure || c.HttpOnly != cookies[i].HttpOnly || c.Domain != "" {
			t.Fatal("borrado incompatible")
		}
	}
	if clear.Body.Len() != 0 || login.Body.Len() != 0 || prepare.Body.Len() != 0 {
		t.Fatal("se expusieron tokens en cuerpo")
	}
}

func TestBrowserOriginPolicy(t *testing.T) {
	tokens, _ := New(bytes.Repeat([]byte{9}, 32))
	for _, origin := range []string{"", "http://portal.example.test", "http://localhost.attacker.test", "https://user:pass@portal.example.test", "https://portal.example.test/", "https://portal.example.test?x=y", "null"} {
		if _, err := NewBrowser(tokens, origin); err == nil {
			t.Fatal("origen inválido aceptado")
		}
	}
	for _, origin := range []string{"http://localhost:8080", "http://127.0.0.1:8080"} {
		b, err := NewBrowser(tokens, origin)
		if err != nil {
			t.Fatal(err)
		}
		w := httptest.NewRecorder()
		if err := b.PrepareCSRF(w, httptest.NewRequest("GET", origin+"/api/v1/auth/csrf", nil)); err != nil {
			t.Fatal(err)
		}
		if w.Result().Cookies()[0].Secure {
			t.Fatal("HTTP local")
		}
	}
}
