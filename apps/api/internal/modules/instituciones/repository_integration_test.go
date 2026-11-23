package instituciones

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestInstitucionesPostgreSQL(t *testing.T) {
	if os.Getenv("HS_TEST_CORE_CATALOG") != "1" {
		t.Skip("requiere PostgreSQL aislado mediante scripts/check-core-readiness.sh")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	config, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
	if err != nil || config.ConnConfig.Host != "db" || config.ConnConfig.User != "hs_core_probe" ||
		config.ConnConfig.Database != "horas_sociales_test" {
		t.Fatal("se exige hs-test-core con la cuenta de pruebas")
	}
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatal("no se pudo preparar PostgreSQL de pruebas")
	}
	defer pool.Close()
	repo := NewRepository(pool)

	uuid := func(raw string) pgtype.UUID {
		t.Helper()
		var id pgtype.UUID
		if err := id.Scan(raw); err != nil {
			t.Fatal(err)
		}
		return id
	}
	id := uuid("00000000-0000-4000-8000-000000046001")
	legacyID := uuid("00000000-0000-4000-8000-000000046002")
	missingID := uuid("00000000-0000-4000-8000-000000046099")
	actor := uuid("00000000-0000-4000-8000-000000000046")

	canonica, err := repo.FindByID(ctx, id)
	if err != nil || canonica.ID != id || !canonica.Activo || canonica.CodigoCSS == nil || *canonica.CodigoCSS != "00046" ||
		canonica.Sede == nil || !canonica.Sede.DistritoID.Valid || canonica.Sede.Nombre != "San Francisco Menéndez" ||
		canonica.UbicacionHeredada != nil {
		t.Fatalf("institución con sede inesperada: %+v, error=%v", canonica, err)
	}
	heredada, err := repo.FindByID(ctx, legacyID)
	if err != nil || heredada.ID != legacyID || heredada.CodigoCSS != nil || heredada.Sede != nil ||
		heredada.UbicacionHeredada == nil || *heredada.UbicacionHeredada != "Descripción histórica de sede" {
		t.Fatalf("institución heredada inesperada: %+v, error=%v", heredada, err)
	}
	if _, err := repo.FindByID(ctx, pgtype.UUID{}); !errors.Is(err, ErrInvalidID) {
		t.Fatalf("UUID inválido: %v", err)
	}
	if _, err := repo.Deactivate(ctx, id, pgtype.UUID{}); !errors.Is(err, ErrInvalidActor) {
		t.Fatalf("actor inválido: %v", err)
	}
	if _, err := repo.Deactivate(ctx, missingID, actor); !errors.Is(err, ErrNotFound) {
		t.Fatalf("UUID inexistente: %v", err)
	}
	for range 2 {
		inactiva, err := repo.Deactivate(ctx, id, actor)
		if err != nil || inactiva.ID != id || inactiva.Activo || inactiva.Sede == nil || *inactiva.Sede != *canonica.Sede {
			t.Fatalf("desactivación inesperada: %+v, error=%v", inactiva, err)
		}
	}
	inactiva, err := repo.FindByID(ctx, id)
	if err != nil || inactiva.Activo || inactiva.CodigoCSS == nil || *inactiva.CodigoCSS != "00046" {
		t.Fatalf("lectura histórica inesperada: %+v, error=%v", inactiva, err)
	}
}
