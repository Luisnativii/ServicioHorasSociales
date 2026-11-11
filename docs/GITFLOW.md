# GitFlow del Portal de Horas Sociales

## Ramas y versión inicial

- **main**: rama reservada con un único `README.md`. No recibe código ni PR de funcionalidades por ahora.
- **develop**: rama predeterminada y de integración. Contiene el proyecto directamente en la raíz: `apps/`, `api/`, `deploy/`, `scripts/`, `docs/`, `tests/` y `.github/`.
- **tipo/NUMERO-descripcion**: una rama por issue, creada desde `develop` y eliminada por GitHub después de integrar el PR.
- **V0.0.1**: tag anotado de la base inicial. Sustituye la numeración provisional `0.1.0-dev` usada durante la preparación local. No representa un despliegue en producción.

`develop` es predeterminada para que GitHub utilice las plantillas y para que `Closes #NUMERO` cierre el issue al integrar el PR. Estas funciones dependen de la rama predeterminada: [plantillas](https://docs.github.com/en/communities/using-templates-to-encourage-useful-issues-and-pull-requests/about-issue-and-pull-request-templates) y [vinculación de issues](https://docs.github.com/en/issues/tracking-your-work-with-issues/using-issues/linking-a-pull-request-to-an-issue).

> **GitHub Actions desactivado.** Desde noviembre de 2026 GitHub rechaza iniciar los jobs porque la cuenta está bloqueada por facturación. Se retiraron los workflows para que cada PR no quede marcado por un fallo ajeno al código. Las pruebas siguen disponibles mediante Docker local; las ramas, estados y etiquetas se gestionan manualmente hasta que el propietario decida reactivar CI.

La publicación inicial del proyecto en `develop` es la única excepción al requisito de PR: establece la base sobre la que podrá funcionar CI. Los siguientes cambios pasan por issue, rama y PR.

## 1. Crear un issue

Abre [Issues → New issue](https://github.com/Luisnativii/ServicioHorasSociales/issues/new/choose) y elige Funcionalidad, Error, Tarea técnica o Documentación.

Cada issue debe tener objetivo, alcance, criterios de aceptación comprobables y referencias/dependencias. No incluyas contraseñas ni información personal. Primero se define y aprueba el trabajo; crear un issue no autoriza implementar un modelo de negocio todavía no aprobado.

Las plantillas asignan un tipo y `status:backlog`. Podemos crear y concretar juntos los issues a medida que decidas qué trabajar. Para pedírmelo, puedes decir: «Crea un issue para [objetivo] con sus criterios de aceptación».

| Etiqueta del issue | Rama | Prefijo del título del PR |
| --- | --- | --- |
| `type:feature` | `feature/12-descripcion` | `feat` |
| `type:fix` | `fix/12-descripcion` | `fix` |
| `type:chore` | `chore/12-descripcion` | `chore` |
| `type:docs` | `docs/12-descripcion` | `docs` |
| `type:refactor` | `refactor/12-descripcion` | `refactor` |
| `type:test` | `test/12-descripcion` | `test` |
| `type:ci` | `ci/12-descripcion` | `ci` |

Se exige exactamente un tipo. El número es el número real del issue; `12` es solo un ejemplo.

## 2. Iniciar el trabajo y generar la rama

Cuando el issue esté definido:

1. Comprueba manualmente sus dependencias nativas. Cada dependencia debe estar integrada, cerrada como completada, con `status:done` y sin una decisión pendiente aplicable.
2. Asigna responsable y cambia `status:backlog` o `status:ready` a `status:in-progress`.
3. Actualiza `develop` sin sobrescribir cambios locales.
4. Crea o reutiliza `tipo/NUMERO-descripcion` desde el commit actual de `develop`.

```bash
git fetch origin --prune --tags
git switch develop
git pull --ff-only origin develop
git switch -c feature/12-definir-modelo-aprobado
```

Si la rama ya existe en el remoto, usa `git switch --track origin/RAMA`; si existe localmente, usa `git switch RAMA`. Nunca crees una segunda rama para el mismo issue ni uses `status:ready` esperando una automatización.

Si una dependencia terminada conserva `needs:decision` porque contiene otros bloques abiertos, la aceptación humana del bloque concreto debe registrarse en el issue dependiente con `<!-- hs:dependency-gate:accepted #NUMERO -->` y una explicación. Esta marca solo resuelve esa decisión: no convierte una tarea abierta, no planificada o sin entrega en una dependencia válida.

Estados manuales: `status:backlog` → `status:ready` opcional → `status:in-progress` → issue cerrado / `status:done`. Al reabrirlo vuelve al backlog.

## 3. Desarrollar y abrir el PR

Trabaja en la rama del issue, ejecuta las comprobaciones relevantes dentro de Docker y crea commits con propósito claro, por ejemplo:

```text
feat(api): agregar validación aprobada (#12)
fix(web): corregir estado de conexión (#18)
docs: documentar operación local (#20)
```

El PR debe apuntar a **develop** y tener:

- Rama que coincida con el número y tipo del issue.
- Título como `feat(api): describir el cambio (#12)`.
- Una línea exacta `Closes #12` en su descripción.
- Explicación del resultado y evidencia de comprobación.
- **Una sola etiqueta** `version:patch`, `version:minor` o `version:major`.

No hay checks automáticos mientras Actions esté desactivado. Antes de abrir o actualizar el PR ejecuta las comprobaciones pertinentes dentro de Docker y registra comandos y resultados reales. Para cambios amplios o antes de integrar, usa `bash scripts/ci.sh`; para documentación basta como mínimo revisar enlaces afectados y ejecutar `git diff --check`.

Para actualizar el PR con nuevos commits, vuelve a subir la misma rama. Si `develop` avanzó y GitHub pide actualizarla, integra ese avance antes de repetir las comprobaciones.

La rama debe conservar el requisito de PR y conversaciones resueltas, pero no puede exigir los checks `GitFlow` o `Docker` mientras sus workflows no existan. El propietario debe retirar esos nombres de las reglas de protección para evitar un bloqueo permanente; esto no autoriza pushes directos ni merges sin evidencia.

## 4. Integrar y aumentar la versión

Se usa **Squash and merge**: el título y cuerpo del PR forman el commit que entra a `develop`. GitHub cierra el issue por `Closes #NUMERO` y elimina la rama remota. Después del merge, el responsable aplica `status:done` manualmente si GitHub no lo conserva.

El incremento que correspondería se conserva como etiqueta del PR para facilitar una reactivación futura:

| Etiqueta elegida en el PR | Desde V0.0.1 | Uso |
| --- | --- | --- |
| `version:patch` | `V0.0.2` | Corrección, documentación, mantenimiento o cambio pequeño |
| `version:minor` | `V0.1.0` | Capacidad nueva que merece incrementar MINOR |
| `version:major` | `V1.0.0` | Cambio incompatible |

No hay incremento en cada commit, al abrir el issue ni al abrir el PR. Un título con `!` exige `version:major`. Mientras Actions permanezca desactivado, integrar un PR no crea un tag. No se deben fabricar o mover tags manualmente para aparentar que una validación automática ocurrió.

Los tags ya publicados siguen siendo la fuente histórica de verdad y no se reescriben. La política de reanudación del versionado se definirá al recuperar CI o en el issue de releases; no se publican imágenes ni se despliegan servicios por este cambio.

`package.json` conserva la versión de arranque del paquete privado, `0.0.1`: no es la fuente del número de release. `scripts/version.sh` obtiene la versión del tag más cercano del historial; `scripts/dc.sh` la pasa a Compose, al frontend y a las etiquetas de las imágenes. La variable `APP_VERSION` permite una construcción local explícita. Si se descarga un ZIP sin Git, el respaldo es `0.0.1`.

Después de una integración:

```bash
git switch develop
git pull --ff-only origin develop
git fetch origin --tags
bash scripts/version.sh
bash scripts/dc.sh up -d --build --wait --wait-timeout 180
```

Las imágenes locales serán `hs-api:VERSION-dev`, `hs-web:VERSION-dev` y `hs-migrator:VERSION-dev`. La pantalla mostrará `VERSION-dev`. Mantén el checkout local en `develop` o en una rama de tarea mientras uses Docker; `main` contiene únicamente el README y no dispone de las fuentes que montan los contenedores.

## Comprobaciones locales

- **Política**: `scripts/check-gitflow.sh` conserva pruebas de las reglas de rama/issue/PR y SemVer, pero no consulta ni modifica GitHub.
- **Dependencias**: se revisan manualmente antes de crear la rama y antes de integrar. Un cierre sin entrega aceptada o con decisión pendiente no desbloquea implementación.
- **Docker**: `scripts/ci.sh` usa exclusivamente `hs-test`, con credenciales, redes y volumen propios; construye imágenes y comprueba aislamiento, Go/TypeScript/Vite, migraciones, recuperación, persistencia, runtime sin root y vulnerabilidades conocidas.
- **Versionado**: suspendido mientras Actions esté desactivado. Las etiquetas del PR documentan la intención y los tags publicados no se mueven.
- **Issues**: ramas, asignación y estados se gestionan manualmente.

```bash
# Pruebas de las reglas de GitFlow, dentro de Node en Docker.
bash scripts/check-gitflow.sh
# Secuencia completa en hs-test, independiente del portal local.
bash scripts/ci.sh
# Comprobaciones habituales con hs-test ya preparado e iniciado.
bash scripts/check.sh
```

`scripts/ci.sh` funciona localmente: detiene únicamente los contenedores de pruebas al terminar y conserva su volumen. No usa `.env` ni modifica o detiene `hs-dev`. Preparación y limpieza en [README](../README.md#pruebas-aisladas-de-desarrollo).

Antes de promover código a `main` o publicar producción, definiremos una etapa de release independiente y ajustaremos su protección; por ahora `main` queda reservada.
