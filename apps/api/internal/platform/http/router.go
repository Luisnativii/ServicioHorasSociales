package httpserver

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"horas-sociales/api/internal/platform/auth"
)

type Pinger interface{ Ping(context.Context) error }

func reply(w http.ResponseWriter, r *http.Request, status int, state string) {
	WriteJSON(w, r, status, map[string]string{"status": state})
}

func routeError(w http.ResponseWriter, r *http.Request, status int, code string) {
	if r.URL.Path == "/api/v1" || strings.HasPrefix(r.URL.Path, "/api/v1/") {
		WriteError(w, r, code)
		return
	}
	if status == http.StatusInternalServerError {
		code = "error"
	}
	reply(w, r, status, code)
}

type Options struct {
	Tokens *auth.Tokens
	// RegisterPrivate mounts future business handlers inside the protected group.
	RegisterPrivate func(chi.Router)
}

func Router(db Pinger, options ...Options) http.Handler {
	var config Options
	if len(options) > 0 {
		config = options[0]
	}
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Request-ID", middleware.GetReqID(r.Context()))
			next.ServeHTTP(w, r)
		})
	})
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, q *http.Request) {
			defer func() {
				if recover() != nil {
					slog.Error("error interno HTTP", "request_id", middleware.GetReqID(q.Context()))
					routeError(w, q, http.StatusInternalServerError, "internal_error")
				}
			}()
			next.ServeHTTP(w, q)
		})
	})
	r.Get("/api/health/live", func(w http.ResponseWriter, q *http.Request) {
		reply(w, q, http.StatusOK, "ok")
	})
	r.Get("/api/health/ready", func(w http.ResponseWriter, q *http.Request) {
		ctx, cancel := context.WithTimeout(q.Context(), 2*time.Second)
		defer cancel()
		if db.Ping(ctx) != nil {
			slog.Warn("PostgreSQL no disponible", "request_id", middleware.GetReqID(q.Context()))
			reply(w, q, http.StatusServiceUnavailable, "unavailable")
			return
		}
		reply(w, q, http.StatusOK, "ok")
	})
	r.Route("/api/v1", func(private chi.Router) {
		private.Use(RequireAdmin(config.Tokens))
		if config.RegisterPrivate != nil {
			config.RegisterPrivate(private)
		}
		// Chi's fallback handlers can run before route middleware when there is
		// no matching endpoint. Protect those explicitly as well.
		private.NotFound(RequireAdmin(config.Tokens)(http.HandlerFunc(func(w http.ResponseWriter, q *http.Request) { WriteError(w, q, "not_found") })).ServeHTTP)
		private.MethodNotAllowed(RequireAdmin(config.Tokens)(http.HandlerFunc(func(w http.ResponseWriter, q *http.Request) { WriteError(w, q, "method_not_allowed") })).ServeHTTP)
	})
	r.NotFound(func(w http.ResponseWriter, q *http.Request) { routeError(w, q, http.StatusNotFound, "not_found") })
	r.MethodNotAllowed(func(w http.ResponseWriter, q *http.Request) {
		routeError(w, q, http.StatusMethodNotAllowed, "method_not_allowed")
	})
	return r
}
