#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
command -v docker >/dev/null || { echo 'No se encuentra Docker en esta terminal.' >&2; exit 1; }
docker info >/dev/null 2>&1 || { echo 'Docker no responde. Inicia Docker Desktop o Docker Engine.' >&2; exit 1; }
docker compose version >/dev/null
if [ ! -e .env ]; then
  docker run --rm --user "$(id -u):$(id -g)" \
    -e LOCAL_UID="$(id -u)" -e LOCAL_GID="$(id -g)" \
    -v "$PWD:/config" node:22.23.4-bookworm-slim node -e '
      const password = require("node:crypto").randomBytes(32).toString("hex");
      const config = `POSTGRES_USER=hs_local\nPOSTGRES_DB=horas_sociales\nPOSTGRES_PASSWORD=${password}\nLOCAL_UID=${process.env.LOCAL_UID}\nLOCAL_GID=${process.env.LOCAL_GID}\nDB_MAX_CONNS=5\n`;
      require("node:fs").writeFileSync("/config/.env", config, { flag: "wx", mode: 0o600 });
    '
fi
chmod 600 .env
bash scripts/dc.sh config --quiet
# Generar únicamente los lockfiles ausentes. Los existentes no se actualizan.
if [ ! -s apps/api/go.sum ]; then
  docker run --rm --user "$(id -u):$(id -g)" \
    -e HOME=/tmp -e GOCACHE=/tmp/cache -e GOPATH=/tmp/go -e GOTOOLCHAIN=local \
    -v "$PWD/apps/api:/src" -w /src golang:1.27.3-bookworm go mod tidy
fi
if [ ! -s apps/web/package-lock.json ]; then
  docker run --rm --user "$(id -u):$(id -g)" -e HOME=/tmp \
    -v "$PWD/apps/web:/src" -w /src node:22.23.4-bookworm-slim \
    npm install --package-lock-only --ignore-scripts --no-fund --no-audit
fi
printf 'Configuración local preparada; secretos existentes conservados.\n'
