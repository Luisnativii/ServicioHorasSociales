#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
dc() { bash scripts/dc-test.sh "$@"; }
dc config --quiet
dc run --rm --no-deps -e HS_TEST_ADMIN_CLI=1 api go test -race -v ./cmd/admin ./internal/platform/password
umask 077
secret=$(mktemp)
email="runtime-cli-$(date +%s)-$$@example.test"
created=0
cleanup() {
  result=$?
  trap - EXIT
  rm -f "$secret"
  if [ "$created" = 1 ]; then
    dc exec -T db sh -c 'exec psql -X -q -v ON_ERROR_STOP=1 -v account="$1" -U "$POSTGRES_USER" -d "$POSTGRES_DB"' sh "$email" <<'SQL' >/dev/null
DELETE FROM horas_sociales.administrador WHERE correo = :'account';
SQL
  fi
  exit "$result"
}
trap cleanup EXIT
# Synthetic test fixture only, never a development account or default password.
printf 'prueba1' > "$secret"
if ! dc run --rm --no-deps -T --user "$(id -u):$(id -g)" \
  --volume "$secret:/run/admin/password:ro" admin create --email "$email" --password-file /run/admin/password >/dev/null 2>&1; then
  echo 'Falló la creación con el runtime de administración en hs-test.' >&2
  exit 1
fi
created=1
count=$(dc exec -T db sh -c 'exec psql -X -q -At -v ON_ERROR_STOP=1 -v account="$1" -U "$POSTGRES_USER" -d "$POSTGRES_DB"' sh "$email" <<'SQL'
SELECT count(*) FROM horas_sociales.administrador WHERE correo = :'account' AND contrasena_hash LIKE '$argon2id$%';
SQL
)
[ "$count" = 1 ]
printf 'CLI runtime: cuenta sintética creada con archivo protegido en hs-test y retirada al terminar.\n'
