# Base compartida en una prueba local

Ejecuta los comandos desde la raíz de ServicioHorasSociales con Docker Engine o Docker Desktop en modo de contenedores Linux, Docker Compose 2.24.4 o posterior y Bash. En Windows puedes usar Git Bash. El repositorio incluye las tres migraciones y Docker compila Go y React.

## Arrancar y entrar

```bash
bash scripts/start-core-local.sh
bash scripts/dc-core-local.sh ps
curl --fail --silent --show-error http://127.0.0.1:18081/api/health/ready
```

Abre [http://127.0.0.1:18081](http://127.0.0.1:18081). Usa esa dirección exacta para que coincida con el origen de las cookies y la protección CSRF. Si el puerto está ocupado, exporta `HS_CORE_LOCAL_PORT` con otro puerto y cambia la dirección que abres.

Para crear una cuenta en una terminal interactiva, indica tu correo de prueba y escribe la contraseña cuando la herramienta la solicite.

```bash
bash scripts/dc-core-local.sh build admin
bash scripts/dc-core-local.sh run --rm --no-deps admin create --email admin@portal.test
```

Si la terminal de Windows no admite el comando interactivo, ejecuta `winpty bash scripts/dc-core-local.sh run --rm --no-deps admin create --email admin@portal.test` desde Git Bash.

El portal disponible incluye inicio de sesión, cierre de sesión y administración de cuentas. Los endpoints de negocio sobre los datos del core se implementarán en sus issues.

## Verificar PostgreSQL

El proyecto Compose se llama `hs-core-local`. Solo tiene un servicio PostgreSQL y una base de aplicación llamada `horas_sociales`.

| Esquema | Contenido |
| --- | --- |
| `public` | Tablas y vistas del core, registro `schema_migrations` y registro de Goose `goose_db_version`. |
| `horas_sociales` | Cuentas administrativas y catálogos propios del portal. |

```bash
bash scripts/dc-core-local.sh exec -T db sh -c 'exec psql -X -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB"' <<'SQL'
SELECT current_database(), current_setting('server_version');
SELECT name, checksum FROM public.schema_migrations ORDER BY name;
SELECT count(*) AS facultades FROM public.faculties;
SELECT count(*) AS carreras FROM public.careers;
SELECT version_id, is_applied FROM public.goose_db_version ORDER BY id;
SELECT to_regclass('horas_sociales.administrador') AS cuentas,
       to_regclass('public.v_student_hours') AS horas;
SELECT has_table_privilege('hs_app', 'public.careers', 'SELECT') AS lee_carreras,
       has_table_privilege('hs_app', 'public.careers', 'UPDATE') AS actualiza_carreras;
SQL
```

Espera tres migraciones del core con SHA-256, tres facultades, 22 carreras, las versiones 1–3 de Goose aplicadas y ambos objetos de consulta visibles. El rol de la API puede leer las carreras. La escritura al catálogo del core requiere su propio acuerdo en #139.

## Detener y reanudar

```bash
bash scripts/dc-core-local.sh stop
bash scripts/start-core-local.sh
```

El arranque reutiliza `hs-core-local_pgdata`, la clave JWT del volumen `hs-core-local_jwt` y los archivos de conexión bajo `.secrets/`. Usa `down` para retirar contenedores y redes conservando los volúmenes. Si falla el arranque, revisa `bash scripts/dc-core-local.sh logs --tail 80 db api web` y conserva los volúmenes para diagnosticarlo.

## Referencia para otra IA

1. Lee `AGENTS.md`, esta guía y el issue de la tarea. Usa exclusivamente `scripts/dc-core-local.sh` para operar este entorno.
2. Ejecuta `bash scripts/start-core-local.sh` y verifica `ps`, readiness y las consultas de ambos esquemas. Si necesitas validar el login, crea una cuenta de prueba con la herramienta administrativa.
3. El SQL en `deploy/core-local/migrations/` es copia byte a byte de `src/database/migrations/` de [horas-sociales-core-backend en 3f8d6a4](https://github.com/jaalgori0/horas-sociales-core-backend/tree/3f8d6a4169b598674b63b4618ced5e4589e96c27/src/database/migrations). Conserva los archivos originales y sus checksums. El entrypoint los ejecuta en una transacción únicamente al inicializar un volumen nuevo.
4. Para cambios futuros del core, coordina la actualización de la revisión y aplica sus nuevas migraciones con su procedimiento. Goose se ocupa de las migraciones del portal. Mantén fuera de Git los secretos, datos personales y respaldos.
5. Este entorno trae el catálogo oficial y arranca con las tablas de negocio vacías. El ETL histórico y la conexión al PostgreSQL de la laptop son tareas posteriores.

Referencia de Compose: [reemplazo de volúmenes con `!override`](https://docs.docker.com/reference/compose-file/merge/#replace-value).
