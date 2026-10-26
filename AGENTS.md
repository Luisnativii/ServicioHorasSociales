# Instrucciones del repositorio

- Trabajar en la raíz; el código está en `apps/`, no dentro de otra carpeta del proyecto.
- Leer `docs/GITFLOW.md` y el issue solicitado antes de implementar una tarea.
- `main` está reservada y solo contiene README. Usar `develop` para integración.
- Cada trabajo posterior a la base V0.0.1 debe usar un issue y su rama `tipo/NUMERO-descripcion` desde `develop`; usar la rama creada por `status:ready` si ya existe.
- El PR apunta a `develop`, tiene título convencional terminado en `(#NUMERO)`, incluye `Closes #NUMERO` y una sola etiqueta `version:patch|minor|major`.
- No integrar el PR hasta que el usuario lo solicite y pasen los checks requeridos. No forzar pushes ni mover tags publicados.
- Los tags `VMAJOR.MINOR.PATCH` son la fuente de verdad. El paquete npm es privado y su versión inicial no sustituye los tags. Usar `scripts/version.sh` y `scripts/dc.sh` para construir con la versión correcta.
- Compilación, instalación de dependencias y ejecución de Go, Node y PostgreSQL dentro de Docker. No requerirlos instalados en el anfitrión.
- Ejecutar comprobaciones apropiadas al cambio. GitHub CI ejecuta las comprobaciones completas con `scripts/ci.sh` en su runner temporal.
- Conservar `.env` y el volumen `hs-dev_pgdata`; no imprimir secretos, reemplazar credenciales existentes ni borrar volúmenes para resolver errores.
- Conservar avances y documentos. No implementar autenticación, tablas ni reglas de negocio que todavía no estén aprobadas.
- No inventar identidad Git ni incluir archivos ajenos a la tarea. Revisar el diff y excluir secretos antes de cada commit.

La publicación inicial solicitada por el usuario crea directamente `develop` y V0.0.1 como única excepción de arranque; los siguientes cambios siguen el flujo de issues y PR.
