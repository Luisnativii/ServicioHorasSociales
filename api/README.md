# API de inicialización

Las rutas conservan `/api` tanto en el proxy de Vite como en Nginx.

| Método y ruta | Respuesta | Dependencia |
| --- | --- | --- |
| `GET /api/health/live` | `200 {"status":"ok"}` | Proceso HTTP |
| `GET /api/health/ready` | `200 {"status":"ok"}` | Ping PostgreSQL, plazo máximo de 2 s |
| `GET /api/health/ready` | `503 {"status":"unavailable"}` | PostgreSQL no disponible |

Todos los cuerpos son JSON con `Content-Type: application/json; charset=utf-8` y `Cache-Control: no-store`.
Una ruta desconocida devuelve `404 {"status":"not_found"}`. Un método no permitido devuelve `405 {"status":"method_not_allowed"}`. Los errores internos devuelven un mensaje genérico sin cadenas de conexión.

Readiness verifica la conexión y, con `CHECK_CORE_CATALOG=true`, el acceso a las tablas de facultades y carreras del core. Ejecutar Goose sigue siendo un paso explícito.

## Contrato común de /api/v1

[OpenAPI](openapi.yaml) describe los endpoints y su contrato. Las rutas desconocidas dentro de `/api/v1` usan este formato:

```json
{"error":{"code":"not_found","message":"No se encontró el recurso.","request_id":"identificador-de-la-petición"}}
```

`X-Request-ID` permite correlacionar la respuesta. Fuera de `/api/v1` se conservan las respuestas iniciales.

Los handlers futuros reutilizan `ReadJSON`, `WriteJSON`, `WriteError`, `WriteInputError` y `ParsePagination`, en `internal/platform/http`. `ReadJSON` exige `application/json`, un único objeto y campos conocidos; rechaza cuerpos superiores a 1 MiB, incluidos espacios finales. La validación funcional de los campos pertenece al módulo aprobado, no al decodificador. Los errores públicos usan un vocabulario cerrado y nunca el texto de errores de SQL/decodificación.

La paginación usa `page=1`, `page_size=25`, máximo 100; rechaza duplicados, vacíos, rangos inválidos y overflow del offset. Cada módulo proporciona su mapa de nombres públicos a columnas fijas y un orden por defecto estable (con desempate por identidad al construir su consulta). `order` admite únicamente `asc` o `desc`. El repositorio usa parámetros para valores, filtros, límite y offset: nunca concatena el nombre recibido del cliente en SQL. El helper no ejecuta consultas ni expone listados ficticios.

Comprobación dentro del entorno de pruebas preparado:

```bash
bash scripts/dc-test.sh run --rm --no-deps api go test -race ./internal/platform/http
```

## Catálogo académico

