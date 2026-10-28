# API de inicialización

Las rutas conservan `/api` tanto en el proxy de Vite como en Nginx.

| Método y ruta | Respuesta | Dependencia |
| --- | --- | --- |
| `GET /api/health/live` | `200 {"status":"ok"}` | Proceso HTTP |
| `GET /api/health/ready` | `200 {"status":"ok"}` | Ping PostgreSQL, plazo máximo de 2 s |
| `GET /api/health/ready` | `503 {"status":"unavailable"}` | PostgreSQL no disponible |

Todos los cuerpos son JSON con `Content-Type: application/json; charset=utf-8` y `Cache-Control: no-store`.
Una ruta desconocida devuelve `404 {"status":"not_found"}`. Un método no permitido devuelve `405 {"status":"method_not_allowed"}`. Los errores internos devuelven un mensaje genérico sin cadenas de conexión.

Readiness verifica conexión, no tablas ni versión de migración. Ejecutar Goose sigue siendo un paso explícito. No hay autenticación ni operaciones de negocio en esta etapa.

## Contrato común de /api/v1

[OpenAPI](openapi.yaml) describe las rutas de salud existentes y los componentes para los módulos siguientes. Las rutas desconocidas dentro de `/api/v1` responden con el nuevo formato:

```json
{"error":{"code":"not_found","message":"No se encontró el recurso.","request_id":"identificador-de-la-petición"}}
```

`X-Request-ID` permite correlacionar la respuesta. Fuera de `/api/v1` se conservan las respuestas iniciales. Todavía no existen endpoints de negocio ni de autenticación.

Los handlers futuros reutilizan `ReadJSON`, `WriteJSON`, `WriteError`, `WriteInputError` y `ParsePagination`, en `internal/platform/http`. `ReadJSON` exige `application/json`, un único objeto y campos conocidos; rechaza cuerpos superiores a 1 MiB, incluidos espacios finales. La validación funcional de los campos pertenece al módulo aprobado, no al decodificador. Los errores públicos usan un vocabulario cerrado y nunca el texto de errores de SQL/decodificación.

La paginación usa `page=1`, `page_size=25`, máximo 100; rechaza duplicados, vacíos, rangos inválidos y overflow del offset. Cada módulo proporciona su mapa de nombres públicos a columnas fijas y un orden por defecto estable (con desempate por identidad al construir su consulta). `order` admite únicamente `asc` o `desc`. El repositorio usa parámetros para valores, filtros, límite y offset: nunca concatena el nombre recibido del cliente en SQL. El helper no ejecuta consultas ni expone listados ficticios.

Comprobación dentro del entorno de pruebas preparado:

```bash
bash scripts/dc-test.sh run --rm --no-deps api go test -race ./internal/platform/http
```
