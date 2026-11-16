#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
config=.secrets/hs-core-local.env
[ -f "$config" ] && [ ! -L "$config" ] || {
  echo 'Prepara el entorno con bash scripts/start-core-local.sh.' >&2
  exit 1
}
for key in POSTGRES_USER POSTGRES_DB POSTGRES_PASSWORD; do
  [ -z "${!key+x}" ] || { echo "Retira $key del entorno antes de arrancar hs-core-local." >&2; exit 1; }
done
APP_VERSION=$(bash scripts/version.sh)
export APP_VERSION
source scripts/db-roles-env.sh
db_roles_env_args hs-core-local
# Git Bash debe pasar las rutas del interior de los contenedores tal cual.
export MSYS_NO_PATHCONV=1
exec docker compose --env-file "$config" ${db_roles_args[@]+"${db_roles_args[@]}"} \
  -p hs-core-local -f compose.yaml -f compose.core-local.yaml "$@"
