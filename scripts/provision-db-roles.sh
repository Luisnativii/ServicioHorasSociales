#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
case "${1:-}" in
  dev) environment=hs-dev; dc() { bash scripts/dc.sh "$@"; };;
  test) environment=hs-test; dc() { bash scripts/dc-test.sh "$@"; };;
  *) echo 'Uso: bash scripts/provision-db-roles.sh dev|test' >&2; exit 1;;
esac
source scripts/db-roles-env.sh
umask 077
mkdir -p .secrets
file=".secrets/$environment-db-roles.env"
if [ ! -e "$file" ]; then
  docker run --rm --user "$(id -u):$(id -g)" -e CONFIG_FILE="$environment-db-roles.env" \
    -v "$PWD/.secrets:/config" node:22.23.4-bookworm-slim node -e '
      const crypto = require("node:crypto"), fs = require("node:fs");
      const contents = `HS_APP_DB_USER=hs_app\nHS_APP_DB_PASSWORD=${crypto.randomBytes(32).toString("hex")}\nHS_MIGRATOR_DB_USER=hs_migrator\nHS_MIGRATOR_DB_PASSWORD=${crypto.randomBytes(32).toString("hex")}\n`;
      fs.writeFileSync("/config/" + process.env.CONFIG_FILE, contents, {flag:"wx", mode:0o600});
    '
fi
check_db_roles_file "$file" || { echo 'Revisa la configuración protegida de roles; no se reemplazó.' >&2; exit 1; }
while IFS='=' read -r key value; do
  case "$key" in HS_APP_DB_PASSWORD) app_password=$value;; HS_MIGRATOR_DB_PASSWORD) migrator_password=$value;; esac
done < "$file"
unset value
dc config --quiet
dc up -d --wait --wait-timeout 90 db
log=$(mktemp)
trap 'rm -f "$log"' EXIT
# Secretos pasan por stdin/entorno de un exec temporal, nunca por el comando SQL
# impreso ni por los argumentos de psql. Capturar posibles errores sin exponerlos.
if ! {
  printf "\\set app_password '%s'\n\\set migrator_password '%s'\n" "$app_password" "$migrator_password"
  cat scripts/provision-db-roles.sql
} | dc exec -T db sh -c 'exec psql -X -q -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB"' >"$log" 2>&1; then
  echo 'No se pudieron provisionar los roles. Comprueba la migración inicial, propiedad y privilegios con la cuenta original. No se reemplazaron credenciales existentes.' >&2
  exit 1
fi
for role in hs_app hs_migrator; do
  password=$app_password
  [ "$role" != hs_migrator ] || password=$migrator_password
  if ! dc exec -T -e "PGPASSWORD=$password" db sh -c 'exec psql -X -w -h 127.0.0.1 -U "$1" -d "$POSTGRES_DB" -tAc "SELECT 1"' sh "$role" >"$log" 2>&1; then
    echo 'Los roles existen pero no coinciden con los secretos conservados. Recupera su archivo original; no regeneres contraseñas ni borres volúmenes.' >&2
    exit 1
  fi
done
unset app_password migrator_password password
# Activar solo después de verificar ambas conexiones. La API se recrea en el
# siguiente up, no se interrumpe aquí ni se reinicia el clúster ya iniciado.
: > "$file.ready"
printf '%s: roles verificados; API hs_app y migraciones hs_migrator. Secretos conservados.\n' "$environment"
