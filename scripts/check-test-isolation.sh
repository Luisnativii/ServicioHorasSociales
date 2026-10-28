#!/usr/bin/env bash
# Contrato de configuración: nunca acceder a hs-dev para comprobar rechazo.
set -euo pipefail
cd "$(dirname "$0")/.."
dc() { bash scripts/dc-test.sh "$@"; }
expect_rejection() {
  if "$@" >/dev/null 2>&1; then
    echo 'Se aceptó una configuración ambigua de pruebas.' >&2
    exit 1
  fi
}
expect_rejection env COMPOSE_PROJECT_NAME=hs-dev bash scripts/dc-test.sh config --quiet
expect_rejection env POSTGRES_DB=horas_sociales bash scripts/dc-test.sh config --quiet
expect_rejection env COMPOSE_FILE=compose.dev.yaml bash scripts/dc-test.sh config --quiet
expect_rejection dc --project-name hs-dev down
expect_rejection dc --env-file .env config --quiet
expect_rejection dc -f compose.dev.yaml config --quiet
expect_rejection dc down --volumes
# Archivos inválidos en una carpeta temporal: nunca editar secretos existentes.
fixture=$(mktemp -d)
cleanup() { rm -rf "$fixture"; }
trap cleanup EXIT
mkdir -p "$fixture/scripts" "$fixture/.secrets"
cp scripts/dc-test.sh "$fixture/scripts/"
printf 'POSTGRES_DB=horas_sociales\n' > "$fixture/.secrets/hs-test.env"
expect_rejection bash "$fixture/scripts/dc-test.sh" config --quiet
printf 'POSTGRES_DB=horas_sociales_test\nPOSTGRES_DB=horas_sociales_test\n' > "$fixture/.secrets/hs-test.env"
expect_rejection bash "$fixture/scripts/dc-test.sh" config --quiet
printf 'POSTGRES_DB=horas_sociales_test\n' > "$fixture/.secrets/hs-test.env"
expect_rejection bash "$fixture/scripts/dc-test.sh" config --quiet
# La configuración incluye secretos: solo viaja por stdin al contenedor.
dc config --format json | docker run --rm -i node:22.23.4-bookworm-slim node -e '
  const assert = require("node:assert/strict");
  let input = "";
  process.stdin.on("data", chunk => input += chunk);
  process.stdin.on("end", () => {
    const c = JSON.parse(input);
    assert.equal(c.name, "hs-test");
    assert.equal(c.services.db.environment.POSTGRES_DB, "horas_sociales_test");
    assert.equal(c.services.db.environment.POSTGRES_USER, "hs_test");
    assert.equal(c.volumes.pgdata.name, "hs-test_pgdata");
    for (const service of Object.values(c.services)) {
      assert.equal((service.ports || []).length, 0);
      assert.ok(!service.container_name);
      for (const mount of service.volumes || []) {
        if (mount.type === "bind") assert.equal(mount.read_only, true);
      }
    }
    for (const network of Object.values(c.networks)) {
      assert.ok(!network.external);
      assert.ok(network.name.startsWith("hs-test_"));
    }
    assert.match(c.services.api.environment.DATABASE_URL, /@db:5432\/horas_sociales_test\?/);
    assert.match(c.services.api.user, /^\d+:\d+$/);
    console.log("hs-test: proyecto, base, redes, volumen, permisos y rechazos comprobados.");
  });
'
dc --runtime config --format json | docker run --rm -i node:22.23.4-bookworm-slim node -e '
  const assert = require("node:assert/strict");
  let input = "";
  process.stdin.on("data", chunk => input += chunk);
  process.stdin.on("end", () => {
    const c = JSON.parse(input);
    assert.equal(c.name, "hs-test-runtime");
    assert.equal(c.networks.data.name, "hs-test_data");
    assert.equal(c.networks.data.external, true);
    assert.match(c.services.api.environment.DATABASE_URL, /@db:5432\/horas_sociales_test\?/);
    for (const service of Object.values(c.services)) assert.equal((service.ports || []).length, 0);
    console.log("hs-test-runtime: PostgreSQL de pruebas y ausencia de puertos comprobados.");
  });
'