El portal consulta directamente las facultades y carreras del core, con una sesión administrativa vigente. El género corresponde al expediente de estudiantes (#55).

| Ruta de solo lectura | Uso |
| --- | --- |
| `GET /api/v1/catalogos/facultades` | Lista las facultades activas. |
| `GET /api/v1/catalogos/carreras` | Lista carreras activas de facultades activas. El filtro `facultad_id` restringe el resultado a una facultad del core. |
| `GET /api/v1/catalogos/facultades/{id}` | Consulta por UUID, incluidas facultades inactivas. |
| `GET /api/v1/catalogos/carreras/{id}` | Consulta por UUID, con el identificador y estado de su facultad. |

Ambos listados aceptan `q` para buscar por nombre sin distinguir mayúsculas ni tildes, `page` y `page_size` para paginar y `incluir_inactivas=true` para consultar el historial. El orden es ascendente por nombre normalizado, con desempate por UUID. `pagina.total` cuenta las coincidencias antes de paginar.

## Catálogo territorial

Las tres rutas de lectura bajo `/api/v1/catalogos` requieren la cookie JWT de administrador.

| Ruta GET | Filtro obligatorio |
| --- | --- |
| `/departamentos` | Ninguno |
| `/municipios` | `departamento_id` |
| `/distritos` | `municipio_id` |

Se aceptan `page` y `page_size` (25 por omisión, máximo 100). La respuesta contiene `lugares` y `pagina`. Cada lugar incluye `id`, `codigo` y `nombre`. Los municipios y distritos incluyen también `padre_id`. Los listados se ordenan por código y UUID. Un filtro ausente o con UUID inválido devuelve 400. Un UUID válido sin coincidencias devuelve 200 con un listado vacío.

La API lee los UUID canónicos de las tres tablas `public.geo_*`. Los nombres libres heredados conservan su tratamiento de legado. El rol de la API necesita `SELECT` sobre esas tablas, como configura `scripts/start-core-local.sh` para el entorno local. La carga del catálogo se ejecuta de forma explícita con el [procedimiento de inventario y carga](../docs/CATALOGO_TERRITORIAL_CORE.md).

## Catálogos de clasificación

Tipos de proyecto y áreas se administran por separado. Usa la base `/api/v1/catalogos/tipos-proyecto` o `/api/v1/catalogos/areas`.

| Método | Ruta relativa | Resultado |
| --- | --- | --- |
| GET | Base | Lista opciones activas, ordenadas por nombre e ID. |
| POST | Base | Crea una opción activa con `{"nombre":"…"}`. |
| GET | `/{id}` | Consulta una opción por UUID, incluso si está inactiva. |
| PATCH | `/{id}` | Cambia el nombre con `{"nombre":"…"}`. |
| POST | `/{id}/desactivar` | Marca la opción como inactiva sin eliminarla. |

Las rutas requieren JWT y las mutaciones exigen origen válido y CSRF. Los nombres se recortan y admiten hasta 200 caracteres. Los duplicados dentro del mismo catálogo devuelven 409 aunque cambie la capitalización. La respuesta de una opción contiene `id`, `nombre` y `activo`.

El listado devuelve `opciones` y `pagina`. Acepta `q` como fragmento literal del nombre sin distinguir mayúsculas, `incluir_inactivos=true`, `page` y `page_size` (25 por omisión y máximo de 100). La desactivación conserva el UUID y las referencias existentes. La API opera sobre los catálogos del portal en `horas_sociales`.

## Coordinador de carrera

`GET /api/v1/catalogos/carreras/{id}/coordinador` consulta el nombre opcional mediante el UUID canónico de la carrera. Devuelve `carrera_id` y `coordinador_nombre`, que es `null` si la carrera existe y el dato está vacío. Las carreras inactivas siguen disponibles para lectura.

`PUT` en la misma ruta acepta `{"coordinador_nombre":"…"}` o `{"coordinador_nombre":null}` para retirar el nombre. Requiere JWT, origen válido y CSRF. El nombre admite hasta 200 caracteres después de recortar espacios exteriores. Un UUID inválido devuelve 400 y una carrera inexistente, 404.

La API verifica la carrera al guardar y mantiene como máximo un nombre por UUID en el esquema del portal. Varias carreras pueden tener el mismo nombre. La unión con el catálogo canónico mantiene visibles únicamente carreras existentes y evita otorgar permisos de escritura sobre el core.

## Instituciones

| Método y ruta | Resultado |
| --- | --- |
| `GET /api/v1/instituciones` | Lista instituciones activas por omisión, ordenadas por nombre y UUID. |
| `POST /api/v1/instituciones` | Crea una institución activa. |
| `GET /api/v1/instituciones/{id}` | Consulta por UUID, incluso si está inactiva. |
| `PUT /api/v1/instituciones/{id}` | Reemplaza el nombre y la sede opcional. |
| `POST /api/v1/instituciones/{id}/desactivar` | Desactiva conservando la institución y sus proyectos. |

GET acepta `q` como fragmento literal del nombre o código CSS sin distinguir mayúsculas, `incluir_inactivas=true`, `page` y `page_size` (25 por omisión, máximo de 100). Devuelve `instituciones` y `pagina`.

POST y PUT aceptan `nombre` y el UUID opcional `distrito_id`. Omitir el distrito o enviar `null` deja la institución sin sede canónica. El código CSS y la ubicación heredada son de lectura. Un nombre duplicado devuelve 409 con `error.field=nombre`. Un distrito inválido o inexistente devuelve 400 con `error.field=distrito_id`.

Las rutas requieren JWT. Las mutaciones exigen origen válido y CSRF. Las escrituras usan los permisos por columna del [contrato con el core](../docs/CONTRATO_POSTGRESQL_CORE.md#instituciones). El contacto principal se incorpora en #48.
