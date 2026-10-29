package httpserver

import (
	"net/http"

	"horas-sociales/api/internal/platform/auth"
)

// logoutHandler implementa POST /api/v1/auth/logout. Exige CSRF y origen
// permitido, pero no un JWT vigente: permite cerrar incluso cuando el token
// ya venció, siempre que se haya preparado CSRF con antelación. Borra ambas
// cookies con los mismos atributos con que se emitieron; no invalida nada en
// servidor (no hay tabla de sesiones ni revocación).
func logoutHandler(browser *auth.Browser) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = r
		browser.ClearSession(w)
		w.WriteHeader(http.StatusNoContent)
	}
}
