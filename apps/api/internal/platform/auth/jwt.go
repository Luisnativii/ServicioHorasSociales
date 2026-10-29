// Package auth implements the approved stateless ADMIN JWT contract.
package auth

import (
	"context"
	"errors"
	"io"
	"os"
	"syscall"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	Issuer     = "horas-sociales-api"
	Audience   = "horas-sociales-web"
	Lifetime   = 2 * time.Hour
	ClockSkew  = 30 * time.Second
	CookieName = "hs_access"
)

var (
	ErrToken = errors.New("credencial no válida")
	ErrKey   = errors.New("JWT_KEY_FILE requiere un archivo regular protegido de 32 bytes; ejecuta scripts/init-jwt-key.sh")
)

type Identity struct{ ID string }
type identityKey struct{}

func FromContext(ctx context.Context) (Identity, bool) {
	i, ok := ctx.Value(identityKey{}).(Identity)
	return i, ok
}

func WithIdentity(ctx context.Context, i Identity) context.Context {
	return context.WithValue(ctx, identityKey{}, i)
}

type Tokens struct {
	key []byte
	now func() time.Time
}

// New copies the key so the caller can clear its input buffer.
func New(key []byte) (*Tokens, error) {
	if len(key) != 32 {
		return nil, ErrKey
	}
	return &Tokens{key: append([]byte(nil), key...), now: time.Now}, nil
}

// LoadKey never generates, replaces or prints a key. Docker Desktop may report
// a protected host bind mount as root-owned; readability still comes from OS permissions.
func LoadKey(path string) (*Tokens, error) {
	f, err := os.OpenFile(path, os.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0)
	if err != nil {
		return nil, ErrKey
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() || (info.Mode().Perm() != 0400 && info.Mode().Perm() != 0600) {
		return nil, ErrKey
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok || (stat.Uid != 0 && int(stat.Uid) != os.Geteuid()) {
		return nil, ErrKey
	}
	key, err := io.ReadAll(io.LimitReader(f, 33))
	defer clear(key)
	if err != nil {
		return nil, ErrKey
	}
	return New(key)
}

// ValidSubject accepts canonical nonzero UUIDs, without accessing storage.
func ValidSubject(id string) bool {
	if len(id) != 36 || id == "00000000-0000-0000-0000-000000000000" {
		return false
	}
	for n, c := range id {
		if n == 8 || n == 13 || n == 18 || n == 23 {
			if c != '-' {
				return false
			}
			continue
		}
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}

// Issue is used only after a successful login in #26; it is not an HTTP endpoint.
func (t *Tokens) Issue(subject string) (string, error) {
	if t == nil || !ValidSubject(subject) {
		return "", ErrToken
	}
	now := t.now().UTC().Truncate(time.Second)
	claims := jwt.RegisteredClaims{Issuer: Issuer, Audience: jwt.ClaimStrings{Audience}, Subject: subject,
		IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(Lifetime))}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(t.key)
}

func (t *Tokens) Verify(raw string) (Identity, error) {
	if t == nil || len(raw) == 0 || len(raw) > 4096 {
		return Identity{}, ErrToken
	}
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, ErrToken
		}
		return t.key, nil
	}, jwt.WithValidMethods([]string{"HS256"}), jwt.WithIssuer(Issuer), jwt.WithAudience(Audience),
		jwt.WithExpirationRequired(), jwt.WithIssuedAt(), jwt.WithLeeway(ClockSkew),
		jwt.WithTimeFunc(t.now), jwt.WithStrictDecoding())
	if err != nil || token == nil || !token.Valid || !ValidSubject(claims.Subject) || claims.IssuedAt == nil || claims.ExpiresAt == nil {
		return Identity{}, ErrToken
	}
	if claims.IssuedAt.Unix() <= 0 || claims.ExpiresAt.Sub(claims.IssuedAt.Time) != Lifetime {
		return Identity{}, ErrToken
	}
	return Identity{ID: claims.Subject}, nil
}
