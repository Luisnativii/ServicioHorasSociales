# Changelog

## Pendiente de integración

- feat(estudiantes): buscar, crear y editar expedientes del core y consultar la historia académica (#55).

- feat(web): gestionar instituciones, su sede y contacto principal desde el portal privado (#49).

- feat(instituciones): consultar y editar el contacto principal conservando el teléfono y la extensión como texto (#48).

- feat(instituciones): exponer detalle, búsqueda paginada, alta, edición de nombre y sede, y desactivación auditada (#47).

- feat(instituciones): leer instituciones históricas del core y desactivarlas desde el repositorio sin romper vínculos con proyectos (#46).

- feat(web): administrar tipos y áreas de proyecto con búsqueda paginada y consulta de inactivos (#43).

- feat(web): consultar y editar el nombre del coordinador desde el detalle de carrera (#45).

- feat(catalogos): guardar el nombre opcional del coordinador por carrera en el esquema del portal (#44).

- feat(catalogos): administrar tipos de proyecto y áreas con una API protegida, filtros paginados y consulta de inactivos (#42).

- feat(web): consultar la jerarquía territorial desde un selector reutilizable con referencias históricas separadas (#39).

- feat(catalogos): consultar departamentos, municipios y distritos del core con JWT, filtros por padre y paginación (#38).

- chore(db): documentar el catálogo territorial del core (#37) y preparar su inventario y carga idempotente a partir de los códigos del Censo 2024.

- feat(web): consultar el catálogo académico del core (#36) desde la navegación privada, con búsqueda, paginación del servidor, filtro por facultad y detalle de registros activos e inactivos.

- feat(catalogos): consultar facultades y carreras del core por la API privada (#35), con búsqueda, paginación, filtro por facultad y lectura histórica.

- feat(catalogos): persistir tipos de proyecto y áreas independientes (#40), con nombres únicos sin distinguir mayúsculas, desactivación sin borrado, consultas sqlc y pruebas PostgreSQL sin semillas.

- ci: desactivar los workflows de GitHub Actions bloqueados por facturación y trasladar creación de ramas, pruebas completas y versionado a un procedimiento manual verificable.

- feat(web): gestión de cuentas administradoras (#33), listado paginado real, confirmación de activación/desactivación y reconciliación de cambios inciertos sin reintentar escrituras.

- feat(web): estructura administrativa Archivo (#32), navegación protegida, menú móvil, confirmación de salida y limpieza de vistas privadas ante caché de navegación.

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
