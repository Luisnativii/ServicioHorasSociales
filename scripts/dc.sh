#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
APP_VERSION=$(bash scripts/version.sh)
export APP_VERSION
source scripts/db-roles-env.sh
db_roles_env_args hs-dev
jwt_args=()
if [ -e .secrets/hs-dev-jwt.key ] || [ -L .secrets/hs-dev-jwt.key ]; then
  [ -f .secrets/hs-dev-jwt.key ] && [ ! -L .secrets/hs-dev-jwt.key ] || { echo 'Clave JWT: se exige un archivo regular, sin enlaces.' >&2; exit 1; }
  jwt_args=(-f compose.jwt.yaml)
fi
exec docker compose --env-file .env ${db_roles_args[@]+"${db_roles_args[@]}"} -p hs-dev -f compose.yaml -f compose.dev.yaml ${jwt_args[@]+"${jwt_args[@]}"} "$@"
