# GitFlow del Portal de Horas Sociales

## Ramas y versión inicial

- **main**: rama reservada con un único `README.md`. No recibe código ni PR de funcionalidades por ahora.
- **develop**: rama predeterminada y de integración. Contiene el proyecto directamente en la raíz: `apps/`, `api/`, `deploy/`, `scripts/`, `docs/`, `tests/` y `.github/`.
- **tipo/NUMERO-descripcion**: una rama por issue, creada desde `develop` y eliminada por GitHub después de integrar el PR.
- **V0.0.1**: tag anotado de la base inicial. Sustituye la numeración provisional `0.1.0-dev` usada durante la preparación local. No representa un despliegue en producción.

`develop` es predeterminada para que GitHub utilice las plantillas y workflows de issues, y para que `Closes #NUMERO` cierre el issue al integrar el PR. Estas funciones dependen de la rama predeterminada: [plantillas](https://docs.github.com/en/communities/using-templates-to-encourage-useful-issues-and-pull-requests/about-issue-and-pull-request-templates) y [vinculación de issues](https://docs.github.com/en/issues/tracking-your-work-with-issues/using-issues/linking-a-pull-request-to-an-issue).

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

Cuando el issue esté definido, asigna responsable y agrega **`status:ready`**. El workflow **Issues**:

1. Verifica que quien inicia la tarea tenga permiso de escritura.
2. Obtiene el número y título real del issue.
3. Elimina acentos y símbolos del título y genera un nombre corto en minúsculas.
4. Crea `tipo/NUMERO-descripcion` desde el commit actual de `develop`.
5. Cambia el estado a `status:in-progress` y deja en el issue un enlace y los comandos de Git.

No crea ramas automáticamente por el simple hecho de proponer una tarea. Al repetir `status:ready` para el mismo tipo/issue se reutiliza la rama existente, sin sobrescribir su trabajo. Evita cambiar el tipo de un issue después de empezar a trabajar.

Ejemplo, cuando la automatización haya creado `feature/12-definir-modelo-aprobado`:

```bash
git fetch origin --prune --tags
git switch --track origin/feature/12-definir-modelo-aprobado
```

Si ya existe localmente, usa `git switch feature/12-definir-modelo-aprobado`. También puedes pedirme: «Trabaja el issue #12 en su rama y prepara el PR hacia develop».

Estados: `status:backlog` → `status:ready` → `status:in-progress` → issue cerrado / `status:done`. Al reabrirlo vuelve al backlog.

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

Los checks **GitFlow** y **Docker** son obligatorios. Para actualizar el PR con nuevos commits, vuelve a subir la misma rama. Si `develop` avanzó y GitHub pide actualizarla, integra ese avance antes de repetir las comprobaciones.

La rama está protegida con PR y checks, sin exigir una aprobación ajena que bloquearía a un único mantenedor. Las conversaciones deben estar resueltas. La excepción administrativa de GitHub queda disponible para recuperar el repo, pero no forma parte del trabajo normal.

## 4. Integrar y aumentar la versión

Se usa **Squash and merge**: el título y cuerpo del PR forman el commit que entra a `develop`. GitHub cierra el issue por `Closes #NUMERO` y elimina la rama remota. El workflow Issues marca el issue como terminado.

La pipeline CI valida la integración de `develop`. Solo después de pasar todas sus comprobaciones crea un tag anotado sobre ese commit:

| Etiqueta elegida en el PR | Desde V0.0.1 | Uso |
| --- | --- | --- |
| `version:patch` | `V0.0.2` | Corrección, documentación, mantenimiento o cambio pequeño |
| `version:minor` | `V0.1.0` | Capacidad nueva que merece incrementar MINOR |
| `version:major` | `V1.0.0` | Cambio incompatible |

No hay incremento en cada commit, al abrir el issue ni al abrir el PR. Un título con `!` exige `version:major`. La pipeline rechaza pushes directos sin un PR integrado, salvo la base inicial ya etiquetada.

Si CI falla, no se crea un tag. Una ejecución cuyo commit ya fue superado por otra integración no publica una versión atrasada: la siguiente integración vigente se valida y etiqueta. GitHub serializa CI de `develop` y puede agrupar ejecuciones pendientes cuando entran muchas integraciones seguidas; el tag representa la integración vigente validada, no un contador de todos los PR históricos.

Los tags son la fuente de verdad y no se reescriben. La pipeline no hace commits de versión, no sube imágenes a un registro y no despliega servicios. Se mantienen permisos de solo lectura durante las pruebas; solo el job de etiquetado puede escribir contenido, y nunca se ejecuta para un PR.

`package.json` conserva la versión de arranque del paquete privado, `0.0.1`: no es la fuente del número de release. `scripts/version.sh` obtiene la versión del tag más cercano del historial; `scripts/dc.sh` la pasa a Compose, al frontend y a las etiquetas de las imágenes. La variable `APP_VERSION` permite que CI construya exactamente la versión propuesta. Si se descarga un ZIP sin Git, el respaldo es `0.0.1`.

Después de una integración:

```bash
git switch develop
git pull --ff-only origin develop
git fetch origin --tags
bash scripts/version.sh
bash scripts/dc.sh up -d --build --wait --wait-timeout 180
```

Las imágenes locales serán `hs-api:VERSION-dev`, `hs-web:VERSION-dev` y `hs-migrator:VERSION-dev`. La pantalla mostrará `VERSION-dev`. Mantén el checkout local en `develop` o en una rama de tarea mientras uses Docker; `main` contiene únicamente el README y no dispone de las fuentes que montan los contenedores.

## Pipeline y comprobaciones

- **GitFlow**: valida issue, rama, título, destino, etiqueta de versión y procedencia de las integraciones; propone la versión.
- **Docker**: ejecuta `scripts/ci.sh`, que inicializa credenciales temporales, construye imágenes, comprueba Go/TypeScript/Vite, migraciones, recuperación, persistencia, runtime sin root y vulnerabilidades conocidas.
- **Versionado**: crea el tag si los checks pasaron, hay un PR integrado y el commit sigue siendo el vigente de `develop`.
- **Issues**: crea ramas y mantiene los estados al iniciar, cerrar o reabrir issues.

Las acciones oficiales están fijadas por SHA. No hay token personal ni secreto guardado en los workflows: usan el `GITHUB_TOKEN` de cada ejecución con permisos limitados.

```bash
# Pruebas de las reglas de GitFlow, dentro de Node en Docker.
bash scripts/check-gitflow.sh
# Comprobaciones habituales del proyecto ya iniciado.
bash scripts/check.sh
```

`scripts/ci.sh` está pensado para el runner temporal de GitHub Actions: detiene sus contenedores al terminar, conservando el volumen. No hace falta ejecutarlo en tu sesión de trabajo local.

Antes de promover código a `main` o publicar producción, definiremos una etapa de release independiente y ajustaremos su protección; por ahora `main` queda reservada.
