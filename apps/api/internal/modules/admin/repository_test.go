package admin

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Synthetic PHC-shaped value: persistence tests do not create a real login.
const testHash = "$argon2id$v=19$m=65536,t=3,p=4$c2FsdC1zaW50ZXRpY28$cmVzdWx0YWRvLXNpbnRldGljbw"

func TestCanonicalEmail(t *testing.T) {
	got, err := CanonicalEmail(" ADMIN@EXAMPLE.TEST ")
	if err != nil || got != "admin@example.test" {
		t.Fatalf("correo=%q err=%v", got, err)
	}
	for _, input := range []string{"", "invalid", "Persona <admin@example.test>", "a@example.test,b@example.test", "a@example.test\nBcc:b@example.test"} {
		if _, err := CanonicalEmail(input); !errors.Is(err, ErrInvalidEmail) {
			t.Fatal("se aceptó correo inválido")
		}
	}
	encoded, err := json.Marshal(Administrator{ContrasenaHash: testHash})
	if err != nil || strings.Contains(string(encoded), testHash) || strings.Contains(string(encoded), "ContrasenaHash") {
		t.Fatal("se expuso el hash en JSON")
	}
	if storageError(&pgconn.PgError{Message: "secret SQL"}) != ErrStorage {
		t.Fatal("error de almacenamiento no controlado")
	}
}

func TestRepositoryPostgreSQL(t *testing.T) {
	if os.Getenv("HS_TEST_IDENTITY") != "1" {
		t.Skip("integración explícita mediante scripts/check-admin.sh")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	config, err := pgx.ParseConfig(os.Getenv("DATABASE_URL"))
	if err != nil || config.Host != "db" || config.Database != "horas_sociales_test" || config.User != "hs_app" {
		t.Fatal("se exige hs-test con cuenta hs_app")
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
	first, err := repository.Create(ctx, " FIRST@EXAMPLE.TEST ", testHash)
	if err != nil || !first.ID.Valid || !first.Activo || first.Correo != "first@example.test" {
		t.Fatalf("alta inicial: %v", err)
	}
	second, err := repository.Create(ctx, "second@example.test", testHash)
	if err != nil || first.ID == second.ID {
		t.Fatalf("múltiples cuentas: %v", err)
	}
	found, err := repository.FindByEmail(ctx, "First@Example.Test")
	if err != nil || found.ID != first.ID || found.ContrasenaHash != testHash {
		t.Fatal("consulta por correo incorrecta")
	}
	found, err = repository.FindByID(ctx, second.ID)
	if err != nil || found.Correo != second.Correo {
		t.Fatal("consulta por identidad incorrecta")
	}
	// Savepoints allow checking rejected statements without aborting the fixture.
	child, err := tx.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, duplicate := NewRepository(child).Create(ctx, "first@EXAMPLE.TEST", testHash)
	_ = child.Rollback(ctx)
	if !errors.Is(duplicate, ErrDuplicate) {
		t.Fatalf("duplicado: %v", duplicate)
	}
	for _, sql := range []string{
		`INSERT INTO horas_sociales.administrador(correo,contrasena_hash) VALUES ('UPPER@EXAMPLE.TEST','$argon2id$synthetic')`,
		`INSERT INTO horas_sociales.administrador(correo,contrasena_hash) VALUES (' plain@example.test ','$argon2id$synthetic')`,
		`INSERT INTO horas_sociales.administrador(correo,contrasena_hash) VALUES ('plain@example.test','plaintext')`,
	} {
		child, err := tx.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		_, rejected := child.Exec(ctx, sql)
		_ = child.Rollback(ctx)
		var postgres *pgconn.PgError
		if !errors.As(rejected, &postgres) || postgres.Code != "23514" {
			t.Fatal("la base no rechazó datos fuera del contrato")
		}
	}
	for _, id := range []Administrator{first, second} {
		changed, err := repository.SetActive(ctx, id.ID, false)
		if err != nil || changed.Activo || changed.ContrasenaHash != testHash {
			t.Fatal("desactivación incorrecta")
		}
	}
	page, err := repository.List(ctx, 100, 0)
	if err != nil || len(page) < 2 {
		t.Fatal("listado incompleto")
	}
	reversed, err := repository.List(ctx, 100, 0, true)
	if err != nil || len(reversed) != len(page) {
		t.Fatal("listado descendente incompleto")
	}
	for i := range page {
		if page[i].ID != reversed[len(page)-1-i].ID {
			t.Fatal("orden descendente no invierte el orden estable")
		}
	}
	slice, err := repository.List(ctx, 1, 1, true)
	if err != nil || len(slice) != 1 || slice[0].ID != reversed[1].ID {
		t.Fatal("paginación descendente incorrecta")
	}
	for _, item := range page {
		if item.ContrasenaHash != "" {
			t.Fatal("el listado no debe consultar hashes")
		}
	}
	if _, err := repository.FindByEmail(ctx, "missing@example.test"); !errors.Is(err, ErrNotFound) {
		t.Fatal("cuenta inexistente")
	}
	if _, err := repository.List(ctx, 101, 0); !errors.Is(err, ErrPagination) {
		t.Fatal("se aceptó un listado sin límite válido")
	}
}
