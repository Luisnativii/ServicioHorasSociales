#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
APP_VERSION=$(bash scripts/version.sh)
export APP_VERSION
source scripts/db-roles-env.sh
db_roles_env_args hs-dev
exec docker compose --env-file .env ${db_roles_args[@]+"${db_roles_args[@]}"} -p hs-dev -f compose.yaml -f compose.dev.yaml "$@"
