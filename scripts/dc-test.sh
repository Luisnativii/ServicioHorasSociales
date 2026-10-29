#!/usr/bin/env bash
# Única entrada de las comprobaciones a Compose; nunca carga .env de desarrollo.
set -euo pipefail
cd "$(dirname "$0")/.."
fail() { echo "hs-test: $1" >&2; exit 1; }
for key in COMPOSE_FILE COMPOSE_PROJECT_NAME COMPOSE_ENV_FILES COMPOSE_PROFILES POSTGRES_USER POSTGRES_DB POSTGRES_PASSWORD LOCAL_UID LOCAL_GID DB_MAX_CONNS DATABASE_URL GOOSE_DBSTRING HS_APP_DB_USER HS_APP_DB_PASSWORD HS_MIGRATOR_DB_USER HS_MIGRATOR_DB_PASSWORD; do
  [ -z "${!key:-}" ] || fail "retira la variable exportada $key; se exige configuración de pruebas explícita."
done
[ -f .secrets/hs-test.env ] && [ ! -L .secrets/hs-test.env ] || fail 'ejecuta scripts/init-test.sh; falta un archivo de secretos regular.'
# Validar sin ejecutar el archivo ni mostrar sus valores. Rechazar duplicados,
# interpolación y destinos distintos del contrato fijo de hs-test.
seen='|'
while IFS='=' read -r key value || [ -n "$key" ]; do
  case "$seen" in *"|$key|"*) fail 'clave duplicada en secretos de pruebas.';; esac
  seen="$seen$key|"
  case "$key" in
    POSTGRES_USER) [ "$value" = hs_test ] || fail 'usuario de pruebas inesperado.';;
    POSTGRES_DB) [ "$value" = horas_sociales_test ] || fail 'base de pruebas inesperada.';;
    POSTGRES_PASSWORD) [[ "$value" =~ ^[a-f0-9]{64}$ ]] || fail 'formato de secreto de pruebas inesperado.';;
    LOCAL_UID|LOCAL_GID) [[ "$value" =~ ^[0-9]+$ ]] || fail 'UID/GID de pruebas inválido.';;
    DB_MAX_CONNS) [ "$value" = 5 ] || fail 'pool de pruebas inesperado.';;
    *) fail 'clave desconocida en secretos de pruebas.';;
  esac
done < .secrets/hs-test.env
for key in POSTGRES_USER POSTGRES_DB POSTGRES_PASSWORD LOCAL_UID LOCAL_GID DB_MAX_CONNS; do
  case "$seen" in *"|$key|"*) ;; *) fail 'configuración de pruebas incompleta.';; esac
done
unset value
project=hs-test
files=(-f compose.yaml -f compose.test.yaml)
if [ "${1:-}" = --runtime ]; then
  project=hs-test-runtime
  files=(-f tests/compose.runtime.yaml)
  shift
fi
for arg in "$@"; do
  case "$arg" in
    -p|--project-name|--project-name=*|-p?*|-f|--file|--file=*|-f?*|--env-file|--env-file=*|--project-directory|--project-directory=*|--profile|--profile=*)
      fail 'no se permite sustituir proyecto, archivos, secretos o perfiles.';;
    -v|--volumes)
      case "${1:-}" in down|rm) fail 'la limpieza de pruebas conserva los volúmenes.';; esac;;
  esac
done
APP_VERSION=$(bash scripts/version.sh)
export APP_VERSION
source scripts/db-roles-env.sh
db_roles_env_args hs-test
exec docker compose --project-directory "$PWD" --env-file .secrets/hs-test.env \
  ${db_roles_args[@]+"${db_roles_args[@]}"} -p "$project" "${files[@]}" "$@"
