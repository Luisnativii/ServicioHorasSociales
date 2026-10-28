# Portal de Horas Sociales

Inicialización local de Go, React y PostgreSQL. Ejecutada y verificada en Mac ARM64 con Docker Desktop. Solo necesitas Git, Docker con Compose y Bash; Go, Node, npm y PostgreSQL se ejecutan dentro de contenedores.

## GitFlow y colaboración

El código está directamente en la raíz de `develop`, rama predeterminada. `main` conserva únicamente un README. La base inicial se etiqueta `V0.0.1`; los siguientes incrementos se crean automáticamente tras integrar PR que pasen CI.

Empieza por [la guía de issues, ramas, PR y versiones](docs/GITFLOW.md). Una tarea se transforma en `tipo/NUMERO-descripcion` al marcar su issue con `status:ready`.

## Preparar y arrancar

En macOS abre Docker Desktop. En Ubuntu inicia Docker Engine con el plugin Compose; en Windows usa Bash dentro de Ubuntu/WSL2 con la integración de Docker Desktop activada. Trabaja desde la raíz de este repositorio, sin `sudo` si Docker ya está accesible al usuario.

```bash
docker version
docker compose version
git --version
bash scripts/init-local.sh
export APP_VERSION="$(bash scripts/version.sh)"
docker compose -p hs-dev -f compose.yaml -f compose.dev.yaml --profile tools build api web migrator
docker compose -p hs-dev -f compose.yaml -f compose.dev.yaml up -d --wait --wait-timeout 180
docker compose -p hs-dev -f compose.yaml -f compose.dev.yaml --profile tools run --rm migrator up
docker compose -p hs-dev -f compose.yaml -f compose.dev.yaml --profile tools run --rm migrator status
```

