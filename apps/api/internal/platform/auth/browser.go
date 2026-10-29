package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const CSRFCookie = "hs_csrf"
const CSRFHeader = "X-CSRF-Token"

var ErrBrowser = errors.New("APP_ORIGIN requiere un origen HTTPS explícito; HTTP solo se permite en localhost o 127.0.0.1")
var ErrCSRF = errors.New("solicitud no permitida")

type Browser struct {
	tokens  *Tokens
	origin  string
	host    string
	secure  bool
	csrfKey []byte
}

// NewBrowser trusts only configured origin, never forwarded request headers.
func NewBrowser(tokens *Tokens, origin string) (*Browser, error) {
	u, err := url.Parse(origin)
	if tokens == nil || err != nil || u == nil || u.User != nil || u.Hostname() == "" || u.Path != "" || u.RawQuery != "" || u.Fragment != "" || u.Opaque != "" || u.ForceQuery {
		return nil, ErrBrowser
	}
	if u.Scheme != "https" && !(u.Scheme == "http" && (u.Hostname() == "localhost" || u.Hostname() == "127.0.0.1")) {
		return nil, ErrBrowser
	}
	if origin != u.Scheme+"://"+u.Host {
		return nil, ErrBrowser
	}
	mac := hmac.New(sha256.New, tokens.key)
	mac.Write([]byte("horas-sociales/csrf-key/v1"))
	return &Browser{tokens: tokens, origin: origin, host: u.Host, secure: u.Scheme == "https", csrfKey: mac.Sum(nil)}, nil
}

// OneCookie rejects ambiguous names instead of choosing an attacker-controlled path.
func OneCookie(r *http.Request, name string) (string, bool) {
	value, count := "", 0
	for _, c := range r.Cookies() {
		if c.Name == name {
			value = c.Value
			count++
		}
	}
	return value, count == 1
}

func (b *Browser) originAllowed(r *http.Request, required bool) bool {
	if b == nil || r.Host != b.host {
		return false
	}
	values := r.Header.Values("Origin")
	if len(values) == 0 {
		return !required
	}
	return len(values) == 1 && values[0] == b.origin
}

// binding distinguishes prelogin/expired-cookie requests from a verified JWT.
// This permits future logout without a live JWT, after fetching a new CSRF cookie.
func (b *Browser) binding(r *http.Request) (string, error) {
	raw, ok := OneCookie(r, CookieName)
	count := 0
	for _, c := range r.Cookies() {
		if c.Name == CookieName {
			count++
		}
	}
	if count > 1 {
		return "", ErrCSRF
	}
	if ok {
		if _, err := b.tokens.Verify(raw); err == nil {
			return raw, nil
		}
	}
	return "", nil
}

func (b *Browser) signature(payload, binding string) []byte {
	mac := hmac.New(sha256.New, b.csrfKey)
	// Hashing keeps the JWT out of the readable token. Fixed-size binding and
	// length-prefixing prevent ambiguous concatenation.
	digest := sha256.Sum256([]byte(binding))
	mac.Write(digest[:])
	mac.Write([]byte(strconv.Itoa(len(payload)) + ":" + payload))
	return mac.Sum(nil)
}

func (b *Browser) mint(binding string) (string, error) {
	nonce := make([]byte, 32)
	if _, err := rand.Read(nonce); err != nil {
		return "", ErrCSRF
	}
	payload := "v1." + strconv.FormatInt(b.tokens.now().Unix(), 10) + "." + base64.RawURLEncoding.EncodeToString(nonce)
	return payload + "." + base64.RawURLEncoding.EncodeToString(b.signature(payload, binding)), nil
}

func (b *Browser) cookie(name, value, path string, private bool, age int) *http.Cookie {
	return &http.Cookie{Name: name, Value: value, Path: path, HttpOnly: private, Secure: b.secure, SameSite: http.SameSiteStrictMode, MaxAge: age}
}

// PrepareCSRF never emits an access cookie or extends the JWT lifetime.
func (b *Browser) PrepareCSRF(w http.ResponseWriter, r *http.Request) error {
	if !b.originAllowed(r, false) || r.Header.Get("Sec-Fetch-Site") == "cross-site" {
		return ErrCSRF
	}
	binding, err := b.binding(r)
	if err != nil {
		return err
	}
	value, err := b.mint(binding)
	if err != nil {
		return err
	}
	http.SetCookie(w, b.cookie(CSRFCookie, value, "/", false, int(Lifetime.Seconds())))
	return nil
}

// Check requires both origin and signed cookie-to-header proof. A signed
// prelogin token cannot authorize a mutation made with an authenticated JWT.
func (b *Browser) Check(r *http.Request) error {
	if !b.originAllowed(r, true) {
		return ErrCSRF
	}
	value, ok := OneCookie(r, CSRFCookie)
	headers := r.Header.Values(CSRFHeader)
	if !ok || len(value) > 180 || len(headers) != 1 || subtle.ConstantTimeCompare([]byte(value), []byte(headers[0])) != 1 {
		return ErrCSRF
	}
	parts := strings.Split(value, ".")
	if len(parts) != 4 || parts[0] != "v1" || len(parts[2]) != 43 || len(parts[3]) != 43 {
		return ErrCSRF
	}
	issued, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || issued <= 0 || strconv.FormatInt(issued, 10) != parts[1] {
		return ErrCSRF
	}
	now := b.tokens.now().Unix()
	if issued > now+int64(ClockSkew.Seconds()) || issued < now-int64((Lifetime+ClockSkew).Seconds()) {
		return ErrCSRF
	}
	if _, err := base64.RawURLEncoding.Strict().DecodeString(parts[2]); err != nil {
		return ErrCSRF
	}
	signature, err := base64.RawURLEncoding.Strict().DecodeString(parts[3])
	if err != nil {
		return ErrCSRF
	}
	binding, err := b.binding(r)
	if err != nil {
		return err
	}
	if !hmac.Equal(signature, b.signature(strings.Join(parts[:3], "."), binding)) {
		return ErrCSRF
	}
	return nil
}

// SetSession is for successful login (#26). Mint before writing either cookie,
// and rotate CSRF to bind it to the newly issued JWT. No token is returned in JSON.
func (b *Browser) SetSession(w http.ResponseWriter, raw string) error {
	if b == nil {
		return ErrCSRF
	}
	if _, err := b.tokens.Verify(raw); err != nil {
		return err
	}
	csrf, err := b.mint(raw)
	if err != nil {
		return err
	}
	w.Header().Set("Cache-Control", "no-store")
	http.SetCookie(w, b.cookie(CookieName, raw, "/api", true, int(Lifetime.Seconds())))
	http.SetCookie(w, b.cookie(CSRFCookie, csrf, "/", false, int(Lifetime.Seconds())))
	return nil
}

// ClearSession deletes the same host-only cookies, even after JWT expiration.
// It does not revoke a token stored elsewhere. The logout handler must Check first.
func (b *Browser) ClearSession(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store")
	for _, c := range []*http.Cookie{b.cookie(CookieName, "", "/api", true, -1), b.cookie(CSRFCookie, "", "/", false, -1)} {
		c.Expires = time.Unix(1, 0).UTC()
		http.SetCookie(w, c)
	}
}
