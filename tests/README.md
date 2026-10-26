# Comprobaciones de inicialización

- `bash scripts/check.sh`: formato, módulos, vet, pruebas Go con detección de carreras, compilaciones Go/TypeScript/Vite y HTTP real por proxy. No detiene servicios.
- `bash scripts/check-integration.sh`: repite migración, caída/recuperación y recreación con persistencia; interrumpe temporalmente `hs-dev`, conserva el volumen y limpia solo su esquema de prueba.
- `bash scripts/check-runtime.sh`: construye las etapas runtime con etiquetas `VERSION-runtime-check`, inicia API/Nginx sin puertos en un proyecto local temporal y usa el PostgreSQL existente. Verifica usuarios sin root y el proxy real; retira esos contenedores al terminar.

- `bash scripts/check-security.sh`: npm audit y govulncheck 1.9.0 con dependencias temporales dentro de Docker. Requiere Internet y no corrige ni modifica archivos automáticamente.

Todos requieren inicialización y servicios `hs-dev` disponibles. No ejecutar pruebas que detienen servicios en paralelo. Los contenedores de pruebas no crean tablas de negocio.

La interfaz, la caída/recuperación y la recarga se verificaron también con navegador real; véase `docs/VALIDACION.md`.

- `bash scripts/check-gitflow.sh`: reglas de rama/issue/PR, SemVer, publicación idempotente y rechazo de ejecuciones obsoletas.
- `scripts/ci.sh`: orquestación completa para el runner temporal de GitHub Actions; termina sus contenedores al salir.
