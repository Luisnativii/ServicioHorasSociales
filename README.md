# Portal de Horas Sociales

Inicialización local de Go, React y PostgreSQL. Ejecutada y verificada en Mac ARM64 con Docker Desktop. Solo necesitas Git, Docker con Compose y Bash; Go, Node, npm y PostgreSQL se ejecutan dentro de contenedores.

## GitFlow y colaboración

El código está directamente en la raíz de `develop`, rama predeterminada. `main` conserva únicamente un README. La base inicial se etiqueta `V0.0.1`; los siguientes incrementos se crean automáticamente tras integrar PR que pasen CI.

Empieza por [la guía de issues, ramas, PR y versiones](docs/GITFLOW.md). Una tarea se transforma en `tipo/NUMERO-descripcion` al marcar su issue con `status:ready`.

Para continuar con OpenCode u otra IA, usa [la guía de relevo y próximos 20 issues](docs/Planeacion/RELEVO_OTRA_IA.md). Incluye el estado real, decisiones del propietario, actualización desde GitHub y entrega por PR sin alterar el flujo.

## Preparar y arrancar

En macOS abre Docker Desktop. En Ubuntu inicia Docker Engine con el plugin Compose; en Windows usa Bash dentro de Ubuntu/WSL2 con la integración de Docker Desktop activada. Trabaja desde la raíz de este repositorio, sin `sudo` si Docker ya está accesible al usuario.

```bash
docker version
docker compose version
git --version
bash scripts/init-local.sh
export APP_VERSION="$(bash scripts/version.sh)"
bash scripts/dc.sh --profile tools build api web migrator
bash scripts/dc.sh up -d --wait --wait-timeout 180
bash scripts/migrate.sh dev
bash scripts/dc.sh up -d --wait --wait-timeout 180 db api web
bash scripts/dc.sh --profile tools run --rm migrator status
```

