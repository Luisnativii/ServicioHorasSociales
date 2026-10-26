# Contexto breve para ejecutar un issue con IA

Este archivo es una instrucción de trabajo acotada, no una orden de implementar todo el backlog.

1. Lee `AGENTS.md`, `docs/GITFLOW.md`, `docs/Planeacion/DECISIONES_VIGENTES.md` y el issue concreto. Después consulta solo las secciones de modelo/reglas y la etapa original necesarias para ese cambio. Evita volver a cargar las 20 etapas en cada sesión.
2. Comprueba dependencias cerradas e integradas en `develop`, decisiones aprobadas y fuentes disponibles. Una rama o PR abierto no cumple la dependencia. Si falta una decisión, informa la pregunta exacta y el bloqueo; no rellenes columnas, catálogos ni políticas por intuición.
3. Usa la rama `tipo/NUMERO-descripcion` creada para el issue. Mantén `main` con su README, conserva trabajo existente y no fuerces el historial ni tags.
4. Stack existente: Go/net/http/chi/pgx, PostgreSQL 18, React/TypeScript/Vite/Tailwind. Añade sqlc, Router, Query, Excelize o Playwright solo en las tareas previstas y con versiones verificadas. Dependencias, compilación y ejecución dentro de Docker.
5. La prioridad de identidad es UN ROL ADMIN con MÚLTIPLES CUENTAS y acceso completo, autenticadas por JWT y sesión revocable. El plan original de una sola cuenta está sustituido. No hay acceso de estudiantes/coordinadores ni registro público por defecto.
6. Implementa exclusivamente el resultado de este issue. Indica primero archivos y comprobaciones relevantes; limita el cambio a su frontera útil. No abras otros módulos ni implementes todos sus dependientes. Si el alcance real resulta mayor, propone un desglose antes de mezclarlo.
7. Mantén SQL parametrizado, contexto/timeouts, integridad y protección de las rutas de negocio. Agrega migraciones nuevas; nunca reescribas las ya aplicadas ni borres volúmenes. Los datos reales y secretos quedan fuera del repo, logs y prompts.
8. Ejecuta pruebas apropiadas en contenedores y PostgreSQL real para integridad/concurrencia. Usa `hs-test` cuando se incorpore el aislamiento; no conectes pruebas destructivas a `hs-dev`. No simules persistencia ni declares pruebas no ejecutadas.
9. Actualiza OpenAPI, documentación y CHANGELOG solo cuando el cambio lo requiera. Entrega un resumen breve, comandos/evidencia, limitaciones reales y criterios para revisión. Mantén el portal ejecutable.
10. Prepara un PR hacia `develop`, título convencional con `(#NUMERO)`, `Closes #NUMERO` y una sola etiqueta de versión. No integres ni crees tags manuales: respeta autorización de merge y pipeline existentes.

Para empezar una sesión basta: «Trabaja únicamente el issue #NUMERO. Lee su prompt y este contexto; verifica sus dependencias antes de editar».
