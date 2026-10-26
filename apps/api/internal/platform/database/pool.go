package database

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func New(ctx context.Context, url string, max int32) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, err
	}
	cfg.MaxConns = max
	cfg.MinConns = 0
	cfg.ConnConfig.ConnectTimeout = 3 * time.Second
	cfg.MaxConnLifetime = 30 * time.Minute
	cfg.MaxConnIdleTime = 5 * time.Minute
	cfg.HealthCheckPeriod = 30 * time.Second
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	startup, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(startup); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}
