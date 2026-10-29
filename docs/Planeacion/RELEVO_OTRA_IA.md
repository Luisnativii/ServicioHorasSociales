# Continuidad con otra IA: OpenCode u otro agente

Fecha de corte: 29 de octubre de 2026, America/El_Salvador. Guía de relevo, issue #115.

Este documento permite continuar sin depender del historial del chat. Es una fotografía del avance y una cola propuesta; **GitHub y el código actualizado prevalecen para conocer el estado vivo**. No autoriza implementar veinte tareas de golpe ni integrar PR sin solicitud del propietario.

## Mensaje inicial para copiar a la otra IA

> Trabaja en este repositorio siguiendo AGENTS.md, docs/GITFLOW.md y docs/Planeacion/CONTEXTO_IA.md. Lee docs/Planeacion/RELEVO_OTRA_IA.md para conocer el avance, las decisiones y la cola. Primero verifica GitHub, ramas, dependencias y cambios locales; no supongas que la carpeta está actualizada. Continúa un solo issue por entrega, empezando por #26 si sigue abierto y sus dependencias están integradas. Usa la rama del issue desde develop, programa, ejecuta comprobaciones específicas dentro de Docker y prepara un PR hacia develop. Detente antes del merge y no crees tags manuales. No repitas toda la batería local sin motivo: GitHub CI ejecuta la completa. Conserva .env, .secrets, hs-dev_pgdata y documentos locales; no imprimas secretos ni uses git add indiscriminadamente. Si falta una decisión importante, prepara una propuesta concreta y pregunta solo lo que realmente falte; no vuelvas a preguntar lo ya aprobado. Comunica brevemente qué quedó funcional, cómo probarlo y qué sigue. La lista de veinte tareas es una secuencia, no autorización para saltar dependencias.

## Fuentes que debe leer y prioridad

1. `AGENTS.md` y [GitFlow](../GITFLOW.md): reglas de trabajo y entrega.
2. [Contexto para IA](CONTEXTO_IA.md), [decisiones vigentes](DECISIONES_VIGENTES.md) y el issue concreto con comentarios y dependencias nativas.
3. [Contrato de acceso aceptado](CONTRATO_ACCESO.md), [identidad Archivo](IDENTIDAD_VISUAL.md) y las reglas pertinentes en [alcance](ALCANCE_VIGENTE.md), [reglas](REGLAS_CONSOLIDADAS.md) y [trazabilidad](FUENTES_Y_TRAZABILIDAD.md).
4. [Modelo nuevo propuesto](MODELO_DATOS_PROPUESTO.md), únicamente los bloques pertinentes. **Solo M10 revisión 0.2 está aceptado** al corte de esta guía. Los demás bloques necesitan aprobación antes del SQL; una regla funcional confirmada no aprueba automáticamente todos sus campos/restricciones.
5. [Plan original](Plan_Programacion_IA_Horas_Sociales.md), como antecedente. No ejecutar sus etapas históricas por encima de acuerdos recientes.

Los encabezados «Instrucción vigente del contrato aceptado» de los issues de identidad sustituyen sus criterios históricos contradictorios: aunque un checklist antiguo mencione refresh, reset o revocación, esas funciones están excluidas. Si aparece una contradicción nueva, documentarla; no resolverla inventando alcance.

## Estado verificado al corte

Repositorio: <https://github.com/Luisnativii/ServicioHorasSociales>. La integración está en `develop`, commit `5627b76`, correspondiente al PR #114. Último tag obtenido: `V0.0.13`; consultar de nuevo antes de empezar. No había PR abiertos al preparar esta guía, antes de su propio PR.

| Disponible | Evidencia y límites |
| --- | --- |
| Base Go/React/PostgreSQL y Docker | Código en `apps/`, migraciones, salud HTTP, persistencia y portal mínimo. No crear otro proyecto dentro de la carpeta. |
| Flujo GitHub y pruebas aisladas | Issues, ramas y PR; `hs-test` separado de `hs-dev`; CI completo en GitHub; versionado por tags. |
| Contrato HTTP y roles de base | Errores/paginación comunes y cuentas separadas `hs_app`/`hs_migrator`. |
| Administradores | Migración y repositorio reales; herramienta Docker para crear primera cuenta y adicionales; hash Argon2id. No hay cuenta predeterminada. |
| JWT | Firma/claims/plazo, identidad por contexto y grupo privado protegido. No implementa login por sí solo. |
| Cookies/CSRF | `GET /api/v1/auth/csrf`, origen exacto, doble envío firmado ligado al JWT y helpers de emisión/borrado. Login/logout aún no tienen endpoints. |
| Identidad visual | Archivo aceptado en #109. Guía versionada; el prototipo mostrado en el chat no es el módulo real ni persiste estudiantes. |

