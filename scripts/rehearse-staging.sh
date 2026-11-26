#!/usr/bin/env bash
# Isolated PostgreSQL recovery and image rollback using two accepted revisions.
set -euo pipefail
cd "$(dirname "$0")/.."
[ "$#" -eq 2 ] || { echo 'Uso: bash scripts/rehearse-staging.sh CANDIDATA ANTERIOR' >&2; exit 2; }
candidate=$(git rev-parse --verify "$1^{commit}")
previous=$(git rev-parse --verify "$2^{commit}")
[ "$candidate" != "$previous" ] && git merge-base --is-ancestor "$previous" "$candidate"
git merge-base --is-ancestor "$candidate" origin/develop
if ! git diff --quiet "$previous" "$candidate" -- apps/api/db/migrations; then
  echo 'Este ensayo requiere migraciones idénticas. Para revertir con cambios de esquema hace falta validar compatibilidad.' >&2
  exit 1
fi

# Prevent Compose from overriding generated credentials with exported values.
for key in POSTGRES_USER POSTGRES_DB POSTGRES_PASSWORD HS_APP_DB_USER HS_APP_DB_PASSWORD HS_MIGRATOR_DB_USER HS_MIGRATOR_DB_PASSWORD; do
  unset "$key"
done
umask 077
[ ! -L .secrets ]
mkdir -p .secrets
evidence=$(mktemp -d "$PWD/.secrets/hs-staging.XXXXXX")
project=hs-staging-$(date +%s)-$$
restore_project=$project-restore
target=$project
export HS_STAGING_DB_IMAGE=${HS_STAGING_DB_IMAGE:-postgres:16-alpine}
docker image inspect "$HS_STAGING_DB_IMAGE" >/dev/null 2>&1 || docker pull "$HS_STAGING_DB_IMAGE"
# Use the same local PostgreSQL artifact for both clusters.
export HS_STAGING_DB_IMAGE=$(docker image inspect --format '{{.Id}}' "$HS_STAGING_DB_IMAGE")
printf 'Evidencia local protegida: %s\n' "$evidence"
printf 'Origen: %s. Restauración: %s.\n' "$project" "$restore_project"

write_config() {
  local name=$1
  {
    printf 'POSTGRES_USER=hs_stage_owner\nPOSTGRES_DB=horas_sociales\n'
    for key in POSTGRES_PASSWORD HS_APP_DB_PASSWORD HS_MIGRATOR_DB_PASSWORD; do
      printf '%s=%s\n' "$key" "$(openssl rand -hex 32)"
    done
    printf 'HS_APP_DB_USER=hs_app\nHS_MIGRATOR_DB_USER=hs_migrator\n'
  } > "$evidence/$name.env"
}
write_config "$project"
write_config "$restore_project"