Abre [http://localhost:8080](http://localhost:8080). Backend y PostgreSQL deben indicar «Disponible». Solo la web publica `127.0.0.1:8080`; la API y la base no publican puertos en el anfitrión.

`init-local.sh` genera `.env` con permisos 600 y contraseña aleatoria, sin imprimirla ni sobrescribir secretos existentes. Genera los lockfiles solo si faltan; ambos están incluidos para instalaciones reproducibles. No copies la contraseña de ejemplo. Si Docker no responde, el script informa el bloqueo antes de cambiar archivos.

En cada terminal nueva, ejecuta `export APP_VERSION="$(bash scripts/version.sh)"` antes de usar los comandos Compose completos para tomar la versión del tag actual. El atajo `bash scripts/dc.sh` obtiene y exporta esa versión automáticamente y equivale a `docker compose -p hs-dev -f compose.yaml -f compose.dev.yaml`. Los comandos siguientes muestran la forma completa.

## Estado, logs y comprobaciones

```bash
docker compose -p hs-dev -f compose.yaml -f compose.dev.yaml ps
docker compose -p hs-dev -f compose.yaml -f compose.dev.yaml logs --tail=100 api web db
docker compose -p hs-dev -f compose.yaml -f compose.dev.yaml logs -f api
bash scripts/check.sh
bash scripts/check-runtime.sh
bash scripts/check-security.sh
```

`check.sh` ejecuta dentro de Docker verificación de módulos, formato Go, `go vet`, pruebas con detector de carreras, compilación Go, TypeScript, build de Vite y respuestas HTTP a través del proxy. Requiere el entorno iniciado y PostgreSQL disponible. `Ctrl+C` termina el seguimiento de logs. `check-runtime.sh` valida las etapas runtime sin root en un proyecto temporal sin puertos. `check-security.sh` consulta npm audit y la base oficial de vulnerabilidades Go mediante govulncheck 1.9.0; requiere acceso a Internet y no modifica los lockfiles.

Para repetir las pruebas de migración, caída, recuperación y persistencia:

```bash
bash scripts/check-integration.sh
```

Esta prueba interrumpe temporalmente el portal, detiene PostgreSQL y recrea los contenedores de `hs-dev`. Crea un esquema exclusivo de prueba, verifica que su dato sobrevive y retira únicamente ese esquema al terminar. Conserva el volumen y vuelve a dejar web/API/db iniciados. No ejecutarla mientras otra persona trabaja en el portal local.

## Migraciones

```bash
docker compose -p hs-dev -f compose.yaml -f compose.dev.yaml --profile tools run --rm migrator up
docker compose -p hs-dev -f compose.yaml -f compose.dev.yaml --profile tools run --rm migrator status
```

Repetir `up` no duplica el esquema ni la versión aplicada. Solo se crea el esquema `horas_sociales`; Goose registra su versión en `public.goose_db_version`. La migración fundacional rechaza retrocesos para conservar datos. No existen tablas de estudiantes, proyectos o movimientos de horas.

## Detener, iniciar y conservar datos

```bash
# Detener sin eliminar contenedores ni datos.
docker compose -p hs-dev -f compose.yaml -f compose.dev.yaml stop
# Iniciar y esperar comprobaciones reales de disponibilidad.
docker compose -p hs-dev -f compose.yaml -f compose.dev.yaml up -d --wait --wait-timeout 180
# Retirar contenedores y redes, conservando el volumen.
docker compose -p hs-dev -f compose.yaml -f compose.dev.yaml down
# Recrear con los mismos datos.
docker compose -p hs-dev -f compose.yaml -f compose.dev.yaml up -d --wait --wait-timeout 180
```

El volumen `hs-dev_pgdata` está montado en `/var/lib/postgresql`, como requiere PostgreSQL 18. Eliminar volúmenes es una operación distinta que destruye datos y no forma parte del procedimiento normal. Mantén el nombre `hs-dev`. Cambiar `.env` no cambia las credenciales de un clúster ya creado: no borres el volumen para resolver una contraseña incorrecta.

## Recursos y operación local

Los contenedores tienen límites de CPU, memoria y procesos y rotación de logs. Son valores iniciales para desarrollo, comprobados en Docker Desktop con aproximadamente 4 GiB de memoria y 8 CPU disponibles; no constituyen metas de rendimiento ni requisitos aprobados del servidor institucional.

| Servicio | CPU máxima | Memoria máxima | Procesos máximos |
| --- | --- | --- | --- |
| PostgreSQL | 1 | 512 MiB | 256 |
| API y herramientas Go | 2 | 1 GiB | 512 |
| Web y herramientas Node | 1 | 512 MiB | 256 |
| Migrator | 0.5 | 256 MiB | 128 |

Los límites no reservan esa memoria. API conserva margen para compilar con recarga y ejecutar pruebas; web incluye Vite/TypeScript; PostgreSQL mantiene el pool reducido existente. Un build de imagen usa BuildKit y no queda limitado por los valores de ejecución del servicio. Con varios entornos simultáneos hay que considerar su consumo conjunto.

Para ajustar un servicio utiliza `HS_DB_CPUS`, `HS_DB_MEMORY`, `HS_DB_PIDS` y los equivalentes `HS_API_*`, `HS_WEB_*` o `HS_MIGRATOR_*`. No es necesario cambiar credenciales ni regenerar `.env`. Por ejemplo:

```bash
HS_API_MEMORY=1536m bash scripts/dc.sh up -d --wait api
```

Mantén el mismo valor en las siguientes recreaciones, exportándolo en tu terminal o agregándolo intencionalmente a la configuración local. Los logs usan `json-file`, archivos de hasta 10 MiB y tres archivos por contenedor; se ajustan con `HS_LOG_MAX_SIZE` y `HS_LOG_MAX_FILES`. Los logs no son un respaldo de datos.

PostgreSQL, API y web usan `restart: unless-stopped`: Docker los reinicia si terminan inesperadamente, pero respeta una parada manual. El migrator no se reinicia automáticamente. Un estado `unhealthy` por sí solo no reinicia un contenedor; consulta sus logs y readiness para diagnosticarlo.

El apagado permite 30 s a PostgreSQL y 15 s a API/web; la API dispone de 10 s internos para terminar peticiones. Usa `stop`, `up` y `down` como se indica arriba: conserva el volumen, sin eliminar datos. API y PostgreSQL siguen sin publicar puertos. Parámetros según la [referencia oficial de servicios Compose](https://docs.docker.com/reference/compose-file/services/).

## Desarrollo con recarga

- React y estilos: edita `apps/web/src/`; Vite actualiza el navegador automáticamente. También están montados `index.html` y `vite.config.ts`.
- Go: edita `apps/api/cmd/` o `apps/api/internal/`; el observador detecta el contenido, compila en `/tmp`, apaga el proceso anterior y arranca el nuevo. Si una edición no compila, conserva la última API funcional y muestra el error en logs. El intervalo de 1 s solo observa archivos; no sustituye los healthchecks.
- Después de cambiar dependencias, Dockerfiles o Compose, reconstruye las imágenes y recrea los servicios. Consulta [las decisiones técnicas](docs/DECISIONES.md) para límites y permisos.

```bash
# Actualizar dependencias Go solo tras editar go.mod intencionalmente.
docker run --rm --user "$(id -u):$(id -g)" -e HOME=/tmp -e GOPATH=/tmp/go -e GOCACHE=/tmp/cache -v "$PWD/apps/api:/src" -w /src golang:1.27.3-bookworm go mod tidy
# Actualizar el lockfile web solo tras editar package.json intencionalmente.
docker run --rm --user "$(id -u):$(id -g)" -e HOME=/tmp -v "$PWD/apps/web:/src" -w /src node:22.23.4-bookworm-slim npm install --package-lock-only --ignore-scripts
docker compose -p hs-dev -f compose.yaml -f compose.dev.yaml --profile tools build api web migrator
docker compose -p hs-dev -f compose.yaml -f compose.dev.yaml up -d --wait --wait-timeout 180
```

## Estructura

```text
apps/api/
  cmd/api/                       Inicio y cierre de la API
  internal/config/               Configuración y pruebas
  internal/platform/database/    Pool PostgreSQL
  internal/platform/http/        Rutas, middleware y pruebas
  internal/modules/              Reservado para módulos aprobados
  db/migrations/                 Migración del esquema
  db/queries/                    Reservado para consultas
apps/web/src/
  app/                           Pantalla principal
  components/                    Reservado para componentes compartidos
  features/                      Reservado para módulos
  lib/                           Solicitudes HTTP
api/                             Contrato HTTP
deploy/docker/                   Dockerfiles multietapa
deploy/nginx/                    Proxy y frontend compilado
scripts/                         Inicialización, operación y pruebas
docs/                            Decisiones, evidencia y primer commit
tests/                           Alcance de integración
```

## Versiones fijadas

| Componente | Versión |
| --- | --- |
| Go | 1.27.3 |
| chi / pgx | 5.5.0 / 5.11.0 |
| x/text (indirecta) | 0.41.0 |
| Goose | 3.28.0 |
| PostgreSQL | 18.6 (bookworm) |
| Node / npm | 22.23.4 / 10.9.9 (incluido en la imagen) |
| React / React DOM | 19.3.1 |
| Vite / plugin React | 7.3.7 / 5.2.0 |
| TypeScript | 5.9.3 |
| Tailwind / plugin Vite | 4.3.3 |
| Tipos React / React DOM | 19.3.0 |
| Nginx sin root | 1.30.3-alpine |
| Distroless | static-debian12:nonroot, digest fijado en Dockerfiles |

Se verificaron registros oficiales y compatibilidad, se construyeron las imágenes ARM64 y se compilaron los proyectos. Fuentes: [Go](https://go.dev/doc/devel/release), [Node](https://nodejs.org/en/blog/vulnerability/october-2026-security-releases), [PostgreSQL](https://www.postgresql.org/docs/release/18.6/), [chi](https://github.com/go-chi/chi/releases), [pgx](https://github.com/jackc/pgx/releases), [Goose](https://github.com/pressly/goose/releases), [Vite](https://vite.dev/guide/). Las dependencias web exactas están fijadas en `package-lock.json`.

Las imágenes locales son `hs-api:0.0.1-dev`, `hs-web:0.0.1-dev` y `hs-migrator:0.0.1-dev`. Las etapas runtime funcionan sin root y sirven el frontend compilado con Nginx; `compose.yaml` las define sin publicar puertos. Esta etapa no incluye despliegue externo.

```bash
# Construcción de comprobación de los runtime, sin sobrescribir las imágenes dev.
docker build -f deploy/docker/api.Dockerfile --target runtime --build-arg APP_VERSION="$APP_VERSION" -t "hs-api:$APP_VERSION-runtime-check" .
docker build -f deploy/docker/web.Dockerfile --target runtime --build-arg APP_VERSION="$APP_VERSION" -t "hs-web:$APP_VERSION-runtime-check" .
```

## Validación y versionamiento

Consulta [la evidencia ejecutada](docs/VALIDACION.md) y [la guía de GitFlow](docs/GITFLOW.md). El repositorio utiliza la identidad Git ya configurada y la base inicial `V0.0.1` en `develop`. `.env` está excluido de Git. Los tags `VMAJOR.MINOR.PATCH` son la fuente de verdad; `scripts/dc.sh` deriva su versión y permite `APP_VERSION` en CI. Las etiquetas `0.0.1-dev` son las de la base; aumentan con los tags. No hay publicación de imágenes ni despliegue externo.

Validación manual pendiente:

- Abrir el portal y confirmar que el texto y los estilos se ven bien en tu navegador y en vista móvil.
- Pulsar «Volver a comprobar» y verificar ambos servicios disponibles.
- Revisar los archivos y la evidencia; registrar la validación manual en su issue antes de iniciar módulos de negocio.

La inicialización termina aquí: no incluye autenticación, formularios administrativos, dashboard, Cloudflare, Portainer ni despliegue en producción.
