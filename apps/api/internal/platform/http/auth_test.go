package httpserver

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"horas-sociales/api/internal/platform/auth"
)

func TestProtectedRouter(t *testing.T) {
	tokens, _ := auth.New(bytes.Repeat([]byte{19}, 32))
	id := "b14f0325-1640-4014-a00e-15b249b78866"
	raw, _ := tokens.Issue(id)
	called := false
	router := Router(&probe{failed: true}, Options{Tokens: tokens, RegisterPrivate: func(r chi.Router) {
		r.Get("/fixture", func(w http.ResponseWriter, r *http.Request) {
			identity, ok := auth.FromContext(r.Context())
			if !ok || identity.ID != id {
				t.Fatal("identidad ausente")
			}
			called = true
			w.WriteHeader(204)
		})
	}})
	for _, value := range []string{"", "alterado", raw} {
		called = false
		req := httptest.NewRequest("GET", "/api/v1/fixture", nil)
		if value != "" {
			req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: value})
		}
		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)
		if value == raw {
			if res.Code != 204 || !called {
				t.Fatal("token válido")
			}
		} else if res.Code != 401 || called {
			t.Fatal("acceso indebido")
		}
	}
	req := httptest.NewRequest("GET", "/api/v1/fixture", nil)
	req.Header.Set("Authorization", "Bearer "+raw)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	if res.Code != 401 {
		t.Fatal("se aceptó bearer")
	}
	req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: raw})
	req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: raw})
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	if res.Code != 401 {
		t.Fatal("cookies ambiguas")
	}
	res = httptest.NewRecorder()
	Router(&probe{}).ServeHTTP(res, httptest.NewRequest("GET", "/api/v1/fixture", nil))
	if res.Code != 401 {
		t.Fatal("denegación por defecto")
	}
}