dc() {
  docker compose --env-file "$evidence/$target.env" -p "$target" \
    -f compose.yaml -f compose.core-local.yaml -f deploy/staging/compose.yaml "$@"
}
cleanup() {
  local code=$?
  trap - EXIT
  for target in "$project" "$restore_project"; do
    dc down >/dev/null || true
  done
  exit "$code"
}
db() {
  dc exec -T db sh -c 'exec psql -XqAt -v ON_ERROR_STOP=1 -v app_role=hs_app -U "$POSTGRES_USER" -d "$POSTGRES_DB"'
}
migrate() { dc run --rm --no-deps -T migrator "$@"; }
select_release() {
  local revision=$1 short
  short=$(git rev-parse --short=12 "$revision")
  export HS_STAGING_API_IMAGE=hs-staging-api:rc-$short
  export HS_STAGING_WEB_IMAGE=hs-staging-web:rc-$short
  export HS_STAGING_MIGRATOR_IMAGE=hs-staging-migrator:rc-$short
  export HS_STAGING_ADMIN_IMAGE=hs-staging-admin:rc-$short
}
build_release() {
  local revision=$1 dir=$evidence/source version tag
  local service dockerfile target_name ref build_target
  mkdir -p "$dir"
  # El directorio padre es privado; los archivos deben ser legibles en imágenes sin root.
  git archive "$revision" | (umask 022; tar -xf - -C "$dir")
  tag=$(git describe --tags --match 'V[0-9]*' --abbrev=0 "$revision")
  version=$(APP_VERSION=${tag#V} bash scripts/version.sh)
  select_release "$revision"
  printf 'commit=%s version=%s\n' "$revision" "$version" | tee -a "$evidence/images.txt"
  for service in api web migrator admin; do
    case "$service" in
      api) dockerfile=api; target_name=runtime; ref=$HS_STAGING_API_IMAGE;;
      web) dockerfile=web; target_name=runtime; ref=$HS_STAGING_WEB_IMAGE;;
      migrator) dockerfile=migrator; target_name=; ref=$HS_STAGING_MIGRATOR_IMAGE;;
      admin) dockerfile=api; target_name=admin-runtime; ref=$HS_STAGING_ADMIN_IMAGE;;
    esac
    build_target=()
    if [ -n "$target_name" ]; then build_target=(--target "$target_name"); fi
    if ! docker image inspect "$ref" >/dev/null 2>&1; then
      docker build --provenance=false ${build_target[@]+"${build_target[@]}"} \
        --build-arg "APP_VERSION=$version" --label "org.opencontainers.image.revision=$revision" \
        --tag "$ref" --file "$dir/deploy/docker/$dockerfile.Dockerfile" "$dir"
    fi
    [ "$(docker image inspect --format '{{index .Config.Labels "org.opencontainers.image.revision"}}' "$ref")" = "$revision" ]
    [ "$(docker image inspect --format '{{index .Config.Labels "org.opencontainers.image.version"}}' "$ref")" = "$version" ]
    printf '%s %s\n' "$ref" "$(docker image inspect --format '{{.Os}}/{{.Architecture}} {{.Id}}' "$ref")" | tee -a "$evidence/images.txt"
  done
  # Each accepted tree is unpacked into its own directory.
  mv "$dir" "$evidence/source-$revision"
}
build_release "$previous"
build_release "$candidate"
select_release "$previous"
trap cleanup EXIT
printf 'PostgreSQL %s\n' "$HS_STAGING_DB_IMAGE" | tee -a "$evidence/images.txt"
docker build -q -t "$project-browser" tests/e2e >/dev/null
dc config --quiet
dc up -d --wait --wait-timeout 120 db

# Load the same synthetic core fixture used by core-local, without bind mounts.
dc cp deploy/core-local/migrations db:/opt/core-migrations
dc exec -T db sh -s < deploy/core-local/init-db.sh
# The foundation runs once using the cluster owner. Subsequent migrations use hs_migrator.
owner_url() {
  local password
  password=$(sed -n 's/^POSTGRES_PASSWORD=//p' "$evidence/$target.env")
  printf 'postgres://hs_stage_owner:%s@db:5432/horas_sociales?sslmode=disable' "$password"
}
GOOSE_DBSTRING=$(owner_url) docker compose --env-file "$evidence/$target.env" -p "$target" \
  -f compose.yaml -f compose.core-local.yaml -f deploy/staging/compose.yaml \
  run --rm --no-deps -T -e GOOSE_DBSTRING migrator up-to 1

role_sql() {
  while IFS='=' read -r key value; do
    case "$key" in
      HS_APP_DB_PASSWORD) printf "\\\\set app_password '%s'\n" "$value";;
      HS_MIGRATOR_DB_PASSWORD) printf "\\\\set migrator_password '%s'\n" "$value";;
    esac
  done < "$evidence/$target.env"
}
{ role_sql; cat scripts/provision-db-roles.sql; } | db > "$evidence/roles.log" 2>&1
migrate up
db <<'SQL'
GRANT USAGE ON SCHEMA public TO hs_app;
GRANT SELECT ON public.faculties, public.careers,
    public.geo_departments, public.geo_municipalities, public.geo_districts TO hs_app;
SQL
for fixture in contacto-institucional estudiantes-vigencias proyectos-permisos participaciones-permisos cierres-permisos horas-permisos territory/load; do
  db < "deploy/core-local/$fixture.sql"
done
db <<'SQL'
INSERT INTO horas_sociales.tipo_proyecto(nombre) VALUES ('Tipo ensayo #93');
INSERT INTO horas_sociales.area(nombre) VALUES ('Área ensayo #93');
INSERT INTO horas_sociales.modalidad(nombre) VALUES ('Modalidad ensayo #93');
SQL
dc run --rm --no-deps -T jwt-init
docker run --rm --network none --mount "type=volume,source=${project}_jwt,target=/e2e-secrets" \
  --mount "type=bind,source=$PWD/tests/e2e/init.cjs,target=/init.cjs,readonly" node:22.23.4-bookworm-slim node /init.cjs
for account in a b; do
  dc run --rm --no-deps -T -v "${project}_jwt:/e2e-secrets:ro" admin create \
    --email "e2e-$account@example.test" --password-file "/e2e-secrets/$account.password"
done