Abre [http://localhost:8080](http://localhost:8080). Backend y PostgreSQL deben indicar «Disponible». Solo la web publica `127.0.0.1:8080`; la API y la base no publican puertos en el anfitrión.

`init-local.sh` genera `.env` con permisos 600 y contraseña aleatoria, sin imprimirla ni sobrescribir secretos existentes. Genera los lockfiles solo si faltan; ambos están incluidos para instalaciones reproducibles. No copies la contraseña de ejemplo. Si Docker no responde, el script informa el bloqueo antes de cambiar archivos.

Usa `bash scripts/dc.sh` para operar el desarrollo: obtiene la versión y, después de provisionar roles, incorpora las credenciales dedicadas. Así las recreaciones conservan la cuenta limitada de la API. Para que tome un tag nuevo, retira cualquier `APP_VERSION` exportada manualmente en la terminal.

## Estado, logs y comprobaciones

```bash
bash scripts/dc.sh ps
bash scripts/dc.sh logs --tail=100 api web db
bash scripts/dc.sh logs -f api
bash scripts/check.sh
bash scripts/check-runtime.sh
bash scripts/check-security.sh
```

Los scripts `check*.sh` de aplicación usan exclusivamente **hs-test**. Antes de ejecutarlos, prepara ese entorno con los comandos de la sección siguiente. `check.sh` verifica módulos, formato Go, `go vet`, carreras, compilación Go/TypeScript/Vite y HTTP por el proxy de pruebas. `check-runtime.sh` comprueba API/Nginx sin root con PostgreSQL de pruebas; `check-security.sh` consulta npm audit y govulncheck 1.9.0, sin modificar lockfiles. `Ctrl+C` termina el seguimiento de logs de desarrollo.

## Pruebas aisladas de desarrollo

No necesitas arrancar `hs-dev` ni crear su `.env` para probar. Desde la raíz:

```bash
bash scripts/init-test.sh
bash scripts/dc-test.sh build api web migrator
bash scripts/dc-test.sh up -d --wait --wait-timeout 180 db api web
bash scripts/migrate.sh test
bash scripts/dc-test.sh up -d --wait --wait-timeout 180 db api web
bash scripts/check-test-isolation.sh
bash scripts/check.sh
bash scripts/check-integration.sh
bash scripts/check-runtime.sh
bash scripts/check-security.sh
# Detener o retirar exclusivamente los contenedores de pruebas; conserva datos.
bash scripts/dc-test.sh stop
bash scripts/dc-test.sh --runtime down
bash scripts/dc-test.sh down
```

`init-test.sh` genera una sola vez `.secrets/hs-test.env` con permisos 600 y contraseña aleatoria propia; conserva los secretos existentes y no lee ni modifica `.env`. `dc-test.sh` fija proyecto `hs-test`, base `horas_sociales_test`, redes `hs-test_app`/`hs-test_data` y volumen `hs-test_pgdata`. No publica puertos; el HTTP se comprueba desde el contenedor web. Las imágenes de herramientas llevan sufijo `-test`, distinto de `-dev`; los montajes de fuentes son de solo lectura y Go usa el UID/GID registrados al inicializar.

`check-integration.sh` detiene/recrea **solo hs-test**, comprueba caída, recuperación, migración y persistencia y elimina únicamente su esquema temporal. `check-runtime.sh` usa `hs-test-runtime` conectado a `hs-test_data`; retira API/Nginx al terminar. Ejecuta las comprobaciones en secuencia, sin dos sesiones simultáneas sobre el mismo entorno de pruebas.

El wrapper rechaza variables exportadas de Compose/base de datos, claves ausentes o duplicadas, destinos inesperados y opciones que sustituyan proyecto, archivos o secretos. Si informa una configuración ambigua, retira las variables indicadas en esa terminal; no cambies credenciales existentes ni borres volúmenes. Los comandos de limpieza conservan `hs-test_pgdata` y nunca actúan sobre `hs-dev_pgdata`.

`bash scripts/ci.sh` ejecuta la secuencia completa anterior, incluidas reglas GitFlow y comprobación de aislamiento, y retira al terminar solo los contenedores/redes de pruebas. Puede ejecutarse junto al portal de desarrollo; no inicia ni detiene `hs-dev`.

## Migraciones

```bash
bash scripts/migrate.sh dev
bash scripts/dc.sh up -d --wait --wait-timeout 180 db api web
bash scripts/dc.sh --profile tools run --rm migrator status
```

`migrate.sh dev` reconstruye el migrator, aplica solo la migración fundacional con la cuenta original cuando hace falta, provisiona roles dedicados y ejecuta las demás migraciones como `hs_migrator`. Repetirlo conserva versiones y cuentas. Goose registra su versión en `public.goose_db_version`; las migraciones fundacional/de identidad rechazan retrocesos para conservar datos.

La migración 2 crea únicamente `horas_sociales.administrador` del contrato aceptado: UUID, correo canónico único, hash Argon2id, estado y fechas técnicas. No crea cuentas, credenciales de ejemplo, sesiones ni tablas de estudiantes/proyectos/horas. La herramienta de #23 calcula hashes reales; todavía no existe login web.

Las consultas se generan desde SQL mediante `bash scripts/sqlc.sh generate` (sqlc 1.31.1 fijado por digest, sin instalar Go/Node en el anfitrión). `bash scripts/sqlc.sh vet` valida las consultas; CI comprueba que regenerar no cambia los archivos versionados. El repositorio de administradores usa esas consultas y no incluye endpoints públicos.

## Separar cuentas de aplicación y migraciones

Después del primer arranque y de aplicar la migración inicial, prepara cuentas dedicadas sin reemplazar `.env` ni reiniciar el clúster:

```bash
bash scripts/provision-db-roles.sh dev
bash scripts/dc.sh up -d --wait --wait-timeout 180 db api web
bash scripts/dc.sh run --rm migrator status
```

`hs_app` permite consultar y modificar filas de `horas_sociales`, sin crear/alterar tablas ni acceder al registro de Goose. `hs_migrator` es propietario del esquema y del registro de migraciones, sin privilegios de superusuario, creación de bases o administración de roles. Los privilegios por defecto se aplican a las tablas/secuencias que cree el migrator; las migraciones deben usar esa cuenta. La cuenta original se conserva para provisión y recuperación técnica, y PostgreSQL sigue sin publicar puertos.

El script genera una sola vez `.secrets/hs-dev-db-roles.env` (permisos 600). Tras verificar las dos conexiones crea su marcador `.ready`; `scripts/dc.sh` carga entonces esas credenciales. El siguiente `up` recrea la API con la cuenta limitada. Antes de la primera provisión, las conexiones originales se usan únicamente para el arranque fundacional; no desarrollar negocio sin completar este paso. Para operar usa el wrapper, que incorpora ambos archivos de configuración; los comandos Compose completos de arranque inicial no cargan el archivo adicional de roles.

La provisión es repetible: no modifica contraseñas de roles ya existentes, no borra tablas y no crea tablas de negocio. Si una cuenta existente tiene privilegios incompatibles o no coincide con el archivo conservado, se detiene. No regeneres ese archivo ni borres el volumen: conserva `.env`, los archivos de roles y sus marcadores al mover el entorno; restaura su copia original por un canal seguro. Si falta el marcador pero los roles y secretos originales existen, repite la provisión. Para diagnosticar errores usa la cuenta original dentro del contenedor PostgreSQL, sin publicar contraseñas, cadenas de conexión ni logs de consultas sensibles. Una corrección de privilegios fuera del contrato requiere revisión explícita.

Puedes abrir la consola técnica local sin mostrar su contraseña con `docker exec -it hs-dev-db-1 sh -c 'exec psql -U "$POSTGRES_USER" -d "$POSTGRES_DB"'`. No ejecutes cambios de contraseña o permisos hasta identificar la discrepancia y conservar una copia protegida de la configuración original.

En pruebas se usa `bash scripts/provision-db-roles.sh test`, con un archivo distinto y las cuentas del clúster `hs-test`. `bash scripts/check-db-roles.sh` crea una migración sintética en ese clúster, verifica permisos y API real y retira únicamente sus tablas de prueba. CI ejecuta esta comprobación antes de las pruebas habituales. Nunca la dirige a desarrollo.

## Crear cuentas administradoras desde Docker

Después de preparar la configuración y migraciones, ejecuta desde una terminal interactiva:

```bash
bash scripts/migrate.sh dev
bash scripts/admin.sh create --email tu-correo@tu-dominio
```

Sustituye el correo por el de la cuenta que quieres crear. La contraseña se pide dos veces **sin mostrarla**. Requiere al menos seis caracteres y un número, máximo 128 bytes; no exige mayúsculas ni símbolos ni recorta espacios. No hay contraseña predeterminada ni cuenta creada automáticamente. El correo se normaliza; repetirlo produce conflicto y conserva su contraseña y estado. La misma operación crea la primera cuenta y las adicionales, incluso cuando no hay ninguna activa.

La herramienta construye `hs-admin:VERSION-dev`, un runtime Distroless sin root, sin puertos y conectado solo a la red de datos. Usa la cuenta PostgreSQL limitada `hs_app`; no admite la cuenta propietaria del arranque. El hashing usa Argon2id v19: 64 MiB, tres iteraciones, paralelismo cuatro, salt aleatorio de 16 bytes y resultado de 32 bytes en formato PHC. Las verificaciones comparan el resultado en tiempo constante y rechazan parámetros no admitidos antes de asignar memoria; se acotan a dos operaciones simultáneas por proceso. Referencias: [Argon2 de Go](https://pkg.go.dev/golang.org/x/crypto/argon2), [entrada de terminal de Go](https://pkg.go.dev/golang.org/x/term), [RFC 9106](https://www.rfc-editor.org/rfc/rfc9106.html#section-4).

Para una automatización local, prepara por un medio seguro un archivo bajo `.secrets/` con la contraseña exacta, sin salto de línea, y permisos 0400 o 0600. El wrapper lo monta en solo lectura y usa el UID/GID del operador para leerlo. Se admiten archivos del usuario o de root que el proceso pueda abrir (Docker Desktop puede informar el montaje como root); se rechazan enlaces, archivos públicos, directorios y contenido excesivo. Nunca se imprime el secreto ni se copia a la imagen:

```bash
bash scripts/admin.sh create --email tu-correo@tu-dominio --password-file .secrets/admin-password
```

No se admite contraseña por argumento, variable de entorno ni pipe. El archivo de entrada se conserva: el operador gestiona su custodia. No hay comandos de reset, reactivación o revocación ni altas web. El navegador ofrece acceso en `/login` y confirma la identidad antes de entrar a Archivo. En `/administradores` permite listar, activar y desactivar cuentas, incluida la propia o la última activa, sin revocar sesiones existentes. Los módulos de negocio se incorporan en sus issues. La comprobación pública de servicios se conserva en `/estado`.

## Validación JWT administrativa

Las rutas `/api/v1` quedan protegidas por defecto. Solo se acepta la cookie `hs_access`; no se leen tokens de URL, cuerpos ni cabeceras Authorization. Se verifica HS256, firma, issuer `horas-sociales-api`, audience `horas-sociales-web`, sujeto UUID y claims obligatorios iat/exp con duración exacta de dos horas. La tolerancia de reloj es 30 segundos para emisión y vencimiento. La identidad se transporta en el contexto de la petición. No se consultan sesiones ni se revocan tokens al desactivar una cuenta, conforme al contrato aceptado.

Para preparar la clave local, ejecuta una sola vez (con la imagen web de desarrollo ya construida):

```bash
bash scripts/init-jwt-key.sh
bash scripts/dc.sh up -d --build --wait --wait-timeout 180
```

El script genera 32 bytes aleatorios dentro de Docker en `.secrets/hs-dev-jwt.key`, con permisos 0600 y creación exclusiva. Repetirlo conserva la clave. No la imprime, no modifica `.env` y no crea cuentas. `dc.sh` monta únicamente ese archivo en solo lectura mediante `compose.jwt.yaml`; no monta el directorio completo de secretos en la API. Conservarlo entre arranques mantiene válidos los tokens emitidos. No hay clave predeterminada ni regeneración al iniciar.

Sin clave configurada, salud sigue pública y el grupo privado deniega todo acceso. Si se configura `JWT_KEY_FILE` y el archivo es inválido, la API rechaza el arranque con un error genérico. El archivo debe ser regular, sin enlaces, de 32 bytes exactos y permisos 0400/0600, propiedad del usuario o de root y legible por el proceso. El montaje local respeta el UID del operador; otros entornos deberán montar su clave protegida explícitamente y definir `JWT_KEY_FILE`, sin reutilizar la clave local.

Los módulos futuros se registran con `Options.RegisterPrivate` para heredar JWT y CSRF. Todavía no existen rutas de login (#26). No enviar JWT por JSON ni almacenarlos en localStorage. Referencia de la biblioteca: [validación JWT en Go](https://golang-jwt.github.io/jwt/usage/parse/).

### Cookies, origen y CSRF

`APP_ORIGIN` define un único origen exacto, sin ruta ni barra final. Producción exige HTTPS; HTTP se admite exclusivamente para `localhost` o `127.0.0.1`. Compose de desarrollo usa `http://localhost:8080`; para usar la otra dirección hay que configurar explícitamente el origen y abrir el portal con esa dirección. El origen nunca se infiere de `X-Forwarded-*`. Con clave JWT configurada, un origen inválido impide arrancar. Sin clave, la preparación CSRF responde 503 y las rutas privadas siguen cerradas.

El intercambio implementado es:

1. `GET /api/v1/auth/csrf` devuelve 204 vacío, `Cache-Control: no-store` y cookie `hs_csrf`. Sin JWT prepara una prueba de prelogin; con JWT válido la vincula a ese JWT. No extiende su vencimiento.
2. En cada POST/PUT/PATCH/DELETE, enviar `Origin` igual a `APP_ORIGIN` y `X-CSRF-Token` con el valor de `hs_csrf`. El grupo privado exige primero JWT y después verifica origen, coincidencia cookie/cabecera y firma HMAC-SHA256. La clave CSRF se deriva de la clave JWT con separación de propósito; incluye nonce aleatorio, fecha y vinculación al JWT. No hay tabla de sesiones. Las cookies/cabeceras duplicadas se rechazan.
3. El futuro login (#26) usará `Browser.SetSession`: cookie `hs_access` HttpOnly, Path=/api, SameSite=Strict, Max-Age=7200, sin Domain; rota la prueba CSRF para vincularla al nuevo JWT. La cookie CSRF usa Path=/ para que el frontend pueda leerla y no es HttpOnly. Ambas usan Secure con HTTPS; solo el origen HTTP local aceptado permite omitirlo.
4. Para el futuro logout (#28), incluso tras vencer el JWT, volver a obtener CSRF y enviarlo con Origin. `Browser.ClearSession` elimina ambas cookies con sus mismos atributos/rutas. No revoca copias del JWT ni renueva sesiones. Los helpers ya están disponibles; login/logout todavía no tienen endpoints.

CSRF caduca a las dos horas con 30 segundos de tolerancia de reloj. Una prueba de prelogin no autoriza mutaciones con JWT y una prueba vinculada a un JWT no se puede reutilizar con otro. Una mutación sin prueba válida devuelve 403 sin ejecutar el handler. No hay CORS genérico ni aceptación de Origin ausente, `null` o exterior. La preparación GET admite Origin ausente, pero exige Host configurado y rechaza Origin exterior o `Sec-Fetch-Site: cross-site`. Referencia: [doble envío firmado de OWASP](https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html#signed-double-submit-cookie-recommended).

## Detener, iniciar y conservar datos

```bash
# Detener sin eliminar contenedores ni datos.
bash scripts/dc.sh stop
# Iniciar y esperar comprobaciones reales de disponibilidad.
bash scripts/dc.sh up -d --wait --wait-timeout 180
# Retirar contenedores y redes, conservando el volumen.
bash scripts/dc.sh down
# Recrear con los mismos datos.
bash scripts/dc.sh up -d --wait --wait-timeout 180
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
bash scripts/dc.sh --profile tools build api web migrator
bash scripts/dc.sh up -d --wait --wait-timeout 180
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

La infraestructura de sesión, navegación y acceso del frontend se describe en [apps/web/src/app/README.md](apps/web/src/app/README.md): cliente HTTP, cookies/CSRF, login, layout administrativo y cachés privadas, sin refresh ni reintentos automáticos. Las comprobaciones específicas se ejecutan con `npm run test:session` dentro de Docker y forman parte de `scripts/check.sh`.

| Componente | Versión |
| --- | --- |
| Go | 1.27.3 |
| chi / pgx | 5.5.0 / 5.11.0 |
| golang-jwt/jwt v5 | 5.3.1 |
| x/crypto / x/term | 0.58.0 / 0.47.0 |
| x/text / x/sys (indirectas) | 0.43.0 / 0.49.0 |
| Goose | 3.28.0 |
| PostgreSQL | 18.6 (bookworm) |
| Node / npm | 22.23.4 / 10.9.9 (incluido en la imagen) |
| React / React DOM | 19.3.1 |
| React Router DOM / TanStack React Query | 7.18.5 / 5.105.0 |
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
