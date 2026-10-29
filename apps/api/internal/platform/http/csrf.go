package httpserver

import (
	"horas-sociales/api/internal/platform/auth"
	"net/http"
)

func RequireCSRF(browser *auth.Browser) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet && r.Method != http.MethodHead && r.Method != http.MethodOptions {
				if browser == nil || browser.Check(r) != nil {
					WriteError(w, r, "forbidden")
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
