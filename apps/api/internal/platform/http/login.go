package httpserver

import (
	"context"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"horas-sociales/api/internal/modules/admin"
	"horas-sociales/api/internal/platform/auth"
	"horas-sociales/api/internal/platform/password"
	"horas-sociales/api/internal/platform/ratelimit"
)

// administradorBuscador recupera cuentas de administrador sin exponer su hash
// al transporte HTTP. La implementación productiva es admin.Repository.
type administradorBuscador interface {
	FindByEmail(context.Context, string) (admin.Administrator, error)
	FindByID(context.Context, pgtype.UUID) (admin.Administrator, error)
	List(context.Context, int32, int64, ...bool) ([]admin.Administrator, error)
	Count(context.Context) (int32, error)
	SetActive(context.Context, pgtype.UUID, bool) (admin.Administrator, error)
}

// hashPrueba se verifica siempre que el correo no exista, para que el costo
// de Argon2id no revele si hay una cuenta detrás del correo consultado.
const hashPrueba = "$argon2id$v=19$m=65536,t=3,p=4$aaaaaaaaaaaaaaaaaaaaaa$aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

type identidadRespuesta struct {
	ID     string `json:"id"`
	Correo string `json:"correo"`
}

func clienteIP(remota string) string {
	if i := strings.LastIndex(remota, ":"); i >= 0 {
		return remota[:i]
	}
	return remota
}

// idDeSesion valida el sub del JWT (UUID canónico) y lo convierte a pgtype.UUID.
func idDeSesion(s string) (pgtype.UUID, error) {
	raw, err := hex.DecodeString(strings.ReplaceAll(s, "-", ""))
	if err != nil || len(raw) != 16 {
		return pgtype.UUID{}, err
	}
	var id pgtype.UUID
	copy(id.Bytes[:], raw)
	id.Valid = true
	return id, nil
}

// loginHandler implementa POST /api/v1/auth/login. Todas las respuestas de
// fallo son idénticas para no revelar existencia ni estado de cuentas.
func loginHandler(buscar administradorBuscador, limiter *ratelimit.Limiter, tokens *auth.Tokens, browser *auth.Browser) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		ip := clienteIP(r.RemoteAddr)
		if retraso, ok := limiter.Allow(ip); !ok {
			w.Header().Set("Retry-After", strconv.FormatInt(int64(retraso.Seconds()), 10))
			WriteError(w, r, "too_many_requests")
			return
		}

		var entrada struct {
			Correo     string `json:"correo"`
			Contrasena string `json:"contrasena"`
		}
		if err := ReadJSON(w, r, &entrada); err != nil {
			WriteInputError(w, r, err)
			return
		}
		correo, err := admin.CanonicalEmail(entrada.Correo)
		if err != nil || entrada.Contrasena == "" {
			WriteError(w, r, "bad_request")
			return
		}

		cuenta, err := buscar.FindByEmail(ctx, correo)
		encontrada := err == nil
		hash := hashPrueba
		var activo bool
		if encontrada {
			hash = cuenta.ContrasenaHash
			activo = cuenta.Activo
		}
		coincide, _ := password.Verify(ctx, []byte(entrada.Contrasena), hash)

		exitoso := byte(0)
		if encontrada && activo && coincide {
			exitoso = 1
		}
		if subtle.ConstantTimeCompare([]byte{exitoso}, []byte{1}) != 1 {
			limiter.RegistrarFallo(ip)
			WriteError(w, r, "unauthorized")
			return
		}
		limiter.RegistrarExito(ip)

		raw, err := tokens.Issue(cuenta.ID.String())
		if err != nil {
			WriteError(w, r, "internal_error")
			return
		}
		if err := browser.SetSession(w, raw); err != nil {
			WriteError(w, r, "internal_error")
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		WriteJSON(w, r, http.StatusOK, identidadRespuesta{ID: cuenta.ID.String(), Correo: cuenta.Correo})
	}
}

// meHandler implementa GET /api/v1/auth/me. Una cuenta desactivada con JWT
// vigente conserva el acceso hasta expirar: no hay revocación.
func meHandler(buscar administradorBuscador) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		identidad, ok := auth.FromContext(r.Context())
		if !ok {
			WriteError(w, r, "unauthorized")
			return
		}
		id, err := idDeSesion(identidad.ID)
		if err != nil {
			WriteError(w, r, "unauthorized")
			return
		}
		cuenta, err := buscar.FindByID(ctx, id)
		if err != nil {
			WriteError(w, r, "unauthorized")
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		WriteJSON(w, r, http.StatusOK, identidadRespuesta{ID: cuenta.ID.String(), Correo: cuenta.Correo})
	}
}
