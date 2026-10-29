#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
usage() { echo 'Uso: bash scripts/admin.sh create --email CORREO [--password-file ARCHIVO_PROTEGIDO]' >&2; exit 1; }
[ "${1:-}" = create ] || usage
shift
email=''
password_file=''
while [ "$#" -gt 0 ]; do
  case "$1" in
    --email) [ "$#" -ge 2 ] && [ -z "$email" ] || usage; email=$2; shift 2;;
    --password-file) [ "$#" -ge 2 ] && [ -z "$password_file" ] || usage; password_file=$2; shift 2;;
    *) usage;;
  esac
done
[ -n "$email" ] || usage
bash scripts/dc.sh config --quiet
bash scripts/dc.sh build admin
if [ -n "$password_file" ]; then
  [ -f "$password_file" ] && [ ! -L "$password_file" ] || { echo 'Indica un archivo regular protegido, no un enlace.' >&2; exit 1; }
  password_file="$(cd "$(dirname "$password_file")" && pwd)/$(basename "$password_file")"
  exec bash scripts/dc.sh run --rm --no-deps -T --user "$(id -u):$(id -g)" \
    --volume "$password_file:/run/admin/password:ro" admin create --email "$email" --password-file /run/admin/password
fi
exec bash scripts/dc.sh run --rm --no-deps admin create --email "$email"
