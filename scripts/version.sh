#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
version=${APP_VERSION:-}
if [ -z "$version" ]; then
  tag=$(git describe --tags --match 'V[0-9]*' --abbrev=0 2>/dev/null || true)
  version=${tag#V}
  version=${version:-0.0.1}
fi
if [[ ! "$version" =~ ^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$ ]]; then
  echo 'Versión inválida: se requiere MAJOR.MINOR.PATCH sin prefijo V.' >&2
  exit 1
fi
printf '%s\n' "$version"