PR #112 (#23), #113 (#24) y #114 (#25) ya están integrados. #22 y #27 se cerraron **como no planificados**, no como funcionalidades construidas. No recrearlos.

Pendientes: login/me (#26), logout (#28), gestión de cuentas, frontend autenticado y módulos de negocio. No existen todavía CRUD de estudiantes/proyectos/horas ni dashboards/Excel funcionales. No presentar la maqueta como producto terminado.

## Cómo obtener el código actualizado

La carpeta local puede tener cambios sin subir; GitHub tiene la integración compartida. Inspeccionar ambos antes de actualizar:

```bash
pwd
git status --short --branch
git remote -v
git fetch origin --prune --tags
git log -5 --oneline
gh pr list --repo Luisnativii/ServicioHorasSociales --state open
```

Si hay cambios locales, conservarlos y determinar su dueño antes de cambiar de rama. No ejecutar `reset --hard`, `clean -fd`, borrados de archivos ni stash automático como solución genérica. No inventar identidad Git.

Con el trabajo previo guardado y sin cambios incompatibles:

```bash
git switch develop
git pull --ff-only origin develop
git fetch origin --tags
bash scripts/version.sh
```

`main` contiene solo README y está reservada. No usarla como base de programación ni destino de PR. Si se recibe una carpeta sin `.git`, clonar el repositorio en otra ubicación y conciliar los archivos locales; no ejecutar `git init` para fabricar un historial nuevo ni copiar secretos al repositorio remoto.

## Ciclo de un issue y cómo subirlo

1. Leer el issue, su instrucción vigente y sus dependencias nativas; comprobar que los prerrequisitos estén aceptados e integrados. Un PR abierto no desbloquea dependientes. Revisar además aprobaciones del modelo: cerrar una fuente documental no aprueba todo el SQL.
2. Asignar responsable y agregar `status:ready`. La automatización crea/reutiliza `tipo/NUMERO-descripcion` y deja el nombre exacto en el issue. No adivinar ni crear una segunda rama si ya existe.
3. Actualizar referencias y usar esa rama. Si ya tiene avances, conservarlos.

```bash
gh issue view NUMERO --repo Luisnativii/ServicioHorasSociales
gh issue edit NUMERO --repo Luisnativii/ServicioHorasSociales --add-assignee @me --add-label status:ready
git fetch origin --prune
# Sustituir RAMA_EXACTA por la publicada por el workflow:
git switch --track origin/RAMA_EXACTA
```

Si la rama existe localmente, usar `git switch RAMA_EXACTA`. `gh` necesita la cuenta GitHub del operador con acceso; si falta autenticación, completar el acceso por sus medios normales. No pedir tokens por el chat ni copiar helpers temporales de otro agente.

4. Implementar exclusivamente el alcance aprobado; reutilizar helpers y componentes existentes. No agregar pantallas/rutas simuladas como si estuvieran conectadas. Documentar contratos cuando cambien.
5. Ejecutar comprobaciones pertinentes dentro de Docker. Cambios de documentación: enlaces/rutas/diff; backend: compilación y pruebas del paquete afectado; persistencia: PostgreSQL real en `hs-test`; frontend: TypeScript/build y comprobar el flujo modificado. No repetir toda CI local si ya pasó y nada lo justifica. **Nunca omitir o desactivar los checks de GitHub para ganar tiempo.**
6. Revisar diff y secretos; agregar solo archivos propios de la tarea. Confirmar que `.env`, `.secrets/` y documentos ajenos queden fuera.

```bash
git diff --check
git diff
git add RUTAS_EXPLICITAS
git diff --cached
git commit -m 'tipo(area): resultado concreto (#NUMERO)'
git push -u origin RAMA_EXACTA
gh pr create --repo Luisnativii/ServicioHorasSociales --base develop \
  --head RAMA_EXACTA --title 'tipo(area): resultado concreto (#NUMERO)' \
  --body-file /ruta/al/cuerpo-del-pr.md --label version:patch
```

El cuerpo incluye una línea `Closes #NUMERO`, resultado, pruebas realmente ejecutadas y límites. Elegir exactamente una etiqueta `version:patch|minor|major` según el cambio y GitFlow; el ejemplo no impone patch para todo. No afirmar pruebas inexistentes. Evitar secretos en archivos temporales o cuerpos de PR.

7. Entregar el enlace y detenerse antes de integrar. Solo hacer merge si el propietario lo solicita y pasan los checks requeridos, mediante squash hacia `develop`; sin bypass ni force push. Si GitHub pide actualizar la rama, integrar `origin/develop` en la rama de tarea, resolver conflictos conservando avances y subir normalmente. No reescribir historia publicada.
8. Después de integrar, volver a actualizar `develop` y Docker. La pipeline crea el tag cuando corresponde. No editar la versión npm para simular una release ni crear/mover tags manualmente.

## Operación local y pruebas sin dañar datos

Todo Go, Node, npm y PostgreSQL se ejecuta en Docker. Git, shell y CLI de GitHub pueden ejecutarse en el anfitrión. Usar los wrappers del repositorio, sin cambiar proyectos/volúmenes ni requerir runtimes instalados en el equipo.

Para una instalación local nueva, consultar primero README. `bash scripts/init-local.sh` prepara configuración ausente y conserva `.env` existente; no sustituye credenciales de una base que ya existe. Para una carpeta ya configurada:

```bash
bash scripts/migrate.sh dev
bash scripts/dc.sh up -d --build --wait --wait-timeout 180
```

Abrir <http://localhost:8080/>. Salud: `/api/health/live` y `/api/health/ready`. Estos endpoints siguen públicos. No hacer migraciones de bloques sin aprobación. Nunca borrar `hs-dev_pgdata`, usar `down -v`, cambiar una contraseña existente ni eliminar datos para resolver un error.

Identidad ya preparada:

```bash
# Solo si falta preparar la clave, con la imagen web de desarrollo construida:
bash scripts/init-jwt-key.sh
# Crear una cuenta solo cuando el operador lo solicite, sin contraseña en argumentos:
bash scripts/admin.sh create --email CORREO_REAL_DEL_OPERADOR
```

No crear cuentas de ejemplo automáticamente. Conservar `.secrets/hs-dev-jwt.key`: cambiarla invalida copias previas del JWT. `APP_ORIGIN` local es `http://localhost:8080`; HTTP solo es válido para localhost/127.0.0.1. No desactivar Secure globalmente para resolver desarrollo; HTTPS usa cookies Secure. Ver detalle en README.

`bash scripts/init-test.sh` prepara `hs-test`; `scripts/dc-test.sh` lo aísla de desarrollo. `bash scripts/ci.sh` ejecuta la batería completa cuando sea necesaria; GitHub ya la ejecuta al abrir/actualizar PR. No correr dos secuencias a la vez sobre `hs-test`. Las pruebas que mutan datos usan exclusivamente fixtures de prueba, no cuentas o estudiantes reales.

## Decisiones que no debe volver a inventar

- Un rol ADMIN, varias cuentas, acceso completo; correo y contraseña. Alta solo mediante herramienta Docker, sin registro público, reset ni cambio de contraseña web.
- Contraseña: mínimo seis caracteres con un dígito ASCII, máximo 128 bytes, sin exigir mayúsculas/símbolos ni recortar espacios. Argon2id conforme al contrato; no guardar texto claro.
- JWT de dos horas fijas, sin refresh, recuperación ni revocación. Desactivar impide nuevos logins; JWT previos vencen por tiempo. Logout borra cookies y no invalida copias en servidor. Se permite cero cuentas activas.
- `hs_access` HttpOnly; JWT nunca en JSON/localStorage/URL. Mutaciones con Origin y CSRF. Reutilizar `auth.Tokens`, `auth.Browser`, `RequireAdmin`, `RequireCSRF` y `Options.RegisterPrivate`.
- Archivo es la identidad visual elegida: paleta, tipografía, navegación y estados están en `IDENTIDAD_VISUAL.md`. «Nuevo estudiante» abre página propia, no formulario debajo del listado ni modal. No rediseñar la marca sin petición.
- Cada proyecto tendrá **un tipo y un área**, con catálogos de N opciones. No interpretar «un tipo o un área» como la decisión vigente.
- El propietario aclaró que existe **un solo concepto llamado modalidad**: se puede elegir una existente o crear una. «Habilidades blandas» es un ejemplo de modalidad, no un campo separado. Módulo/cardinalidad/obligatoriedad se concretan en #13 antes del SQL.
- El propietario quiere coordinador/contacto escrito como dato simple; no tabla/catálogo de coordinadores ni enum por deducción. #11 debe precisar la traducción del acuerdo y revisar su efecto en #44/#45, cuyos títulos históricos contemplaban otra estructura.
- Datos académicos pertenecen al estudiante; #12 concreta historia/reportes sin exigir volver a explicar esa respuesta. El propietario hará carga inicial/saldos por SQL; no construir importador web por iniciativa propia. Eso no decide por sí solo correcciones, corte histórico o idempotencia.
- Se indicó que una participación abierta puede terminar finalizada o abandonada. #8 concreta los demás efectos de cierre sin inventar reapertura ni estados adicionales.

Estas aclaraciones proceden del chat del propietario y deben incorporarse a sus contratos/issues al ejecutarlos. No equivalen a aprobación de todos los bloques físicos propuestos. Preguntar solo por detalles materiales aún sin resolver.

## Cola de los siguientes 20 issues existentes

No crear duplicados de estos issues. El orden favorece terminar acceso real y después concretar decisiones/catálogos. Cada fila es una entrega con PR propio; pasar a la siguiente requiere cumplir su condición, no solo haber abierto un PR. Estado y dependencias son del corte: consultar nuevamente GitHub.

| Orden | Issue y resultado | Condición de inicio / límite principal |
| --- | --- | --- |
| 1 | [#26 · Login y consulta de identidad](https://github.com/Luisnativii/ServicioHorasSociales/issues/26) | #23/#24/#25 integrados: listo al corte. Argon2id, me, errores genéricos, límite de intentos, cookies reales; sin refresh. |
| 2 | [#28 · Cerrar sesión](https://github.com/Luisnativii/ServicioHorasSociales/issues/28) | Integrar #26; #25 ya integrado. Borrar cookies incluso con JWT vencido, sin revocación. Ignorar criterios antiguos de revocar/reiniciar revocación. |
| 3 | [#29 · Administrar cuentas](https://github.com/Luisnativii/ServicioHorasSociales/issues/29) | Integrar #28; #20 aceptado. Listar/activar/desactivar; sin alta/reset/cambio de contraseña web ni mínimo de activos. |
| 4 | [#30 · Cliente de sesión, CSRF y vencimiento](https://github.com/Luisnativii/ServicioHorasSociales/issues/30) | Integrar #26/#28; #25 integrado. Manejar 401/vencimiento coherentemente; no renovación ni reenvío ciego de escrituras. |
| 5 | [#31 · Pantalla real de acceso](https://github.com/Luisnativii/ServicioHorasSociales/issues/31) | Integrar #30; #26/#25/#109 integrados. Diseño Archivo, acceso real y teclado; sin recuperación ficticia. |
| 6 | [#32 · Navegación y layout](https://github.com/Luisnativii/ServicioHorasSociales/issues/32) | Integrar #31; #28/#109 integrados. Recarga con sesión, salida, estados vacíos reales y móvil; no inventar módulos/datos. |
| 7 | [#33 · Gestión web de cuentas](https://github.com/Luisnativii/ServicioHorasSociales/issues/33) | Integrar #32 y #29. Listado/estado usando API, sin formulario de alta ni nuevos roles. |
| 8 | [#19 · Guardia de dependencias en CI](https://github.com/Luisnativii/ServicioHorasSociales/issues/19) | #3 integrado; puede adelantarse si conviene. Revisar dependencias y evidencia real, fallar de forma segura ante errores de API; no bloquear documentación por supuestos inventados. |
| 9 | [#7 · Tipos y áreas](https://github.com/Luisnativii/ServicioHorasSociales/issues/7) | #5/#6 cerrados. Registrar un tipo y un área por proyecto; precisar historia de reclasificación y fuente del catálogo. No implementar #40 todavía. |
| 10 | [#11 · Contactos y coordinadores](https://github.com/Luisnativii/ServicioHorasSociales/issues/11) | #5/#6 cerrados. Aplicar dato escrito simple, concretar campos/relaciones y ajustar dependientes #44/#45 si corresponden. Sin tabla de coordinadores por el plan antiguo. |
| 11 | [#12 · Historia académica/reportes](https://github.com/Luisnativii/ServicioHorasSociales/issues/12) | Integrar #7. Concretar datos actuales/históricos y clasificación temporal a partir de lo ya respondido; aprobar solo el bloque pertinente. |
| 12 | [#13 · Modalidad](https://github.com/Luisnativii/ServicioHorasSociales/issues/13) | #5/#6 cerrados. Un concepto seleccionable/creable; confirmar ubicación y cardinalidad pendientes, sin campo adicional de habilidades blandas. |
| 13 | [#8 · Estados y cierres](https://github.com/Luisnativii/ServicioHorasSociales/issues/8) | #5/#6 cerrados. Documentar abierta/finalizada/abandonada y efectos pendientes sobre participantes/horas. Resolver casos concretos antes de código transaccional. |
| 14 | [#9 · Correcciones de horas](https://github.com/Luisnativii/ServicioHorasSociales/issues/9) | #5/#6 cerrados. Decisión o exclusión expresa; la carga manual del propietario no autoriza editar/borrar historia ni nuevas horas tras cierre. |
| 15 | [#10 · Saldos anteriores](https://github.com/Luisnativii/ServicioHorasSociales/issues/10) | #5/#6 cerrados. Carga por SQL, procedencia/corte/deduplicación; sin inventar fechas ni importador web. |
| 16 | [#34 · Estructura académica](https://github.com/Luisnativii/ServicioHorasSociales/issues/34) | #5/#18/#21 cerrados **y bloque académico revisado/aceptado antes del SQL**. Migración nueva, facultad/carrera, integridad y consultas sqlc. No todo el modelo está aprobado por cerrar #5. |
| 17 | [#35 · API académica protegida](https://github.com/Luisnativii/ServicioHorasSociales/issues/35) | Integrar #34; #28/#17 integrados. Endpoints reales con JWT/CSRF, paginación, validación e inactivos históricos. |
| 18 | [#36 · Web de facultades/carreras](https://github.com/Luisnativii/ServicioHorasSociales/issues/36) | Integrar #35; #32 integrado. Archivo, persistencia real, errores que conservan formulario, sin paginación ficticia. |
| 19 | [#37 · Catálogo territorial](https://github.com/Luisnativii/ServicioHorasSociales/issues/37) | #5/#18 cerrados y estructura territorial aceptada. Fuente oficial verificada, jerarquía y carga repetible; sin inventar etiquetas/geografía. |
| 20 | [#38 · API territorial](https://github.com/Luisnativii/ServicioHorasSociales/issues/38) | Integrar #37; #28/#17 integrados. Jerarquía/filtros reales, SQL parametrizado y acceso protegido. |

No es un orden cronológico rígido entre decisiones independientes: se puede adelantar una fila sin bloqueos si el propietario lo pide, conservando una tarea por entrega. Las filas de aclaración pueden terminar en exclusión o ajuste de alcance documentado, nunca en una implementación inventada.

Después siguen #39–#43, instituciones, estudiantes, proyectos, participaciones, horas y reportes según las dependencias vigentes. #44/#45 requieren conciliar el acuerdo de coordinador simple antes de seguir. También permanecen abiertos #4 (seguimiento) y #14 (equipo/metas); esta lista no afirma que el resto haya desaparecido ni autoriza despliegue/producción.

## Pasar la carpeta y coordinar agentes

- En el mismo equipo, abrir la carpeta raíz existente y conservar `.git` para ver historial/ramas. El documento local `docs/Planeacion/Logica_Negocio.md` está sin versionar y pertenece al propietario: no agregarlo ni eliminarlo automáticamente.
- Los volúmenes Docker **no viajan dentro de la carpeta**. En otro equipo, clonar trae código, no la base real. Para datos existentes hace falta respaldo/restauración acordados; no afirmar que copiar `.env` importa la base.
- No pegar `.env`, claves, hashes, cookies o tokens en prompts ni subir una carpeta comprimida con secretos a una IA remota. Un agente local puede usar la configuración para ejecutar el proyecto sin mostrarla en la conversación. Los secretos permanecen fuera de Git e imágenes.
- No ejecutar dos agentes escribiendo en la misma rama/carpeta, ni dos pruebas sobre el mismo `hs-test`. Terminar/guardar una entrega y hacer relevo explícito. No mezclar una tarea de otro agente en un commit propio.
- Cada entrega informa: issue/PR, comportamiento funcional, comprobaciones realmente realizadas, cómo probarlo, límites y siguiente dependencia. Si el propietario pide avanzar, trabajar: no limitarse a decir que se puede hacer.
