package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type probe struct {
	failed   bool
	called   bool
	deadline time.Time
}

func (p *probe) Ping(ctx context.Context) error {
	p.called = true
	p.deadline, _ = ctx.Deadline()
	if p.failed {
		return errors.New("postgres://secret:password@db")
	}
	return nil
}

func TestHealth(t *testing.T) {
	for _, test := range []struct {
		method, path string
		failed       bool
		code         int
		status       string
	}{
		{"GET", "/api/health/live", true, 200, "ok"},
		{"GET", "/api/health/ready", true, 503, "unavailable"},
		{"GET", "/api/health/ready", false, 200, "ok"},
		{"POST", "/api/health/live", false, 405, "method_not_allowed"},
		{"GET", "/api/missing", false, 404, "not_found"},
	} {
		t.Run(test.method+test.path+test.status, func(t *testing.T) {
			db := &probe{failed: test.failed}
			response := httptest.NewRecorder()
			Router(db).ServeHTTP(response, httptest.NewRequest(test.method, test.path, nil))
			if response.Code != test.code {
				t.Fatalf("status=%d", response.Code)
			}
			var body map[string]string
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if body["status"] != test.status || len(body) != 1 {
				t.Fatalf("body=%v", body)
			}
			if !strings.HasPrefix(response.Header().Get("Content-Type"), "application/json") || response.Header().Get("Cache-Control") != "no-store" {
				t.Fatal("cabeceras inválidas")
			}
			if test.path == "/api/health/live" && db.called {
				t.Fatal("liveness no debe depender de PostgreSQL")
			}
			if test.path == "/api/health/ready" && (db.deadline.IsZero() || time.Until(db.deadline) > 2*time.Second) {
				t.Fatal("readiness requiere timeout")
			}
		})
	}
}

func TestReadinessPreservesCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	response := httptest.NewRecorder()
	Router(cancelProbe{t}).ServeHTTP(response, httptest.NewRequest("GET", "/api/health/ready", nil).WithContext(ctx))
	if response.Code != 503 {
		t.Fatalf("status=%d", response.Code)
	}
}

type cancelProbe struct{ t *testing.T }

func (p cancelProbe) Ping(ctx context.Context) error {
	if !errors.Is(ctx.Err(), context.Canceled) {
		p.t.Fatal("contexto no cancelado")
	}
	return ctx.Err()
}
