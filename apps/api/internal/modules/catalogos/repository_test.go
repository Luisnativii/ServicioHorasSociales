package catalogos

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestVisibleName(t *testing.T) {
	got, err := VisibleName("  Apoyo Comunitario  ")
	if err != nil || got != "Apoyo Comunitario" {
		t.Fatalf("nombre=%q err=%v", got, err)
	}
	for _, input := range []string{"", "   ", "\t\n"} {
		if _, err := VisibleName(input); !errors.Is(err, ErrInvalidName) {
			t.Fatalf("se aceptó nombre vacío %q", input)
		}
	}
	if storageError(&pgconn.PgError{Message: "secret SQL"}) != ErrStorage {
		t.Fatal("error de almacenamiento no controlado")
	}
}

func TestRepositoryPostgreSQL(t *testing.T) {
	if os.Getenv("HS_TEST_CATALOGS") != "1" {
		t.Skip("integración explícita mediante scripts/check-catalogs.sh")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	config, err := pgx.ParseConfig(os.Getenv("DATABASE_URL"))
	if err != nil || config.Host != "db" || !strings.HasPrefix(config.Database, "hs_catalogs_probe_") || config.User != "hs_app" {
		t.Fatal("se exige la base efímera de catálogos con cuenta hs_app")
	}
	conn, err := pgx.ConnectConfig(ctx, config)
	if err != nil {
		t.Fatal("no se pudo conectar a hs-test")
	}
	defer conn.Close(context.Background())
	tx, err := conn.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(context.Background())
	repository := NewRepository(tx)

	tipo, err := repository.CreateTipo(ctx, "  Apoyo Comunitario  ")
	if err != nil || !tipo.ID.Valid || tipo.Nombre != "Apoyo Comunitario" || !tipo.Activo {
		t.Fatalf("alta de tipo: %#v err=%v", tipo, err)
	}
	area, err := repository.CreateArea(ctx, "Apoyo Comunitario")
	if err != nil || !area.ID.Valid || area.Nombre != tipo.Nombre || !area.Activo {
		t.Fatalf("catálogos independientes: %#v err=%v", area, err)
	}

	child, err := tx.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, duplicate := NewRepository(child).CreateTipo(ctx, "apoyo comunitario")
	_ = child.Rollback(ctx)
	if !errors.Is(duplicate, ErrDuplicate) {
		t.Fatalf("duplicado de tipo: %v", duplicate)
	}

	second, err := repository.CreateTipo(ctx, "Educación")
	if err != nil || second.ID == tipo.ID {
		t.Fatalf("segundo tipo: %#v err=%v", second, err)
	}
	child, err = tx.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, duplicate = NewRepository(child).RenameTipo(ctx, second.ID, " APOYO COMUNITARIO ")
	_ = child.Rollback(ctx)
	if !errors.Is(duplicate, ErrDuplicate) {
		t.Fatalf("renombrado duplicado: %v", duplicate)
	}
	second, err = repository.RenameTipo(ctx, second.ID, "Educación y Cultura")
	if err != nil || second.Nombre != "Educación y Cultura" {
		t.Fatalf("renombrado: %#v err=%v", second, err)
	}

	deactivated, err := repository.SetTipoActive(ctx, tipo.ID, false)
	if err != nil || deactivated.Activo || deactivated.Nombre != tipo.Nombre {
		t.Fatalf("desactivación: %#v err=%v", deactivated, err)
	}
	active, err := repository.ListTipos(ctx, false)
	if err != nil || len(active) != 1 || active[0].ID != second.ID {
		t.Fatalf("listado activo: %#v err=%v", active, err)
	}
	all, err := repository.ListTipos(ctx, true)
	if err != nil || len(all) != 2 {
		t.Fatalf("listado histórico: %#v err=%v", all, err)
	}
	reactivated, err := repository.SetTipoActive(ctx, tipo.ID, true)
	if err != nil || !reactivated.Activo || reactivated.Nombre != tipo.Nombre {
		t.Fatalf("reactivación: %#v err=%v", reactivated, err)
	}

	found, err := repository.FindAreaByID(ctx, area.ID)
	if err != nil || found != area {
		t.Fatalf("consulta de área: %#v err=%v", found, err)
	}
	area, err = repository.SetAreaActive(ctx, area.ID, false)
	if err != nil || area.Activo {
		t.Fatalf("desactivación de área: %#v err=%v", area, err)
	}
	areas, err := repository.ListAreas(ctx, false)
	if err != nil || len(areas) != 0 {
		t.Fatalf("áreas activas: %#v err=%v", areas, err)
	}
	areas, err = repository.ListAreas(ctx, true)
	if err != nil || len(areas) != 1 || areas[0].ID != area.ID {
		t.Fatalf("áreas históricas: %#v err=%v", areas, err)
	}

	invalid := pgtype.UUID{}
	if _, err := repository.FindTipoByID(ctx, invalid); !errors.Is(err, ErrInvalidID) {
		t.Fatalf("identificador inválido: %v", err)
	}
	missing := pgtype.UUID{Bytes: [16]byte{15: 1}, Valid: true}
	if _, err := repository.SetAreaActive(ctx, missing, false); !errors.Is(err, ErrNotFound) {
		t.Fatalf("referencia inexistente: %v", err)
	}
	if _, err := repository.CreateArea(ctx, " \n "); !errors.Is(err, ErrInvalidName) {
		t.Fatalf("nombre inválido: %v", err)
	}

	for _, statement := range []string{
		`INSERT INTO horas_sociales.tipo_proyecto(nombre) VALUES ('')`,
		`INSERT INTO horas_sociales.tipo_proyecto(nombre) VALUES (' exterior ')`,
		`INSERT INTO horas_sociales.area(nombre) VALUES (' ')`,
		`INSERT INTO horas_sociales.area(nombre) VALUES (E'\texterior')`,
		`INSERT INTO horas_sociales.area(nombre) VALUES (E'exterior\n')`,
	} {
		child, err := tx.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		_, rejected := child.Exec(ctx, statement)
		_ = child.Rollback(ctx)
		var postgres *pgconn.PgError
		if !errors.As(rejected, &postgres) || postgres.Code != "23514" {
			t.Fatalf("la base aceptó nombre fuera del contrato (%s): %v", statement, rejected)
		}
	}
}
