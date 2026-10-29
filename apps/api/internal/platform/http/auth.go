package httpserver

import (
	"net/http"

	"horas-sociales/api/internal/platform/auth"
)

// RequireAdmin accepts only the browser cookie. Authorization headers, query
// strings and body fields cannot choose a different identity or bypass cookies.
func RequireAdmin(tokens *auth.Tokens) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var raw string
			count := 0
			for _, cookie := range r.Cookies() {
				if cookie.Name == auth.CookieName {
					raw = cookie.Value
					count++
				}
			}
			identity, err := tokens.Verify(raw)
			if count != 1 || err != nil {
				WriteError(w, r, "unauthorized")
				return
			}
			next.ServeHTTP(w, r.WithContext(auth.WithIdentity(r.Context(), identity)))
		})
	}
}
