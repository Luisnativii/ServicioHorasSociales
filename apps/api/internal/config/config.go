package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
)

type Config struct {
	URL string
	Max int32
}

func Load() (Config, error) {
	raw := os.Getenv("DATABASE_URL")
	u, err := url.Parse(raw)
	if err != nil || u == nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") || u.Hostname() == "" || u.User == nil || u.User.Username() == "" || u.Path == "" || u.Path == "/" {
		return Config{}, fmt.Errorf("DATABASE_URL debe indicar servidor, usuario y base PostgreSQL")
	}
	password, ok := u.User.Password()
	if !ok || password == "" {
		return Config{}, fmt.Errorf("DATABASE_URL requiere contraseña")
	}
	n, err := strconv.Atoi(os.Getenv("DB_MAX_CONNS"))
	if err != nil || n < 1 || n > 50 {
		return Config{}, fmt.Errorf("DB_MAX_CONNS debe estar entre 1 y 50")
	}
	return Config{URL: raw, Max: int32(n)}, nil
}
