package config

import (
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	for _, test := range []struct {
		name, url, max string
		valid          bool
	}{
		{"valid", "postgres://hs:private@db:5432/horas_sociales?sslmode=disable", "5", true},
		{"missing URL", "", "5", false},
		{"wrong scheme", "https://hs:private@db/horas_sociales", "5", false},
		{"missing database", "postgres://hs:private@db/", "5", false},
		{"missing password", "postgres://hs@db/horas_sociales", "5", false},
		{"missing limit", "postgres://hs:private@db/horas_sociales", "", false},
		{"zero limit", "postgres://hs:private@db/horas_sociales", "0", false},
		{"large limit", "postgres://hs:private@db/horas_sociales", "51", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("DATABASE_URL", test.url)
			t.Setenv("DB_MAX_CONNS", test.max)
			t.Setenv("CHECK_CORE_CATALOG", "")
			cfg, err := Load()
			if (err == nil) != test.valid {
				t.Fatalf("valid=%v, error=%v", test.valid, err)
			}
			if err != nil && strings.Contains(err.Error(), "private") {
				t.Fatal("error expone secreto")
			}
			if test.valid && cfg.Max != 5 {
				t.Fatalf("max=%d", cfg.Max)
			}
		})
	}
}

func TestLoadCoreCatalog(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://hs:private@db/horas_sociales")
	t.Setenv("DB_MAX_CONNS", "5")
	for _, tc := range []struct {
		value   string
		enabled bool
		invalid bool
	}{
		{value: ""},
		{value: "false"},
		{value: "true", enabled: true},
		{value: "yes", invalid: true},
	} {
		t.Run("value="+tc.value, func(t *testing.T) {
			t.Setenv("CHECK_CORE_CATALOG", tc.value)
			cfg, err := Load()
			if (err != nil) != tc.invalid || cfg.CheckCoreCatalog != tc.enabled {
				t.Fatalf("enabled=%v, error=%v", cfg.CheckCoreCatalog, err)
			}
		})
	}
}
