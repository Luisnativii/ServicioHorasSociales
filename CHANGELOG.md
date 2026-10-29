# Changelog

## Pendiente de integración

- feat(web): pantalla de acceso Archivo conectada a login y comprobación de identidad (#31), con errores genéricos, protección contra doble envío y rutas de entrada; salud local disponible en /estado.

- feat(web): cliente central de sesión/CSRF, React Router y TanStack Query (#30), sin refresh ni reintentos automáticos; limpieza de datos privados al cambiar o vencer la sesión.

- feat(auth): administrar cuentas ADMIN (#29): GET /api/v1/admin/cuentas (listado con paginación) y POST /api/v1/admin/cuentas/{id}/activar y /desactivar, con JWT y CSRF en mutaciones, registro de actor en logs sin secretos, y rutas documentadas en OpenAPI.
- feat(auth): cerrar sesión eliminando cookies (#28): POST /api/v1/auth/logout con CSRF y origen, borra hs_access y hs_csrf con los mismos atributos de emisión, funciona con JWT vencido sin revocación ni tabla de sesiones, y rutas documentadas en OpenAPI.
- feat(auth): login y consulta de identidad (#26): POST /api/v1/auth/login y GET /api/v1/auth/me con cookies reales hs_access y hs_csrf, verificación Argon2id, errores genéricos, control de intentos por IP con 429 y Retry-After, y rutas documentadas en OpenAPI.

## V0.0.1 — 2026-10-25

- Entorno local Docker validado en macOS ARM64 con proyecto `hs-dev`.
- API Go con chi, pgxpool, configuración validada, salud JSON y apagado ordenado.
- React, TypeScript, Vite y Tailwind con conexión real, carga, errores y reintento.
- PostgreSQL 18, migración Goose idempotente y persistencia con volumen nombrado.
- Imágenes de desarrollo y etapas runtime sin root; proxy Nginx y redes separadas.
- Corregidos GO-2026-6629 y GO-2026-5970 con x/text 0.41.0; revisión npm y Go reproducible.
- Versiones y lockfiles fijados; scripts operativos y pruebas de recuperación/persistencia.

- Publicación de la base en `develop`; `main` reservada con README.
- GitFlow basado en issues, ramas automáticas, PR, squash y checks obligatorios.
- CI y tags SemVer automáticos después de integraciones verificadas.

La numeración de preparación `0.1.0-dev` se normaliza a `V0.0.1` por decisión del propietario. Sin autenticación, tablas de negocio ni despliegue externo.
