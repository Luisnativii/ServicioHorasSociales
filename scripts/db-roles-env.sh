#!/usr/bin/env bash
# Biblioteca de configuración; nunca ejecutar archivos de secretos como shell.
check_db_roles_file() {
  local file=$1 key value seen='|'
  [ -f "$file" ] && [ ! -L "$file" ] || { echo 'Falta el archivo regular de roles de base.' >&2; return 1; }
  while IFS='=' read -r key value || [ -n "$key" ]; do
    case "$seen" in *"|$key|"*) echo 'Configuración de roles duplicada.' >&2; return 1;; esac
    seen="$seen$key|"
    case "$key" in
      HS_APP_DB_USER) [ "$value" = hs_app ] || return 1;;
      HS_MIGRATOR_DB_USER) [ "$value" = hs_migrator ] || return 1;;
      HS_APP_DB_PASSWORD|HS_MIGRATOR_DB_PASSWORD) [[ "$value" =~ ^[a-f0-9]{64}$ ]] || return 1;;
      *) echo 'Clave de roles de base desconocida.' >&2; return 1;;
    esac
  done < "$file"
  for key in HS_APP_DB_USER HS_MIGRATOR_DB_USER HS_APP_DB_PASSWORD HS_MIGRATOR_DB_PASSWORD; do
    case "$seen" in *"|$key|"*) ;; *) echo 'Configuración de roles incompleta.' >&2; return 1;; esac
    [ -z "${!key:-}" ] || { echo "Retira la variable exportada $key antes de usar roles de base." >&2; return 1; }
  done
}

db_roles_env_args() {
  db_roles_args=()
  local file=".secrets/$1-db-roles.env"
  if [ -e "$file.ready" ]; then
    [ -f "$file.ready" ] && [ ! -L "$file.ready" ] || return 1
    check_db_roles_file "$file" || return 1
    db_roles_args=(--env-file "$file")
  fi
}
