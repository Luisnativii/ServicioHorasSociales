package main

import (
	"bytes"
	"context"
	"crypto/subtle"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/term"

	"horas-sociales/api/internal/config"
	"horas-sociales/api/internal/modules/admin"
	"horas-sociales/api/internal/platform/password"
)

const usage = "Uso: admin create --email CORREO [--password-file ARCHIVO_PROTEGIDO]"

var (
	errUsage        = errors.New(usage)
	errDatabase     = errors.New("no se pudo preparar identidad; ejecuta la migración y provisión de roles antes de crear cuentas")
	errSecretFile   = errors.New("el archivo debe ser regular, protegido (0400/0600), del usuario o de root, y contener la contraseña exacta sin saltos de línea")
	errTerminal     = errors.New("usa una terminal interactiva o --password-file con un archivo protegido; no se admite contraseña por argumentos, entorno o pipe")
	errConfirmation = errors.New("las contraseñas no coinciden")
)

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func run(args []string, input *os.File, output, prompt io.Writer) error {
	if len(args) == 1 && (args[0] == "--help" || args[0] == "-h") {
		fmt.Fprintln(output, usage)
		return nil
	}
	if len(args) == 0 || args[0] != "create" {
		return errUsage
	}
	flags := flag.NewFlagSet("admin create", flag.ContinueOnError)
	flags.SetOutput(io.Discard) // Never echo unsupported argument values.
	correo := flags.String("email", "", "")
	file := flags.String("password-file", "", "")
	if flags.Parse(args[1:]) != nil || flags.NArg() != 0 {
		return errUsage
	}
	canonical, err := admin.CanonicalEmail(*correo)
	if err != nil {
		return err
	}
	cfg, err := config.Load()
	if err != nil {
		return errDatabase
	}
	poolConfig, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil || poolConfig.ConnConfig.User != "hs_app" {
		return errDatabase
	}
	poolConfig.MaxConns = 1
	poolConfig.MinConns = 0
	poolConfig.ConnConfig.ConnectTimeout = 3 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		cancel()
		return errDatabase
	}
	defer pool.Close()
	var ready bool
	if pool.Ping(ctx) != nil || pool.QueryRow(ctx, "SELECT to_regclass('horas_sociales.administrador') IS NOT NULL").Scan(&ready) != nil || !ready {
		cancel()
		return errDatabase
	}
	repository := admin.NewRepository(pool)
	_, lookup := repository.FindByEmail(ctx, canonical)
	cancel()
	if lookup == nil {
		return admin.ErrDuplicate
	}
	if !errors.Is(lookup, admin.ErrNotFound) {
		return errDatabase
	}
	secret, err := readSecret(input, prompt, *file)
	if err != nil {
		return err
	}
	defer clear(secret)
	hash, err := password.Hash(context.Background(), secret)
	if err != nil {
		return err
	}
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := repository.Create(ctx, canonical, hash); err != nil {
		return err
	}
	fmt.Fprintln(output, "Cuenta ADMIN creada.")
	return nil
}

func readSecret(input *os.File, prompt io.Writer, path string) ([]byte, error) {
	if path != "" {
		return readSecretFile(path)
	}
	if !term.IsTerminal(int(input.Fd())) {
		return nil, errTerminal
	}
	fmt.Fprint(prompt, "Contraseña: ")
	value, err := term.ReadPassword(int(input.Fd()))
	fmt.Fprintln(prompt)
	if err != nil {
		clear(value)
		return nil, errTerminal
	}
	fmt.Fprint(prompt, "Confirmar contraseña: ")
	confirmation, err := term.ReadPassword(int(input.Fd()))
	fmt.Fprintln(prompt)
	defer clear(confirmation)
	if err != nil {
		clear(value)
		return nil, errTerminal
	}
	if subtle.ConstantTimeCompare(value, confirmation) != 1 {
		clear(value)
		return nil, errConfirmation
	}
	if err := password.Validate(value); err != nil {
		clear(value)
		return nil, err
	}
	return value, nil
}

func readSecretFile(path string) ([]byte, error) {
	file, err := os.OpenFile(path, os.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0)
	if err != nil {
		return nil, errSecretFile
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() || (info.Mode().Perm() != 0400 && info.Mode().Perm() != 0600) {
		return nil, errSecretFile
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	// Docker Desktop can report a host-owned bind mount as root-owned on
	// fstat. Root-owned secrets are accepted only if this process can already
	// open them; modes stay 0400/0600 and symlinks stay forbidden.
	if !ok || (int(stat.Uid) != os.Geteuid() && stat.Uid != 0) {
		return nil, errSecretFile
	}
	value, err := io.ReadAll(io.LimitReader(file, password.MaxBytes+1))
	if err != nil || len(value) > password.MaxBytes || bytes.ContainsAny(value, "\r\n") {
		clear(value)
		return nil, errSecretFile
	}
	if err := password.Validate(value); err != nil {
		clear(value)
		return nil, err
	}
	return value, nil
}
