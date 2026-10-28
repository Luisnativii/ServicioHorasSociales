\set VERBOSITY terse
BEGIN;
SELECT pg_advisory_xact_lock(hashtextextended('hs-db-roles', 0));
SELECT format('CREATE ROLE hs_app LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS PASSWORD %L', :'app_password')
WHERE NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'hs_app') \gexec
SELECT format('CREATE ROLE hs_migrator LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS PASSWORD %L', :'migrator_password')
WHERE NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'hs_migrator') \gexec
DO $$
BEGIN
  IF EXISTS (SELECT FROM pg_roles WHERE rolname IN ('hs_app', 'hs_migrator')
    AND (rolsuper OR rolcreatedb OR rolcreaterole OR rolreplication OR rolbypassrls OR NOT rolcanlogin))
    OR EXISTS (SELECT FROM pg_auth_members WHERE member IN
      (SELECT oid FROM pg_roles WHERE rolname IN ('hs_app', 'hs_migrator')))
    OR EXISTS (SELECT FROM pg_database WHERE datdba IN
      (SELECT oid FROM pg_roles WHERE rolname IN ('hs_app', 'hs_migrator')))
    OR EXISTS (SELECT FROM pg_namespace WHERE nspowner = 'hs_app'::regrole)
    OR EXISTS (SELECT FROM pg_class WHERE relowner = 'hs_app'::regrole)
  THEN RAISE EXCEPTION 'Roles existentes incompatibles; revisar sin reemplazar'; END IF;
  IF to_regnamespace('horas_sociales') IS NULL OR to_regclass('public.goose_db_version') IS NULL
  THEN RAISE EXCEPTION 'Aplicar primero la migración fundacional'; END IF;
END $$;
ALTER SCHEMA horas_sociales OWNER TO hs_migrator;
ALTER TABLE public.goose_db_version OWNER TO hs_migrator;
REVOKE ALL ON SCHEMA horas_sociales FROM PUBLIC;
REVOKE ALL ON SCHEMA horas_sociales FROM hs_app;
GRANT USAGE ON SCHEMA horas_sociales TO hs_app;
GRANT USAGE ON SCHEMA public TO hs_migrator;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA horas_sociales TO hs_app;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA horas_sociales TO hs_app;
ALTER DEFAULT PRIVILEGES FOR ROLE hs_migrator IN SCHEMA horas_sociales
  GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO hs_app;
ALTER DEFAULT PRIVILEGES FOR ROLE hs_migrator IN SCHEMA horas_sociales
  GRANT USAGE, SELECT ON SEQUENCES TO hs_app;
SELECT format('GRANT CONNECT ON DATABASE %I TO hs_app, hs_migrator', current_database()) \gexec
DO $$
BEGIN
  IF has_schema_privilege('hs_app', 'public', 'CREATE')
    OR has_schema_privilege('hs_app', 'horas_sociales', 'CREATE')
    OR has_database_privilege('hs_app', current_database(), 'CREATE')
  THEN RAISE EXCEPTION 'Privilegios heredados incompatibles con la cuenta limitada'; END IF;
END $$;
COMMIT;
