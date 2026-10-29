package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"horas-sociales/api/internal/config"
	"horas-sociales/api/internal/modules/admin"
	"horas-sociales/api/internal/platform/auth"
	"horas-sociales/api/internal/platform/database"
	httpserver "horas-sociales/api/internal/platform/http"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		client := http.Client{Timeout: 3 * time.Second}
		response, err := client.Get("http://127.0.0.1:8080/api/health/live")
		if err != nil {
			os.Exit(1)
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusOK {
			os.Exit(1)
		}
		return
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	if err := run(); err != nil {
		// Solo errores controlados; no imprimir errores pgx ni cadenas de conexión.
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func run() error {
	var tokens *auth.Tokens
	var browser *auth.Browser
	if path := os.Getenv("JWT_KEY_FILE"); path != "" {
		var err error
		tokens, err = auth.LoadKey(path)
		if err != nil {
			return err
		}
		browser, err = auth.NewBrowser(tokens, os.Getenv("APP_ORIGIN"))
		if err != nil {
			return err
		}
	}
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("configuración inválida: %s", err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	pool, err := database.New(ctx, cfg.URL, cfg.Max)
	if err != nil {
		return errors.New("no se pudo conectar con PostgreSQL al iniciar")
	}
	defer pool.Close()
	srv := &http.Server{
		Addr: ":8080", Handler: httpserver.Router(pool, httpserver.Options{
			Tokens:   tokens,
			Browser:  browser,
			Buscador: admin.NewRepository(pool),
		}),
		ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second,
		WriteTimeout: 10 * time.Second, IdleTimeout: 60 * time.Second,
	}
	failed := make(chan error, 1)
	go func() { failed <- srv.ListenAndServe() }()
	slog.Info("API iniciada", "port", 8080, "db_max_conns", cfg.Max)
	select {
	case err := <-failed:
		if !errors.Is(err, http.ErrServerClosed) {
			return errors.New("fallo del servidor HTTP")
		}
	case <-ctx.Done():
	}
	shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdown); err != nil {
		_ = srv.Close()
		return errors.New("el servidor HTTP no terminó dentro del plazo de apagado")
	}
	slog.Info("API detenida ordenadamente")
	return nil
}
