#!/bin/sh
# El entrypoint oficial puede ejecutar o cargar este archivo con source.
(
    set -eu
    cd /opt/core-migrations
    baseline=001_init_unified_schema.sql
    history=002_legacy_css_import_support.sql
    faculties=003_seed_uca_faculties.sql

    psql -X --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" \
        --set ON_ERROR_STOP=1 --single-transaction \
        --set "baseline_hash=$(sha256sum "$baseline" | cut -d ' ' -f 1)" \
        --set "history_hash=$(sha256sum "$history" | cut -d ' ' -f 1)" \
        --set "faculties_hash=$(sha256sum "$faculties" | cut -d ' ' -f 1)" \
        --file "$baseline" --file "$history" --file "$faculties" --file - <<'SQL'
UPDATE public.schema_migrations
SET checksum = :'baseline_hash'
WHERE name = '001_init_unified_schema.sql';

INSERT INTO public.schema_migrations (name, checksum) VALUES
    ('002_legacy_css_import_support.sql', :'history_hash'),
    ('003_seed_uca_faculties.sql', :'faculties_hash');
SQL
)
