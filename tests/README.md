# Comprobaciones aisladas

Preparación y comandos exactos en [README](../README.md#pruebas-aisladas-de-desarrollo). Solo se requieren Docker/Compose y Bash; no Go, Node ni PostgreSQL en el anfitrión.

- `scripts/init-test.sh`: secretos propios en `.secrets/hs-test.env`; no modifica `.env` ni regenera credenciales existentes.
- `scripts/dc-test.sh`: proyecto fijo `hs-test`, base `horas_sociales_test`, redes/volumen propios, sin puertos publicados. Rechaza configuración ambigua antes de ejecutar Compose.
- `scripts/check-test-isolation.sh`: verifica destinos, redes, volumen, permisos de montajes y rechazo de sustituciones peligrosas. Procesa la configuración por stdin dentro de Node en Docker sin imprimir secretos.
- `scripts/check.sh`: formato, módulos, vet, pruebas Go con carreras, compilaciones Go/TypeScript/Vite y HTTP real de hs-test.
- `scripts/check-integration.sh`: migración, caída/recuperación y persistencia sobre hs-test. Retira su esquema temporal, conserva volumen y deja web/API/db de pruebas iniciados.
- `scripts/check-runtime.sh`: API/Nginx sin root en `hs-test-runtime`, sin puertos y conectado exclusivamente a PostgreSQL de hs-test. Retira sus contenedores al terminar.
- `scripts/check-security.sh`: npm audit y govulncheck con dependencias temporales en Docker; requiere Internet y no modifica lockfiles.
- `scripts/check-gitflow.sh`: reglas de rama/issue/PR, SemVer y publicación de tags; Node en Docker.
- `scripts/ci.sh`: secuencia completa; prepara hs-test, prueba y retira contenedores/redes de pruebas al salir. Conserva volúmenes.
- `scripts/check-db-roles.sh`: provisiona cuentas dedicadas en hs-test, aplica dos veces una migración sintética con hs_migrator, verifica DML y rechazo de DDL/lectura del registro Goose con hs_app, repite provisión y comprueba readiness de la API con la cuenta limitada. Retira únicamente las tablas sintéticas.
- `scripts/check-admin.sh`: aplica las migraciones dos veces en una base nueva de prueba propiedad del migrator, comprueba ausencia de cuentas automáticas y prueba el repositorio mediante hs_app. Sus filas sintéticas se revierten; solo se elimina la base creada por esa ejecución. No elimina volúmenes ni usa desarrollo.
- `scripts/check-admin-cli.sh`: Argon2id/PHC, política de contraseña, entrada protegida y CLI real con PostgreSQL; múltiples altas, duplicado sin cambios, otra cuenta tras desactivar las anteriores y creación con el runtime Distroless. Retira solo sus propias cuentas y archivo sintético. No crea credenciales predeterminadas en desarrollo.

No ejecutar dos secuencias a la vez sobre hs-test. Desarrollo puede permanecer iniciado: ninguna comprobación detiene ni escribe en su base. No se crean tablas de negocio; los ensayos usan su esquema temporal y la migración fundacional existente.
