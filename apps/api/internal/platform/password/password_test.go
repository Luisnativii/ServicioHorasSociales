package password

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestPolicy(t *testing.T) {
	for _, v := range []string{"abc12", "abcdef", strings.Repeat("a", 128) + "1", string([]byte{0xff, 'a', 'b', 'c', 'd', '1'})} {
		if Validate([]byte(v)) != ErrPolicy {
			t.Fatal("se aceptó contraseña fuera de política")
		}
	}
	for _, v := range []string{"abcde1", "ábcde1", " clave1 "} {
		if Validate([]byte(v)) != nil {
			t.Fatal("se impuso una restricción adicional")
		}
	}
}

func TestPHCRoundTrip(t *testing.T) {
	start := time.Now()
	value := []byte("prueba1") // Synthetic test input; never a default account.
	defer clear(value)
	one, err := Hash(context.Background(), value)
	if err != nil {
		t.Fatal(err)
	}
	two, err := Hash(context.Background(), value)
	if err != nil || one == two || !strings.HasPrefix(one, prefix) {
		t.Fatal("salt/parámetros incorrectos")
	}
	for _, encoded := range []string{one, two} {
		ok, err := Verify(context.Background(), value, encoded)
		if err != nil || !ok {
			t.Fatal("falló verificación del PHC generado")
		}
	}
	ok, err := Verify(context.Background(), []byte("otra1"), one)
	if err != nil || ok {
		t.Fatal("se aceptó contraseña incorrecta")
	}
	t.Logf("Dos hashes y tres verificaciones Argon2id (64 MiB, t=3, p=4): %s", time.Since(start))
	for _, encoded := range []string{"", "$argon2id$", strings.Replace(one, "m=65536", "m=999999999", 1), strings.Replace(one, "v=19", "v=16", 1), one + "$extra", prefix + "bad$base64"} {
		if _, err := Verify(context.Background(), value, encoded); !errors.Is(err, ErrHash) {
			t.Fatal("se aceptó PHC inválido/no compatible")
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := Hash(ctx, value); !errors.Is(err, context.Canceled) {
		t.Fatal("se ignoró cancelación")
	}
}
