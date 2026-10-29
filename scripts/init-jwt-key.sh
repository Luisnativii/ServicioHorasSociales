#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
[ "$#" -eq 0 ] || { echo 'Uso: bash scripts/init-jwt-key.sh' >&2; exit 1; }
[ ! -L .secrets ] || { echo '.secrets debe ser un directorio regular.' >&2; exit 1; }
umask 077
mkdir -p .secrets
# Node runs inside the existing development image, never on the host.
# Exclusive creation preserves an existing key and never prints its contents.
bash scripts/dc.sh run --rm --no-deps -T --user "$(id -u):$(id -g)" \
  --volume "$PWD/.secrets:/keys" web node -e '
const fs = require("node:fs"), crypto = require("node:crypto");
const path = "/keys/hs-dev-jwt.key";
try {
  const fd = fs.openSync(path, "wx", 0o600);
  try { fs.writeFileSync(fd, crypto.randomBytes(32)); } finally { fs.closeSync(fd); }
  console.log("Clave JWT creada; se conservará entre arranques.");
} catch (e) {
  if (e.code !== "EEXIST") { console.error("No se pudo preparar la clave JWT."); process.exit(1); }
  const s = fs.lstatSync(path);
  if (!s.isFile() || s.isSymbolicLink() || s.size !== 32 || ![0o400, 0o600].includes(s.mode & 0o777)) {
    console.error("La clave existente no es un archivo protegido válido; se conserva sin modificar."); process.exit(1);
  }
  console.log("Se conserva la clave JWT existente.");
}'
