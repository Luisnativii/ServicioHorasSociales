package httpserver

import (
	"bytes"
	"github.com/go-chi/chi/v5"
	"horas-sociales/api/internal/platform/auth"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCSRFRouter(t *testing.T) {
	tokens, _ := auth.New(bytes.Repeat([]byte{11}, 32))
	browser, _ := auth.NewBrowser(tokens, "http://localhost:8080")
	called := false
	router := Router(&probe{}, Options{Tokens: tokens, Browser: browser, RegisterPrivate: func(r chi.Router) {
		r.Post("/fixture", func(w http.ResponseWriter, r *http.Request) { called = true; w.WriteHeader(204) })
	}})
	get := httptest.NewRecorder()
	router.ServeHTTP(get, httptest.NewRequest("GET", "http://localhost:8080/api/v1/auth/csrf", nil))
	if get.Code != 204 || get.Body.Len() != 0 || get.Header().Get("Cache-Control") != "no-store" || len(get.Result().Cookies()) != 1 {
		t.Fatal("prelogin CSRF")
	}
	raw, _ := tokens.Issue("b14f0325-1640-4014-a00e-15b249b78866")
	login := httptest.NewRecorder()
	browser.SetSession(login, raw)
	for _, proof := range []bool{false, true} {
		called = false
		r := httptest.NewRequest("POST", "http://localhost:8080/api/v1/fixture", nil)
		r.Header.Set("Origin", "http://localhost:8080")
		for _, c := range login.Result().Cookies() {
			r.AddCookie(c)
			if proof && c.Name == auth.CSRFCookie {
				r.Header.Set(auth.CSRFHeader, c.Value)
			}
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)
		if proof {
			if w.Code != 204 || !called {
				t.Fatal("prueba válida")
			}
		} else if w.Code != 403 || called {
			t.Fatal("mutación sin prueba")
		}
	}
	get = httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://localhost:8080/api/v1/auth/csrf", nil)
	r.Header.Set("Origin", "https://attacker.test")
	router.ServeHTTP(get, r)
	if get.Code != 403 || len(get.Result().Cookies()) != 0 {
		t.Fatal("origen exterior")
	}
	get = httptest.NewRecorder()
	Router(&probe{}).ServeHTTP(get, httptest.NewRequest("GET", "http://localhost:8080/api/v1/auth/csrf", nil))
	if get.Code != 503 {
		t.Fatal("sin configuración debe denegar")
	}
}
