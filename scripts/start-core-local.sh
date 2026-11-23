#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
docker version --format '{{.Server.Version}}' >/dev/null
[ ! -L .secrets ] || { echo '.secrets debe ser un directorio regular.' >&2; exit 1; }
umask 077
mkdir -p .secrets
config=.secrets/hs-core-local.env
if [ ! -e "$config" ] && [ ! -L "$config" ]; then
  password=$(docker run --rm node:22.23.4-bookworm-slim \
    node -p 'require("node:crypto").randomBytes(32).toString("hex")')
  [[ "$password" =~ ^[a-f0-9]{64}$ ]]
  (set -C; printf 'POSTGRES_USER=hs_core_local\nPOSTGRES_DB=horas_sociales\nPOSTGRES_PASSWORD=%s\n' "$password" > "$config")
  unset password
fi

dc() { bash scripts/dc-core-local.sh "$@"; }
dc config --quiet
dc up -d --wait --wait-timeout 180 db
dc run --rm --no-deps -T jwt-init
bash scripts/migrate.sh core-local

# Los catálogos del core son de solo lectura. Los permisos institucionales se conceden por columna.
dc exec -T db sh -c 'exec psql -X -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB"' <<'SQL'
GRANT USAGE ON SCHEMA public TO hs_app;
GRANT SELECT ON TABLE public.faculties, public.careers,
    public.geo_departments, public.geo_municipalities, public.geo_districts TO hs_app;
SQL

dc exec -T db sh -c 'exec psql -X -v ON_ERROR_STOP=1 -v app_role=hs_app -U "$POSTGRES_USER" -d "$POSTGRES_DB"' \
  < deploy/core-local/contacto-institucional.sql

dc exec -T db sh -c 'exec psql -X -v ON_ERROR_STOP=1 -v app_role=hs_app -U "$POSTGRES_USER" -d "$POSTGRES_DB"' \
  < deploy/core-local/estudiantes-vigencias.sql

dc exec -T db sh -c 'exec psql -X -v ON_ERROR_STOP=1 -v app_role=hs_app -U "$POSTGRES_USER" -d "$POSTGRES_DB"' \
  < deploy/core-local/proyectos-permisos.sql

dc exec -T db sh -c 'exec psql -X -v ON_ERROR_STOP=1 -v app_role=hs_app -U "$POSTGRES_USER" -d "$POSTGRES_DB"' \
  < deploy/core-local/participaciones-permisos.sql

dc exec -T db sh -c 'exec psql -X -v ON_ERROR_STOP=1 -v app_role=hs_app -U "$POSTGRES_USER" -d "$POSTGRES_DB"' \
  < deploy/core-local/cierres-permisos.sql

dc exec -T db sh -c 'exec psql -X -v ON_ERROR_STOP=1 -v app_role=hs_app -U "$POSTGRES_USER" -d "$POSTGRES_DB"' \
  < deploy/core-local/horas-permisos.sql

dc up -d --build --wait --wait-timeout 180 db api web
printf 'Portal local: http://127.0.0.1:%s/\n' "${HS_CORE_LOCAL_PORT:-18081}"
