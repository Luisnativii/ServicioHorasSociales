package auth

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestJWTContract(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	key := bytes.Repeat([]byte{42}, 32)
	service, _ := New(key)
	service.now = func() time.Time { return now }
	id := "b14f0325-1640-4014-a00e-15b249b78866"
	raw, err := service.Issue(id)
	if err != nil {
		t.Fatal(err)
	}
	identity, err := service.Verify(raw)
	if err != nil || identity.ID != id {
		t.Fatal("round trip")
	}
	other := "613699dd-13dc-47f2-943d-3e099e1f313e"
	second, _ := service.Issue(other)
	if identity, err := service.Verify(second); err != nil || identity.ID != other {
		t.Fatal("se mezclaron identidades")
	}
	claims := func() jwt.RegisteredClaims {
		return jwt.RegisteredClaims{Issuer: Issuer, Audience: jwt.ClaimStrings{Audience}, Subject: id,
			IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(Lifetime))}
	}
	for _, name := range []string{"issuer", "audience", "subject", "missing_iat", "missing_exp", "future", "expired", "duration", "hs384", "none", "signature"} {
		t.Run(name, func(t *testing.T) {
			c := claims()
			method := jwt.SigningMethodHS256
			var signingKey any = key
			switch name {
			case "issuer":
				c.Issuer = "otro"
			case "audience":
				c.Audience = jwt.ClaimStrings{"otro"}
			case "subject":
				c.Subject = "admin"
			case "missing_iat":
				c.IssuedAt = nil
			case "missing_exp":
				c.ExpiresAt = nil
			case "future":
				c.IssuedAt = jwt.NewNumericDate(now.Add(31 * time.Second))
				c.ExpiresAt = jwt.NewNumericDate(c.IssuedAt.Add(Lifetime))
			case "expired":
				c.ExpiresAt = jwt.NewNumericDate(now.Add(-31 * time.Second))
				c.IssuedAt = jwt.NewNumericDate(c.ExpiresAt.Add(-Lifetime))
			case "duration":
				c.ExpiresAt = jwt.NewNumericDate(now.Add(3 * time.Hour))
			case "hs384":
				method = jwt.SigningMethodHS384
			case "signature":
				signingKey = bytes.Repeat([]byte{1}, 32)
			}
			token := jwt.NewWithClaims(method, c)
			if name == "none" {
				token = jwt.NewWithClaims(jwt.SigningMethodNone, c)
				signingKey = jwt.UnsafeAllowNoneSignatureType
			}
			bad, err := token.SignedString(signingKey)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := service.Verify(bad); err == nil {
				t.Fatal("token inválido aceptado")
			}
		})
	}
	service.now = func() time.Time { return now.Add(Lifetime + 29*time.Second) }
	if _, err := service.Verify(raw); err != nil {
		t.Fatal("tolerancia documentada")
	}
	service.now = func() time.Time { return now.Add(Lifetime + ClockSkew) }
	if _, err := service.Verify(raw); err == nil {
		t.Fatal("límite de vencimiento")
	}
	var disabled *Tokens
	if _, err := disabled.Verify(raw); err == nil {
		t.Fatal("sin clave debe denegar")
	}
}

func TestKeyFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "key")
	key := bytes.Repeat([]byte{7}, 32)
	if err := os.WriteFile(path, key, 0600); err != nil {
		t.Fatal(err)
	}
	first, err := LoadKey(path)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := first.Issue("b14f0325-1640-4014-a00e-15b249b78866")
	second, err := LoadKey(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := second.Verify(raw); err != nil {
		t.Fatal("reinicio invalidó token")
	}
	os.Chmod(path, 0644)
	if _, err := LoadKey(path); err == nil {
		t.Fatal("archivo público")
	}
	os.Chmod(path, 0600)
	link := path + "-link"
	os.Symlink(path, link)
	if _, err := LoadKey(link); err == nil {
		t.Fatal("enlace")
	}
	if _, err := New(key[:31]); err == nil {
		t.Fatal("clave corta")
	}
}