start_portal() {
  local service container ref expected actual user
  dc up -d --no-deps --no-build --wait --wait-timeout 120 api web
  for service in api web; do
    container=$(dc ps -q "$service")
    ref=$HS_STAGING_API_IMAGE
    [ "$service" != web ] || ref=$HS_STAGING_WEB_IMAGE
    expected=$(docker image inspect --format '{{.Id}}' "$ref")
    actual=$(docker inspect --format '{{.Image}}' "$container")
    [ "$expected" = "$actual" ]
    user=$(docker inspect --format '{{.Config.User}}' "$container")
    case "$user" in ''|0|0:*|root|root:*) return 1;; esac
    [ -z "$(docker port "$container")" ]
    [ "$(docker inspect --format '{{range .Mounts}}{{if eq .Type "bind"}}bind{{end}}{{end}}' "$container")" = '' ]
  done
  [ -z "$(docker port "$(dc ps -q db)")" ]
  dc exec -T web wget -q -O - http://127.0.0.1:8080/api/health/ready | grep -qx '{"status":"ok"}'
}
journey() {
  docker run --rm --init --shm-size=256m --network "container:$(dc ps -q web)" \
    --mount "type=volume,source=${project}_jwt,target=/e2e-secrets,readonly" \
    -e "HS_E2E_RUN=$(date +%s)$$$1" "$project-browser" "$2"
}
snapshot() {
  db <<'SQL'
WITH items AS (
    SELECT 'core_migraciones' AS name, to_jsonb(t)::text AS item FROM public.schema_migrations t
    UNION ALL SELECT 'instituciones', to_jsonb(t)::text FROM public.institutions t
    UNION ALL SELECT 'estudiantes', to_jsonb(t)::text FROM public.students t
    UNION ALL SELECT 'proyectos', to_jsonb(t)::text FROM public.projects t
    UNION ALL SELECT 'participaciones', to_jsonb(t)::text FROM public.project_enrollments t
    UNION ALL SELECT 'horas', to_jsonb(t)::text FROM public.hour_movements t
    UNION ALL SELECT 'vigencias', to_jsonb(t)::text FROM public.student_career_periods t
    UNION ALL SELECT 'cierres', to_jsonb(t)::text FROM horas_sociales.cierre_evento t
    UNION ALL SELECT 'autores', to_jsonb(t)::text FROM horas_sociales.autor_movimiento t
    UNION ALL SELECT 'administradores', to_jsonb(t)::text FROM horas_sociales.administrador t
    UNION ALL SELECT 'portal_migraciones', to_jsonb(t)::text FROM public.goose_db_version t
)
SELECT name, count(*), md5(string_agg(item, ',' ORDER BY item)) FROM items GROUP BY name ORDER BY name;
SQL
}
start_portal
journey 1 journey-without-reports
snapshot > "$evidence/before.txt"
dc exec -T db sh -c 'exec pg_dump -Fc -U "$POSTGRES_USER" -d "$POSTGRES_DB"' > "$evidence/backup.dump"

select_release "$candidate"
migrate up
start_portal
journey 0 smoke
snapshot | diff - "$evidence/before.txt"
select_release "$previous"
start_portal
journey 0 smoke
snapshot | diff - "$evidence/before.txt"
echo 'Actualización y rollback de imágenes sin alterar datos ni migraciones.'

dc stop api web db
target=$restore_project
select_release "$candidate"
dc up -d --wait --wait-timeout 120 db
{
  role_sql
  cat <<'SQL'
CREATE ROLE hs_app LOGIN PASSWORD :'app_password';
CREATE ROLE hs_migrator LOGIN PASSWORD :'migrator_password';
SQL
} | db > "$evidence/restore-roles.log" 2>&1
dc exec -T db sh -c 'exec pg_restore --exit-on-error --single-transaction -U "$POSTGRES_USER" -d "$POSTGRES_DB"' \
  < "$evidence/backup.dump"
migrate up
snapshot | diff - "$evidence/before.txt"
dc run --rm --no-deps -T jwt-init
start_portal
journey 2 journey

docker image save -o "$evidence/candidate.tar" \
  "$HS_STAGING_API_IMAGE" "$HS_STAGING_WEB_IMAGE" "$HS_STAGING_MIGRATOR_IMAGE" "$HS_STAGING_ADMIN_IMAGE"
(cd "$evidence" && shasum -a 256 candidate.tar backup.dump > checksums.txt)
echo "Ensayo completado: Chromium sobre la candidata restaurada, datos verificados y artefactos en $evidence."
