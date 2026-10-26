package httpserver

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Pinger interface{ Ping(context.Context) error }

func reply(w http.ResponseWriter, status int, state string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": state})
}

func Router(db Pinger) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, q *http.Request) {
			defer func() {
				if recover() != nil {
					slog.Error("error interno HTTP", "request_id", middleware.GetReqID(q.Context()))
					reply(w, http.StatusInternalServerError, "error")
				}
			}()
			next.ServeHTTP(w, q)
		})
	})
	r.Get("/api/health/live", func(w http.ResponseWriter, q *http.Request) {
		reply(w, http.StatusOK, "ok")
	})
	r.Get("/api/health/ready", func(w http.ResponseWriter, q *http.Request) {
		ctx, cancel := context.WithTimeout(q.Context(), 2*time.Second)
		defer cancel()
		if db.Ping(ctx) != nil {
			slog.Warn("PostgreSQL no disponible", "request_id", middleware.GetReqID(q.Context()))
			reply(w, http.StatusServiceUnavailable, "unavailable")
			return
		}
		reply(w, http.StatusOK, "ok")
	})
	r.NotFound(func(w http.ResponseWriter, q *http.Request) { reply(w, http.StatusNotFound, "not_found") })
	r.MethodNotAllowed(func(w http.ResponseWriter, q *http.Request) {
		reply(w, http.StatusMethodNotAllowed, "method_not_allowed")
	})
	return r
}
