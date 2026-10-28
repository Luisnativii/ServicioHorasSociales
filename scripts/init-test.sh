#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
docker info >/dev/null 2>&1 || { echo 'Inicia Docker para preparar hs-test.' >&2; exit 1; }
umask 077
mkdir -p .secrets
if [ ! -e .secrets/hs-test.env ]; then
  docker run --rm --user "$(id -u):$(id -g)" \
    -e LOCAL_UID="$(id -u)" -e LOCAL_GID="$(id -g)" \
    -v "$PWD/.secrets:/config" node:22.23.4-bookworm-slim node -e '
      const password = require("node:crypto").randomBytes(32).toString("hex");
      const config = `POSTGRES_USER=hs_test\nPOSTGRES_DB=horas_sociales_test\nPOSTGRES_PASSWORD=${password}\nLOCAL_UID=${process.env.LOCAL_UID}\nLOCAL_GID=${process.env.LOCAL_GID}\nDB_MAX_CONNS=5\n`;
      require("node:fs").writeFileSync("/config/hs-test.env", config, { flag: "wx", mode: 0o600 });
    '
fi
bash scripts/dc-test.sh config --quiet
printf 'hs-test preparado: base horas_sociales_test, secretos independientes conservados.\n'
