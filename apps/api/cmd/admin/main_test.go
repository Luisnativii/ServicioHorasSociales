package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"horas-sociales/api/internal/modules/admin"
	"horas-sociales/api/internal/platform/password"
)

func TestSecretFileAndArguments(t *testing.T) {
	file := filepath.Join(t.TempDir(), "password")
	if err := os.WriteFile(file, []byte(" clave1 "), 0600); err != nil {
		t.Fatal(err)
	}
	value, err := readSecretFile(file)
	if err != nil || string(value) != " clave1 " {
		t.Fatal("se modificó el secreto")
	}
	clear(value)
	if err := os.Chmod(file, 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := readSecretFile(file); !errors.Is(err, errSecretFile) {
		t.Fatal("se aceptó archivo legible por terceros")
	}
	if err := os.Chmod(file, 0600); err != nil {
		t.Fatal(err)
	}
	link := file + "-link"
	if err := os.Symlink(file, link); err != nil {
		t.Fatal(err)
	}
	if _, err := readSecretFile(link); !errors.Is(err, errSecretFile) {
		t.Fatal("se aceptó enlace")
	}
	for _, bad := range []string{"clave1\n", strings.Repeat("x", 129)} {
		if err := os.WriteFile(file, []byte(bad), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := readSecretFile(file); !errors.Is(err, errSecretFile) {
			t.Fatal("se aceptó archivo fuera del formato")
		}
	}
	input, err := os.Open(file)
	if err != nil {
		t.Fatal(err)
	}
	defer input.Close()
	var out bytes.Buffer
	if _, err := readSecret(input, &out, ""); !errors.Is(err, errTerminal) {
		t.Fatal("se aceptó stdin sin TTY")
	}
	for _, args := range [][]string{{"reset"}, {"create", "--email", "admin@example.test", "--password", "secreto123"}, {"create", "--desconocido=secreto123"}} {
		out.Reset()
		if run(args, input, &out, &out) == nil || strings.Contains(out.String(), "secreto123") {
			t.Fatal("argumentos inseguros o eco de secreto")
		}
	}
}

func TestCreatePostgreSQL(t *testing.T) {
	if os.Getenv("HS_TEST_ADMIN_CLI") != "1" {
		t.Skip("integración explícita mediante check-admin-cli.sh")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cfg, err := pgx.ParseConfig(os.Getenv("DATABASE_URL"))
	if err != nil || cfg.Host != "db" || cfg.Database != "horas_sociales_test" || cfg.User != "hs_app" {
		t.Fatal("se exige hs-test/hs_app")
	}
	conn, err := pgx.ConnectConfig(ctx, cfg)
	if err != nil {
		t.Fatal("no conecta hs-test")
	}
	defer conn.Close(context.Background())
	repo := admin.NewRepository(conn)
	suffix := time.Now().UnixNano()
	emails := []string{fmt.Sprintf("cli-a-%d@example.test", suffix), fmt.Sprintf("cli-b-%d@example.test", suffix), fmt.Sprintf("cli-c-%d@example.test", suffix)}
	created := []string{}
	defer func() {
		if _, err := conn.Exec(context.Background(), "DELETE FROM horas_sociales.administrador WHERE correo = ANY($1::text[])", created); err != nil {
			t.Error("no se limpiaron fixtures de CLI")
		}
	}()
	secret := filepath.Join(t.TempDir(), "password")
	if err := os.WriteFile(secret, []byte("prueba1"), 0600); err != nil {
		t.Fatal(err)
	}
	input, err := os.Open(secret)
	if err != nil {
		t.Fatal(err)
	}
	defer input.Close()
	var output bytes.Buffer
	create := func(email string) error {
		output.Reset()
		err := run([]string{"create", "--email", email, "--password-file", secret}, input, &output, &output)
		if err == nil {
			canonical, _ := admin.CanonicalEmail(email)
			created = append(created, canonical)
		}
		return err
	}
	if err := create(" " + strings.ToUpper(emails[0]) + " "); err != nil {
		t.Fatal(err)
	}
	a, err := repo.FindByEmail(ctx, emails[0])
	if err != nil {
		t.Fatal(err)
	}
	if ok, err := password.Verify(ctx, []byte("prueba1"), a.ContrasenaHash); err != nil || !ok {
		t.Fatal("no se persistió un hash verificable")
	}
	if strings.Contains(output.String(), "prueba1") || strings.Contains(output.String(), a.ContrasenaHash) {
		t.Fatal("se expuso secreto")
	}
	if err := create(emails[1]); err != nil {
		t.Fatal(err)
	}
	b, err := repo.FindByEmail(ctx, emails[1])
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(secret, []byte("otraClave2"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := create(emails[0]); !errors.Is(err, admin.ErrDuplicate) {
		t.Fatal("no se rechazó duplicado")
	}
	after, err := repo.FindByID(ctx, a.ID)
	if err != nil || after.ContrasenaHash != a.ContrasenaHash {
		t.Fatal("se sobrescribió A")
	}
	after, err = repo.FindByID(ctx, b.ID)
	if err != nil || after.ContrasenaHash != b.ContrasenaHash {
		t.Fatal("se modificó B")
	}
	for _, account := range []admin.Administrator{a, b} {
		if _, err := repo.SetActive(ctx, account.ID, false); err != nil {
			t.Fatal(err)
		}
	}
	if err := create(emails[2]); err != nil {
		t.Fatal("no se permitió otra cuenta tras desactivar las existentes")
	}
}
