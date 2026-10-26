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
