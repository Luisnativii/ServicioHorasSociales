// Package password implements the accepted Argon2id/PHC contract.
package password

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"strings"
	"unicode/utf8"

	"golang.org/x/crypto/argon2"
)

const (
	MaxBytes   = 128
	parameters = "m=65536,t=3,p=4"
	prefix     = "$argon2id$v=19$" + parameters + "$"
)

var (
	ErrPolicy = errors.New("la contraseña requiere al menos seis caracteres, un número y un máximo de 128 bytes")
	ErrHash   = errors.New("hash de contraseña no válido o no compatible")
	ErrRandom = errors.New("no se pudo generar el salt de contraseña")
	workers   = make(chan struct{}, 2)
)

func Validate(value []byte) error {
	if len(value) > MaxBytes || !utf8.Valid(value) || utf8.RuneCount(value) < 6 {
		return ErrPolicy
	}
	for _, c := range value {
		if c >= '0' && c <= '9' {
			return nil
		}
	}
	return ErrPolicy
}

func acquire(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	select {
	case workers <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func Hash(ctx context.Context, value []byte) (string, error) {
	if err := Validate(value); err != nil {
		return "", err
	}
	if err := acquire(ctx); err != nil {
		return "", err
	}
	defer func() { <-workers }()
	salt := make([]byte, 16)
	defer clear(salt)
	if _, err := rand.Read(salt); err != nil {
		return "", ErrRandom
	}
	key := argon2.IDKey(value, salt, 3, 65536, 4, 32)
	defer clear(key)
	return prefix + base64.RawStdEncoding.EncodeToString(salt) + "$" + base64.RawStdEncoding.EncodeToString(key), nil
}

// Verify checks the fixed, accepted parameters before allocating Argon2 memory.
// Unsupported PHC parameters cannot request arbitrary memory/CPU resources.
func Verify(ctx context.Context, value []byte, encoded string) (bool, error) {
	if len(value) > MaxBytes || !utf8.Valid(value) {
		return false, ErrPolicy
	}
	if !strings.HasPrefix(encoded, prefix) || len(encoded) > 160 {
		return false, ErrHash
	}
	parts := strings.Split(strings.TrimPrefix(encoded, prefix), "$")
	if len(parts) != 2 || len(parts[0]) != 22 || len(parts[1]) != 43 {
		return false, ErrHash
	}
	salt, err := base64.RawStdEncoding.Strict().DecodeString(parts[0])
	if err != nil || len(salt) != 16 {
		return false, ErrHash
	}
	defer clear(salt)
	expected, err := base64.RawStdEncoding.Strict().DecodeString(parts[1])
	if err != nil || len(expected) != 32 {
		return false, ErrHash
	}
	defer clear(expected)
	if err := acquire(ctx); err != nil {
		return false, err
	}
	defer func() { <-workers }()
	actual := argon2.IDKey(value, salt, 3, 65536, 4, 32)
	defer clear(actual)
	return subtle.ConstantTimeCompare(expected, actual) == 1, nil
}
