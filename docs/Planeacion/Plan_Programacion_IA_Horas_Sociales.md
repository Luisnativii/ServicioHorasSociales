# Plan de programación del portal de horas sociales con inteligencia artificial

Guía por etapas para construir validar y entregar el sistema

Versión 1.0 del plan  |  23 de octubre de 2026

Este plan organiza el desarrollo en 20 etapas pequeñas. Cada etapa explica su propósito, entrega un prompt para la IA y define cinco comprobaciones manuales. La propuesta utiliza Go para el backend, React con Tailwind CSS para el frontend y PostgreSQL, con desarrollo y ejecución dentro de Docker.

### Go es una buena opción para este proyecto

Go permite construir una API HTTP y trabajar con PostgreSQL mediante transacciones y un pool de conexiones. Se propone una sola aplicación backend organizada por módulos: facilita entender, probar y desplegar el portal en una laptop. El rendimiento se comprobará con datos de prueba; usar Go por sí solo no garantiza estabilidad ni evita errores en los registros de horas.

La principal inversión adicional será aprender Go idiomático, SQL y manejo de errores. La IA deberá explicar lo que implemente y entregar cambios que puedas revisar. No se incorporan microservicios, Kubernetes, Redis ni colas mientras no exista una necesidad medida.

### Cómo utilizar esta guía

- Revisar primero arquitectura, reglas de negocio y política de versiones.

- Pegar el prompt maestro al abrir una sesión con la IA; después pegar solo el prompt de la etapa actual.

- Pedir la implementación y comprobar sus cinco criterios manuales con datos de prueba.

- Registrar resultado, commit e imágenes. Avanzar únicamente cuando la etapa esté aceptada.

Presupuesto orientativo: 300 horas reloj para el proyecto completo, incluyendo trabajo previo y validación. La meta de cada estudiante es otra: 600 horas sociales, divididas en 300 internas y 300 externas.

## Tecnologías y responsabilidades

| Componente | Elección propuesta | Aplicación en el portal |
|---|---|---|
| Frontend | React + TypeScript + Vite + Tailwind CSS 4 | Tailwind obligatorio, compilado localmente en Docker. No CDN. |
| Interfaz y datos | React Router + TanStack Query | Navegación y solicitudes centralizadas. Formularios reutilizables. |
| Backend | Go + net/http + chi | Una API REST por módulos. Despliegue como binario en contenedor. |
| Persistencia | PostgreSQL 18 + pgx/v5 + pgxpool | SQL parametrizado, conexiones acotadas y transacciones explícitas. |
| SQL y esquema | sqlc + Goose + OpenAPI | Consultas tipadas, migraciones numeradas y contrato de integración. |
| Identidad | Argon2id + JWT corto + sesión revocable | Cookies protegidas y renovación controlada. Sin registro público. |
| Exportación y pruebas | Excelize + Go testing + Playwright | XLSX en backend; pruebas críticas con PostgreSQL real. |
| Entrega y operación | Docker Compose + GHCR + GitHub Actions | Tags por versión, digests y despliegue sin compilar en producción. |

### Versiones y dependencias

En la inicialización se verificarán versiones estables compatibles y se fijarán versiones exactas en go.mod, go.sum, package-lock.json y Dockerfiles. Para Go se elegirá una versión soportada compatible con pgx y Excelize; la documentación consultada de Excelize requiere al menos Go 1.25. No usar tags latest ni instalaciones sin versión para herramientas de compilación.

### Acceso y sesiones

Se mantiene una sola cuenta ADMIN. JWT se propone con expiración corta y sesión revocable en PostgreSQL, cookies HttpOnly y protección CSRF. sesion_admin es una tabla técnica adicional a las 19 tablas del modelo de negocio; debe documentarse mediante migración. Gmail y el acceso con Google quedan fuera de esta entrega inicial.

### Regla de tamaño del proyecto

Los listados serán paginados; dashboard y exportaciones consultarán al backend. No se cargarán todos los estudiantes o movimientos en memoria del navegador. El pool, los timeouts y los índices se ajustarán con mediciones del equipo real.

## Arquitectura y estructura del repositorio

El navegador accede a un solo dominio. web sirve la interfaz y dirige /api hacia Go; solo Go accede a PostgreSQL. En producción, cloudflared conecta el dominio con web dentro de Docker. PostgreSQL no tiene exposición pública.

| Carpeta o archivo | Responsabilidad |
|---|---|
| apps/api/cmd/api | Inicio del servidor y carga de configuración. |
| apps/api/cmd/admin | Alta inicial y recuperación de la cuenta administradora. |
| apps/api/internal/platform | HTTP, base de datos, configuración, logs y seguridad comunes. |
| apps/api/internal/modules/<modulo> | handler.go recibe HTTP; service.go aplica reglas; repository.go consulta datos. |
| apps/api/internal/db | Código generado por sqlc. No editar manualmente. |
| apps/api/db/queries y apps/api/db/migrations | Consultas SQL y cambios numerados del esquema. |
| apps/web/src/features/<modulo> | Pantallas, hooks, tipos y formularios de cada funcionalidad. |
| apps/web/src/components y lib | Componentes compartidos, cliente HTTP y utilidades. |
| api/openapi.yaml y docs | Contrato, reglas, decisiones, validaciones y manuales. |
| deploy y scripts | Dockerfiles, proxy, operaciones y manifiestos de releases. |
| .github/workflows y tests | Integración continua, pruebas y datos sintéticos. |

### Criterio para organizar el código

Usar módulos como auth, catalogos, instituciones, proyectos, estudiantes, participaciones, horas y reportes. El handler traduce HTTP; el service aplica las reglas; el repository ejecuta SQL. Crear interfaces donde una frontera o prueba las necesite, evitando capas vacías. La estructura crece con cada etapa; no hace falta llenar carpetas sin implementación.

## Docker en desarrollo y producción

| Servicio | Función | Red |
|---|---|---|
| web | React compilado servido por Nginx; /api se redirige a Go. | edge + app |
| api | Reglas, autenticación, consultas y exportaciones. | app + data |
| db | PostgreSQL con volumen persistente y sin puerto público. | data |
| migrator | Proceso temporal que aplica migraciones antes de habilitar una versión. | data |
| cloudflared | Conecta el dominio con http://web:8080 mediante salida hacia Cloudflare. | edge |
| tools / backup | Contenedores temporales para pruebas, generación y respaldo. | Solo las redes necesarias |
| portainer | Administración técnica de Docker con acceso restringido. | Administración separada |

### Ambientes separados

compose.yaml es la base; compose.dev.yaml habilita recarga y puertos locales; compose.prod.yaml usa imágenes probadas sin montar código. Usar proyectos hs-dev, hs-test, hs-staging y hs-prod, con credenciales y volúmenes distintos. Nunca conectar pruebas a producción.

Go, Node, sqlc, migraciones, pruebas y backups se ejecutan dentro de contenedores. La laptop necesita Docker, Git y editor. Cockpit se instala en Ubuntu porque administra el sistema operativo; Docker Engine y las tareas del host tampoco se ejecutan como servicios de la aplicación. Portainer y cloudflared sí estarán dockerizados.

### Persistencia y administración

PostgreSQL 18 utiliza el volumen en /var/lib/postgresql. No incluir datos dentro de la imagen ni usar down -v en el flujo normal. Los respaldos deben salir de la laptop y probarse mediante restauración. Las credenciales se montan desde archivos protegidos; .env.example solo contiene nombres y ejemplos ficticios.

Compose y Git serán la fuente de configuración. Portainer ayudará a consultar estado y logs; cualquier cambio persistente se reflejará en los archivos versionados. Su acceso al socket Docker equivale a privilegios elevados y debe reservarse al acceso técnico protegido.

## Versiones de código imágenes y base de datos

| Elemento | Ejemplo | Qué identifica |
|---|---|---|
| Git | v0.12.0 y commit 8c4d2ab | Código aceptado y documentación de la etapa. |
| Imagen Docker | ghcr.io/<propietario>/hs-api:0.12.0 | Artefacto ejecutable. API web y migrator comparten versión de release. |
| Digest | ghcr.io/<propietario>/hs-api@sha256:<digest> | Contenido exacto de una imagen; es la referencia del despliegue probado. |
| Migración | 00012_agregar_idempotencia.sql | Cambio del esquema, independiente del tag Docker. |

Los tags pueden cambiar técnicamente. La política del proyecto será no sobrescribir tags de release y registrar siempre sus digests. Una corrección produce 0.12.1; una nueva etapa compatible produce la siguiente versión menor. Después de 1.0.0, usar parches para correcciones y revisar compatibilidad antes de incrementar versiones mayores.

### Ejemplo de cierre de una etapa

```sh
git tag -a v0.12.0 -m "Movimientos de horas aceptados"
docker build -f deploy/docker/api.Dockerfile -t hs-api:0.12.0 .
docker build -f deploy/docker/web.Dockerfile -t hs-web:0.12.0 .
docker build -f deploy/docker/migrator.Dockerfile -t hs-migrator:0.12.0 .
```

Los nombres de Dockerfile son el contrato propuesto. La etapa 01 creará los comandos; migrator se incorporará al implementar migraciones. La IA prepara commits y tags locales; tú autorizas la publicación después de validar. Desde la etapa 16, GitHub Actions publica las imágenes aprobadas en GHCR.

### Una versión debe poder explicarse

El manifiesto de release registra commit, imágenes, digests, arquitectura, migración máxima aplicada, compatibilidad con la versión anterior y pruebas realizadas. Los archivos de release no contienen secretos. Volver a una imagen anterior no revierte los datos ni el esquema; solo se hará si conserva compatibilidad con la base actual.

## Reglas funcionales y decisiones de diseño

### Reglas confirmadas

- Una cuenta administradora; no hay acceso de estudiantes ni catálogo público. Coordinadores y contactos no son usuarios.

- Meta estudiantil: 300 internas y 300 externas, independientes. Horas enteras agregadas por movimientos; saldos iniciales posibles.

- El administrador informa fecha_horas; registrado_en conserva el instante de captura. No admitir horas después del cierre.

- Una institución por proyecto y varios proyectos por institución; una sola ubicación física de ejecución.

- Proyecto EXTERNA por defecto. Cambiarlo a INTERNA solo cambia la naturaleza de horas nuevas.

- Abandono individual conserva horas e historial. Si todos abandonan, cerrar el proyecto como ABANDONADO; no reincorporación.

- Los cupos solo advierten. Dashboard y XLSX comparten filtros y cifras; estudiantes únicos se cuentan aparte de participaciones.

### Decisiones que deben cerrarse antes de implementarlas

- Clasificación: confirmar un tipo y varias áreas por proyecto, y si las áreas son independientes del tipo.

- Correcciones: aprobar o ajustar el reverso; definir corrección administrativa posterior al cierre sin admitir horas tardías.

- Saldo previo: fecha de corte, prevención de doble conteo y datos históricos desconocidos.

- Cierre: estados de participaciones al cancelar/finalizar; combinación de personas finalizadas y abandonadas.

- Coordinadores y contactos: persona en varias carreras, relación directa con proyectos y uno o varios responsables.

- Historia académica: conservar carrera/facultad del momento en reportes; campos de contacto obligatorios y fecha final.

- Modalidad y habilidades blandas: definir su significado antes de crear campos o catálogos.

La etapa 00 registra estas decisiones. Se puede avanzar con infraestructura y módulos independientes mientras se resuelven. La corrección de horas no se habilita hasta definir cómo conservar la historia y quién valida el cambio.

## Forma de trabajar con la IA

Conserva este prompt en docs/prompts/00-contexto.md. En una sesión nueva compártelo junto con el repositorio y las reglas vigentes. Después agrega el prompt específico de una sola etapa.

### Prompt maestro reutilizable

```text
Actúa como desarrollador senior y mentor de este portal de horas sociales. Trabaja exclusivamente en la etapa que te indicaré. Antes de editar, lee README, AGENTS.md si existe, docs/reglas-negocio.md, docs/decisiones.md, docs/modelo-datos.md y el estado de Git. Conserva el trabajo existente y no inventes requisitos.
Stack obligatorio: React, TypeScript y Tailwind CSS; API Go; PostgreSQL; desarrollo, compilación, pruebas y aplicación ejecutados en Docker. Mantén un monorepositorio con módulos claros y SQL versionado. Usa Go idiomático y dependencias justificadas.
Primero indica en pocas líneas qué archivos cambiarás y qué validarás; luego implementa dentro de la etapa autorizada. Divide el trabajo en cambios pequeños. Usa datos ficticios en pruebas. Nunca guardes secretos ni datos reales en Git, logs, imágenes o prompts. No borres volúmenes ni reinicies bases para resolver errores. Las migraciones aplicadas no se reescriben.
Implementa backend y frontend cuando la etapa lo requiera; no simules persistencia. Protege los endpoints y conserva la integridad de horas e historia. Una decisión pendiente bloquea solo la parte dependiente: no la inventes.
Ejecuta comprobaciones apropiadas dentro de contenedores. Si una herramienta o prueba no puede ejecutarse, indícalo con su error real. Entrega resumen, archivos cambiados, comandos exactos, resultados esperados y evidencias de las comprobaciones realizadas.
Actualiza OpenAPI, documentación y CHANGELOG cuando corresponda. Prepara un commit pequeño y el siguiente tag Docker local, sin sobrescribir releases. Publica o despliega solo cuando sea parte de la etapa expresamente autorizada. Al finalizar detente y entrégame los criterios manuales de esta etapa; no avances a la siguiente hasta que valide el resultado.
```

### Cierre de cada tarea pequeña

Para tareas extensas, dividir la etapa en cambios revisables: primero contrato y persistencia, después API y por último pantalla e integración. Comprobar después de cada cambio y mantener el portal ejecutable. Las pruebas automáticas las ejecuta la IA; tu aceptación manual comprueba que el resultado cumple el uso real.

Guardar evidencia en docs/validaciones/etapa-XX.md: versión, comando o acción, resultado esperado, resultado obtenido y observaciones. Si falla un criterio, solicitar una corrección acotada y volver a comprobarlo antes de cerrar la etapa.

## Etapas y presupuesto de trabajo

| Etapa | Producto principal | Horas |
|---|---|---|
| 00 | Alcance y decisiones de negocio | 10 |
| 01 | Inicialización y primer arranque en Docker | 10 |
| 02 | Redes persistencia y herramientas de Docker | 18 |
| 03 | Esquema relacional y migraciones | 20 |
| 04 | Base de la API y contrato de integración | 14 |
| 05 | Autenticación y cuenta administradora | 20 |
| 06 | Estructura del frontend con Tailwind CSS | 18 |
| 07 | Catálogos y configuración académica | 14 |
| 08 | Instituciones y contactos | 12 |
| 09 | Proyectos y ciclo de vida | 18 |
| 10 | Expediente y búsqueda de estudiantes | 16 |
| 11 | Asignaciones y abandono de estudiantes | 16 |
| 12 | Movimientos de horas y saldos iniciales | 26 |
| 13 | Dashboard y consultas estadísticas | 18 |
| 14 | Exportación de datos a Excel | 10 |
| 15 | Pruebas integrales rendimiento y recuperación | 16 |
| 16 | Publicación de imágenes y versiones reproducibles | 12 |
| 17 | Preparación de Ubuntu Server y administración | 12 |
| 18 | Ensayo de producción dominio y carga inicial | 14 |
| 19 | Entrega operación y primera versión estable | 6 |
| Total | Proyecto completo | 300 |

Las horas son una estimación de planificación, no una garantía ni horas ya ejecutadas. Incluyen 20 horas de ajuste distribuidas en las etapas 12, 15 y 18. Registrar el tiempo real y revisar el saldo al terminar cada bloque.

Si parte del trabajo ya se completó desde septiembre, se valida y acredita en su etapa; no se repite ni se agregan otras 300 horas. Para mantener el objetivo del 10 de diciembre de 2026, calcular horas pendientes entre jornadas realmente disponibles. Sin horas ya consumidas no corresponde prometer una fecha por etapa.

## Etapa 00 Alcance y decisiones de negocio

Estimación 10 horas  |  Requiere Modelo de datos v0.2 y entrevistas

### Qué vamos a hacer y para qué

Convertir lo conversado en reglas que la IA pueda implementar sin inventar comportamientos. Esta etapa evita rehacer tablas, formularios y reportes después.

Producto de la etapa: docs/alcance.md, docs/reglas-negocio.md, docs/decisiones.md y tablero de tareas.

### Orden de trabajo

- Separar requisitos confirmados de propuestas; trasladar el diccionario de datos a docs/modelo-datos.md.

- Registrar las decisiones pendientes con responsable y etapa afectada. Identificar lo ya realizado y sus horas.

- Inventariar CPU, RAM, disco y arquitectura de la laptop; fijar metas de prueba y respaldo.

### Prompt para la IA

```text
Trabaja únicamente en la etapa 00. Lee el modelo de datos v0.2 y las entrevistas. Documenta módulos, reglas confirmadas, exclusiones y decisiones pendientes. Distingue las 300 horas de desarrollo de las 600 horas de cada estudiante. No conviertas propuestas en requisitos aprobados. Registra especialmente selección de tipos/áreas, corrección de horas, saldos previos, cierre de participaciones, coordinadores e historia académica. Marca qué decisión debe resolverse antes de cada implementación. Propón una lista breve de tareas verificables y registra el equipo disponible. No programes módulos todavía. Entrega los documentos y una lista de decisiones para mi revisión.
```

### Criterios que validaré manualmente

- [ ] Abrir alcance.md: figuran acceso exclusivo ADMIN, módulos administrativos y exportación a Excel; no hay catálogo público.

- [ ] Leer reglas-negocio.md: completar exige 300 internas Y 300 externas; no se compensan categorías.

- [ ] Comprobar que las propuestas del modelo aparecen como pendientes y no como decisiones ya aprobadas.

- [ ] Revisar el tablero: cada tarea tiene etapa y resultado; el trabajo previo aparece identificado, sin contarlo otra vez.

- [ ] Verificar el inventario del servidor y que el plan no presupone capacidad o arquitectura no comprobadas.

Versión al aceptar: Commit de documentación; sin imagen de aplicación todavía.

## Etapa 01 Inicialización y primer arranque en Docker

Estimación 10 horas  |  Requiere Etapa 00

### Qué vamos a hacer y para qué

Crear un repositorio ordenado y una primera versión mínima que pueda ejecutarse sin instalar Go, Node ni PostgreSQL en tu computadora.

Producto de la etapa: Monorepositorio, frontend mínimo con Tailwind, API de prueba y Compose de desarrollo.

### Orden de trabajo

- Inicializar Git y la estructura acordada, conservando cualquier trabajo existente.

- Crear React con TypeScript y Tailwind CSS; crear módulo Go con una ruta de estado.

- Fijar versiones y lockfiles; añadir Dockerfiles, .dockerignore, .gitignore y .env.example sin secretos.

### Prompt para la IA

```text
Implementa solo la etapa 01 sobre el repositorio existente; no lo regeneres si ya contiene avances. Crea apps/api con Go y apps/web con React, TypeScript, Vite y Tailwind CSS. Usa versiones estables compatibles y documenta sus versiones exactas. Instala y compila dependencias dentro de contenedores. Entrega compose.yaml y compose.dev.yaml con web, api y db, además de README para iniciar desde cero. Expón únicamente la web de desarrollo en 127.0.0.1:8080 y conecta /api mediante proxy. La pantalla solo debe mostrar el nombre del portal y estado de conexión; no implementes CRUD. Incluye Dockerfiles multietapa y comandos reproducibles.
```

### Criterios que validaré manualmente

- [ ] Ejecutar el comando de arranque del README: los tres servicios quedan en ejecución.

- [ ] Abrir http://localhost:8080: aparece el portal y se observa una clase Tailwind aplicada.

- [ ] Consultar http://localhost:8080/api/health/live: responde HTTP 200, sin datos personales.

- [ ] Revisar el repositorio: existen go.mod, go.sum y package-lock.json; no hay contraseñas reales.

- [ ] Seguir el README sin usar una instalación local de Go, npm o PostgreSQL.

Versión al aceptar: Git v0.1.0; imágenes locales hs-api:0.1.0 y hs-web:0.1.0.

## Etapa 02 Redes persistencia y herramientas de Docker

Estimación 18 horas  |  Requiere Etapa 01

### Qué vamos a hacer y para qué

Asegurar que los datos sobrevivan al reemplazo de contenedores y que desarrollo y producción tengan configuraciones separadas.

Producto de la etapa: Compose por ambiente, volumen de datos, healthchecks y herramientas auxiliares dockerizadas.

### Orden de trabajo

- Separar redes edge, app y data; PostgreSQL no publica un puerto del host.

- Definir perfiles tools y ops para migraciones, pruebas y tareas administrativas.

- Añadir límites configurables, rotación de logs, apagado ordenado y políticas de reinicio.

### Prompt para la IA

```text
Implementa solo la etapa 02. La base Compose usa imágenes; el override de desarrollo agrega build, montajes y recarga. Prepara compose.prod.yaml sin montajes del código. Usa PostgreSQL 18 con volumen en /var/lib/postgresql y versión exacta fijada; verifica la ruta oficial para esa versión. Configura redes separadas, healthchecks reales y espera de disponibilidad sin sleeps arbitrarios. API y web de producción deben ejecutarse sin root; las imágenes oficiales de base de datos mantienen su mecanismo soportado. Agrega servicios de herramientas para Go, Node, sqlc y migraciones. Protege los archivos de secretos y documenta que un volumen no es un respaldo.
```

### Criterios que validaré manualmente

- [ ] Crear un dato de prueba mediante el procedimiento documentado, ejecutar down sin -v y volver a iniciar: el dato persiste.

- [ ] Ver docker compose ps: db y api llegan al estado healthy mediante comprobaciones reales.

- [ ] Revisar puertos publicados: PostgreSQL y API no están disponibles directamente en el host.

- [ ] Detener db: readiness informa no disponible; al reiniciarlo, la API recupera conexión sin reconstruirla.

- [ ] Revisar la configuración de producción: no usa Vite como servidor público ni monta carpetas del código.

Versión al aceptar: Git v0.2.0; hs-api:0.2.0 y hs-web:0.2.0.

## Etapa 03 Esquema relacional y migraciones

Estimación 20 horas  |  Requiere Etapas 00 y 02

### Qué vamos a hacer y para qué

Transformar el modelo aprobado en tablas reales, conservando un historial de cambios que permita instalar y actualizar la base de forma controlada.

Producto de la etapa: Migraciones SQL con Goose, consultas con sqlc y catálogos iniciales.

### Orden de trabajo

- Crear tablas, PK, FK, restricciones e índices del modelo aprobado.

- Separar credenciales de migración y aplicación; preparar carga inicial repetible.

- Cargar 12 áreas, 26 tipos y catálogos territoriales verificados; datos ficticios solo en desarrollo.

### Prompt para la IA

```text
Implementa solo la etapa 03. Lee las decisiones aprobadas antes de generar SQL. Usa PostgreSQL, pgx/v5, sqlc y Goose; no uses automigración ORM al iniciar. Crea migraciones numeradas, las restricciones del diccionario y roles de mínimo privilegio. No inventes nombres ni relaciones territoriales. Mantén separados catálogos y datos de demostración. Usa uuid para identificadores e INTEGER para horas. Conserva historial y evita cascadas destructivas. Si una cardinalidad aún no está decidida, documenta el bloqueo de esa parte y avanza en las tablas independientes. Agrega pruebas de integración de unicidad, FK y migración desde una base vacía.
```

### Criterios que validaré manualmente

- [ ] Ejecutar migraciones en un volumen desechable nuevo: finalizan correctamente y muestran la versión del esquema.

- [ ] Volver a ejecutar migraciones: no se duplican tablas ni registros de catálogo.

- [ ] Consultar catálogos: hay exactamente 12 áreas y 26 tipos, con las etiquetas revisadas.

- [ ] Probar carnet duplicado y una FK inexistente mediante los comandos de prueba: ambos se rechazan.

- [ ] Con la cuenta de aplicación, intentar crear una tabla: se rechaza; las operaciones de negocio necesarias sí funcionan.

Versión al aceptar: Git v0.3.0; imágenes 0.3.0 y versión de migración registrada.

## Etapa 04 Base de la API y contrato de integración

Estimación 14 horas  |  Requiere Etapa 03

### Qué vamos a hacer y para qué

Establecer cómo se comunican frontend y backend, con errores consistentes, consultas limitadas y comportamiento predecible ante fallos.

Producto de la etapa: API /api/v1, contrato OpenAPI y estructura de módulos Go.

### Orden de trabajo

- Configurar net/http, router chi, logs estructurados y validación de configuración.

- Agregar límites de solicitudes, timeouts, contexto de consultas y pool acotado.

- Definir paginación, filtros, orden permitido y formato común de errores.

### Prompt para la IA

```text
Implementa solo la etapa 04. Organiza cada módulo con handler, service y repository; conserva la lógica de negocio fuera de HTTP. Usa net/http, chi y pgxpool con context y timeouts. Crea /api/health/live y /api/health/ready, un identificador por petición y manejo central de errores sin mostrar SQL ni secretos. Define OpenAPI bajo api/openapi.yaml: rutas, cuerpos, paginación y códigos 400, 401, 403, 404, 409 y 500. Usa SQL parametrizado y listas permitidas para ordenamiento. Propón inicialmente páginas de 25 y máximo 100 elementos. No implementes aún los CRUD completos.
```

### Criterios que validaré manualmente

- [ ] Abrir el contrato OpenAPI y comprobar que describe el formato que devuelve la API de ejemplo.

- [ ] Enviar JSON inválido y un parámetro fuera de rango: se recibe 400 legible, sin error SQL.

- [ ] Pedir una ruta inexistente: responde 404 con el formato común de error.

- [ ] Provocar una indisponibilidad de base: readiness falla y los logs muestran el identificador de petición, sin credenciales.

- [ ] Revisar la configuración: el pool tiene un máximo explícito y el servidor tiene timeouts documentados.

Versión al aceptar: Git v0.4.0; imágenes 0.4.0.

## Etapa 05 Autenticación y cuenta administradora

Estimación 20 horas  |  Requiere Etapa 04

### Qué vamos a hacer y para qué

Permitir el acceso exclusivo del administrador y proteger cada operación del backend, incluso si alguien intenta evitar la pantalla de inicio de sesión.

Producto de la etapa: Login, logout, renovación de sesión, creación inicial segura y recuperación por consola.

### Orden de trabajo

- Crear un único administrador mediante un comando dockerizado; no incluir registro público.

- Añadir sesiones revocables, cookies protegidas, defensa CSRF y límites de intentos.

- Documentar recuperación de acceso sin correo ni dependencia de Gmail.

### Prompt para la IA

```text
Implementa solo la etapa 05. Usa hash de contraseña con Argon2id y una biblioteca JWT mantenida, sin criptografía propia. Propón JWT de acceso de 15 minutos y renovación mediante token opaco rotatorio con duración máxima de sesión de 8 horas. Guárdalos en cookies HttpOnly, Secure en producción y SameSite; protege mutaciones contra CSRF y valida Origin. Añade sesion_admin como tabla técnica mediante migración: usuario_id, identificador de sesión, hash del token de renovación, expiración y revocación; resuelve rotación atómica y reutilización. Valida firma, algoritmo permitido, issuer, audience y expiración del JWT; comprueba la sesión revocable en cada petición. No guardes tokens en localStorage ni logs. Añade bootstrap y reset de administrador con secretos leídos desde un archivo protegido. Expón /api/v1/auth/login, refresh, logout y me, sin registro público. Documenta pruebas con comandos antes de la pantalla de login.
```

### Criterios que validaré manualmente

- [ ] Consultar /api/v1/auth/me sin sesión con el comando documentado: se obtiene 401 y no se entrega información.

- [ ] Enviar login con contraseña correcta e incorrecta mediante los comandos: la primera funciona y la segunda devuelve un error genérico.

- [ ] Ejecutar logout y repetir una petición autenticada anterior: se rechaza porque la sesión está revocada.

- [ ] Revisar cookies: HttpOnly activo y Secure exigido en producción. Probar escritura sin CSRF: se rechaza; HTTPS se revalida en 18.

- [ ] Revisar la demostración de expiración y renovación de la IA; intentar crear un segundo ADMIN activo se rechaza.

Versión al aceptar: Git v0.5.0; imágenes 0.5.0. Añade tabla técnica, sin crear nuevos roles.

## Etapa 06 Estructura del frontend con Tailwind CSS

Estimación 18 horas  |  Requiere Etapas 04 y 05

### Qué vamos a hacer y para qué

Construir una interfaz consistente y fácil de mantener antes de agregar formularios de negocio. Tailwind CSS será obligatorio en todo el portal.

Producto de la etapa: Layout administrativo, navegación, componentes comunes y cliente de API.

### Orden de trabajo

- Crear módulos por funcionalidad y componentes accesibles de formularios y tablas.

- Integrar sesión, navegación protegida, carga, vacío, error y confirmación.

- Configurar React Router y TanStack Query; invalidar datos después de guardar.

### Prompt para la IA

```text
Implementa solo la etapa 06 con React, TypeScript estricto y Tailwind CSS instalado mediante el plugin oficial de Vite. No uses Tailwind por CDN ni introduzcas Bootstrap. Crea layout, menú, encabezado y componentes reutilizables: campos con etiqueta, tabla paginada, selector, diálogo y mensajes. Organiza features por módulo. Usa un cliente de API central con cookies y CSRF; coordina una sola renovación de sesión ante peticiones concurrentes y evita duplicar escrituras. Crea pantallas iniciales de login, dashboard, estudiantes, proyectos, instituciones y configuración, sin simular registros reales. En formularios, conserva lo escrito ante errores y advierte antes de perder cambios.
```

### Criterios que validaré manualmente

- [ ] Navegar por los módulos: la sesión y el menú funcionan sin recargar toda la página.

- [ ] Usar teclado para abrir un diálogo y cerrarlo: el foco vuelve al control original.

- [ ] Probar a 1366 px y 390 px de ancho: navegación y acciones principales siguen accesibles.

- [ ] Simular un error de API: aparece un mensaje claro y el formulario conserva lo escrito.

- [ ] Revisar la compilación: TypeScript y estilos Tailwind se generan dentro de Docker sin errores.

Versión al aceptar: Git v0.6.0; imágenes 0.6.0.

## Etapa 07 Catálogos y configuración académica

Estimación 14 horas  |  Requiere Etapa 06

### Qué vamos a hacer y para qué

Permitir que el administrador mantenga las opciones de los formularios y los coordinadores sin modificar código.

Producto de la etapa: CRUD administrativo de catálogos y asociaciones carrera coordinador.

### Orden de trabajo

- Implementar facultades, carreras, género, coordinadores, áreas y tipos.

- Agregar búsqueda y desactivación; conservar referencias históricas.

- Usar el catálogo territorial como selección jerárquica verificada.

### Prompt para la IA

```text
Implementa solo la etapa 07 como flujo completo API y frontend. Agrega catálogos académicos y de clasificación, coordinadores y sus asociaciones por carrera. Permite crear, editar y desactivar según las decisiones aprobadas; no elimines opciones ya utilizadas. Valida duplicados sin distinguir mayúsculas y mantiene tipo y área separados. Implementa departamento, municipio y distrito como selección jerárquica sobre datos verificados. Las opciones inactivas no deben ofrecerse en altas, pero sí verse en registros históricos. Actualiza OpenAPI y agrega pruebas de integridad. No implementes modalidad ni habilidades blandas sin definición.
```

### Criterios que validaré manualmente

- [ ] Crear una facultad y su carrera: al seleccionar la carrera se reconoce la facultad correcta.

- [ ] Asignar dos coordinadores a una carrera: ambos aparecen sin crear cuentas de acceso.

- [ ] Intentar duplicar un nombre cambiando solo mayúsculas: aparece validación y no se duplica.

- [ ] Desactivar una opción usada: permanece visible en el registro anterior y no se ofrece para un alta nueva.

- [ ] Elegir departamento y municipio: los distritos ofrecidos pertenecen a esa selección.

Versión al aceptar: Git v0.7.0; imágenes 0.7.0.

## Etapa 08 Instituciones y contactos

Estimación 12 horas  |  Requiere Etapa 07

### Qué vamos a hacer y para qué

Registrar las organizaciones y consultar sus proyectos sin mezclar el contacto institucional con el responsable de cada proyecto.

Producto de la etapa: Listado, alta, edición, detalle y desactivación de instituciones.

### Orden de trabajo

- Crear endpoints y pantallas de instituciones con búsqueda y paginación.

- Validar contactos y dirección; conservar teléfonos como texto.

- Mostrar los proyectos asociados sin permitir borrados destructivos.

### Prompt para la IA

```text
Implementa solo la etapa 08 como módulo completo de instituciones. Utiliza los campos aprobados y un responsable principal, salvo que decisiones.md indique varios contactos. Agrega búsqueda, paginación, alta, edición, detalle y desactivación. Conserva teléfonos y extensiones como texto y valida correo en el backend. La dirección de la sede debe ser independiente de la ejecución de proyectos. Desde el detalle, deja preparada la consulta de proyectos relacionados, sin inventar datos. Protege todas las rutas y evita perder la información ingresada cuando falle una validación.
```

### Criterios que validaré manualmente

- [ ] Crear una institución con teléfono y extensión: al reabrir conserva exactamente los valores.

- [ ] Editar su responsable: no cambia el responsable específico de proyectos existentes.

- [ ] Buscar por nombre y cambiar de página: el filtro se conserva.

- [ ] Ingresar correo inválido: se impide guardar y se señala el campo correspondiente.

- [ ] Intentar eliminar una institución con proyectos: no se pierden vínculos ni antecedentes.

Versión al aceptar: Git v0.8.0; imágenes 0.8.0.

## Etapa 09 Proyectos y ciclo de vida

Estimación 18 horas  |  Requiere Etapa 08 y decisiones de clasificación cerradas

### Qué vamos a hacer y para qué

Administrar los proyectos y sus estados con una sola institución y una ubicación física, conservando el historial de cada cambio.

Producto de la etapa: CRUD de proyectos, filtros, historial y cierre controlado.

### Orden de trabajo

- Registrar clasificación, ubicación, cupo orientativo y contacto del proyecto.

- Implementar transiciones desde ACTIVO y fecha informada al finalizar.

- Mostrar historial y coordinar bloqueos con futuras asignaciones y horas.

### Prompt para la IA

```text
Implementa solo la etapa 09. Crea proyectos con institución única, ubicación física única, naturaleza EXTERNA por defecto y la cardinalidad de tipo/áreas aprobada. No agregues Publicar. La fecha final no se exige en el alta y se captura al finalizar. Solo admite ACTIVO hacia ABANDONADO, CANCELADO o FINALIZADO; no reapertura. Guarda estado y evento histórico en la misma transacción, bloqueando la fila de proyecto. Conserva contactos independientes de la institución. Los cierres impedirán nuevas asignaciones y horas; prepara la integración con participaciones sin inventar su cierre. Actualiza contrato y pantallas.
```

### Criterios que validaré manualmente

- [ ] Crear un proyecto sin fecha final: se guarda ACTIVO y EXTERNA por defecto.

- [ ] Seleccionar institución y ubicación: el proyecto tiene una sola de cada una.

- [ ] Cambiar a FINALIZADO con fecha válida: el detalle e historial muestran el cierre y su responsable.

- [ ] Intentar volver de FINALIZADO a ACTIVO: la operación se rechaza.

- [ ] Filtrar por estado, institución y clasificación: aparecen resultados coherentes sin proyectos duplicados.

Versión al aceptar: Git v0.9.0; imágenes 0.9.0.

## Etapa 10 Expediente y búsqueda de estudiantes

Estimación 16 horas  |  Requiere Etapas 07 y 09

### Qué vamos a hacer y para qué

Tener una ficha única por estudiante, encontrarla por carnet y preparar la vista de proyectos y horas que luego completará el libro de movimientos.

Producto de la etapa: Alta, edición, búsqueda y detalle del estudiante.

### Orden de trabajo

- Conservar carnet como texto y vincular carrera con facultad.

- Agregar filtros académicos y vista del expediente.

- Mostrar áreas de participaciones y resumen de horas sin cifras ficticias.

### Prompt para la IA

```text
Implementa solo la etapa 10. Crea alta y edición manual de estudiantes con carnet único como texto, nombres, apellidos, carrera y género según el modelo aprobado. La facultad se obtiene de la carrera. Agrega búsqueda por carnet y nombre, filtros por facultad/carrera y paginación. Diseña el expediente con resumen de internas/externas y lista desplegable por proyecto; mientras no exista el libro de horas, muestra estados vacíos correctos. No guardes un total editable en estudiante. Si se permite cambiar carrera, aplica la decisión de historia académica y documenta su efecto en reportes.
```

### Criterios que validaré manualmente

- [ ] Registrar el carnet 00200022: conserva los ceros al consultar y editar.

- [ ] Intentar crear el mismo carnet: se rechaza sin sobrescribir el expediente anterior.

- [ ] Buscar por carnet y filtrar por facultad: se encuentra al estudiante esperado.

- [ ] Cambiar carrera: la facultad actual se actualiza según la relación académica.

- [ ] Abrir un estudiante sin proyectos: se muestra un estado vacío, no filas ni horas inventadas.

Versión al aceptar: Git v0.10.0; imágenes 0.10.0.

## Etapa 11 Asignaciones y abandono de estudiantes

Estimación 16 horas  |  Requiere Etapas 09 y 10; cierre individual definido

### Qué vamos a hacer y para qué

Vincular estudiantes y proyectos, permitiendo varias participaciones y separando el abandono individual del abandono del proyecto completo.

Producto de la etapa: Asignación, salida, historial y abandono automático total.

### Orden de trabajo

- Validar una participación por estudiante/proyecto y bloquear reinscripción.

- Advertir exceso de cupo sin impedir guardar.

- Aplicar abandono total y cierre individual de forma atómica.

### Prompt para la IA

```text
Implementa solo la etapa 11. Permite asignar varios estudiantes a proyectos ACTIVOS y que una persona tenga varios proyectos. Mantén unicidad estudiante/proyecto y prohíbe reincorporación tras abandono. Superar alumnos_solicitados genera advertencia, no bloqueo. Al abandonar, conserva participación e historial. Bloquea el proyecto durante la operación; si tiene al menos una participación y todas quedan ABANDONADAS, abandona el proyecto en la misma transacción y cierra nuevas asignaciones. No confundas cero activos con todos abandonados. Aplica los cierres individuales ya aprobados. Añade pruebas de concurrencia con altas y cierres.
```

### Criterios que validaré manualmente

- [ ] Asignar el mismo estudiante a dos proyectos: aparecen ambas participaciones.

- [ ] Inscribir una sexta persona en un proyecto que solicita cinco: se advierte y se permite guardar.

- [ ] Abandonar uno de dos alumnos: el proyecto sigue ACTIVO para el otro.

- [ ] Abandonar al último y comprobar que todos abandonaron: el proyecto queda ABANDONADO y rechaza nuevos alumnos.

- [ ] Intentar reinscribir a quien abandonó o duplicar una asignación: se rechaza y su historial permanece.

Versión al aceptar: Git v0.11.0; imágenes 0.11.0.

## Etapa 12 Movimientos de horas y saldos iniciales

Estimación 26 horas  |  Requiere Etapa 11; saldo previo e historia académica definidos

### Qué vamos a hacer y para qué

Implementar el núcleo del sistema: horas enteras que se suman por estudiante y proyecto, con trazabilidad y protección contra duplicados.

Producto de la etapa: Libro de horas, saldos iniciales, cálculo de metas y detalle por participación.

### Orden de trabajo

- Agregar horas con fecha administrativa y captura técnica separadas.

- Usar idempotencia y transacciones; conservar la categoría de cada movimiento.

- Aplicar correcciones únicamente si se aprueba su mecanismo.

### Prompt para la IA

```text
Implementa solo la etapa 12. Agrega movimientos INTEGER positivos y saldos previos separados por categoría; no sobrescribas totales. Valida enteros antes de convertir y rechaza decimales. Solo acredita con proyecto y participación ACTIVOS; bloquea el proyecto durante la transacción. Copia naturaleza vigente y los datos académicos históricos aprobados. Añade por migración solicitud_id UUID único para evitar doble registro al reintentar la misma operación; si cambia el cuerpo para esa clave, devuelve conflicto. La UI reutiliza la clave al reintentar. Muestra por proyecto y total 300 internas y 300 externas sin compensarlas. No admitas horas tardías. Si correcciones sigue pendiente, no expongas edición, borrado ni reversos: registra la decisión pendiente.
```

### Criterios que validaré manualmente

- [ ] Agregar 20 y luego 15 horas: se ven dos líneas y el total es 35.

- [ ] Registrar 180 internas y 225 externas: muestra 405 realizadas y 195 pendientes.

- [ ] Probar 350 internas y 250 externas: no declara cumplimiento; faltan 50 externas.

- [ ] Registrar 40 externas, cambiar naturaleza a interna y agregar 10: conserva 40 externas y suma 10 internas.

- [ ] Repetir una solicitud con la misma clave: queda un solo movimiento. Probar 1.5 horas y un proyecto cerrado: ambos se rechazan.

Versión al aceptar: Git v0.12.0; imágenes 0.12.0. Incluye 4 h de ajuste del flujo crítico.

## Etapa 13 Dashboard y consultas estadísticas

Estimación 18 horas  |  Requiere Etapa 12

### Qué vamos a hacer y para qué

Convertir los registros en indicadores útiles sin duplicar personas ni horas al combinar proyectos, áreas y facultades.

Producto de la etapa: Dashboard web y consultas por período, academia, género y territorio.

### Orden de trabajo

- Separar proyectos, estudiantes únicos, participaciones y horas.

- Filtrar actividad por fecha_horas y tratar saldos previos aparte.

- Implementar agregaciones SQL e índices medidos, con estados vacíos y errores.

### Prompt para la IA

```text
Implementa solo la etapa 13. Construye consultas SQL y dashboard por período, facultad, carrera, género, institución, ubicación, tipo y área. Participaron significa estudiantes únicos con horas realizadas en el período. No atribuyas saldos iniciales sin historia a un año o territorio. No multipliques movimientos al unir varias áreas: usa EXISTS o deduplica antes de sumar. Distingue avance acumulado actual de actividad del período; no presentes completados en el año sin una regla aprobada para esa fecha. Mantén clasificación histórica solo si se aprobó. Ofrece tarjetas claras y tablas que permitan comprobar cada cifra.
```

### Criterios que validaré manualmente

- [ ] Un estudiante con horas en dos proyectos aparece como una persona y dos participaciones.

- [ ] Cambiar el período: las horas responden a fecha_horas, no al instante de captura.

- [ ] Agregar saldo inicial sin fecha: cambia el avance acumulado, no la actividad anual conocida.

- [ ] Filtrar un proyecto con varias áreas: sus horas no se multiplican.

- [ ] Comparar las cifras del dashboard con un conjunto pequeño calculado a mano: coinciden.

Versión al aceptar: Git v0.13.0; imágenes 0.13.0.

## Etapa 14 Exportación de datos a Excel

Estimación 10 horas  |  Requiere Etapa 13

### Qué vamos a hacer y para qué

Permitir que el administrador descargue información verificable con los mismos filtros del dashboard.

Producto de la etapa: Exportación XLSX protegida, con filtros y tipos de celdas correctos.

### Orden de trabajo

- Generar XLSX en Go con Excelize y consultas compartidas con reportes.

- Mantener carnet como texto y horas como números.

- Limitar exportación y memoria; evitar fórmulas introducidas como texto de usuario.

### Prompt para la IA

```text
Implementa solo la etapa 14. Exporta XLSX mediante Excelize desde un endpoint autenticado. Reutiliza las reglas de filtros y agregaciones del dashboard. Incluye hojas de resumen, detalle y filtros aplicados, con fecha de generación. Guarda carnet como texto y horas como números; utiliza valores de texto para contenido de usuarios, sin convertirlo en fórmulas. Usa escritura por flujo cuando sea necesario y un límite inicial documentado de 50000 filas; si se supera, solicita reducir filtros con un mensaje claro. Cancela la consulta si el cliente desconecta. No cargues todos los registros de la base en el navegador.
```

### Criterios que validaré manualmente

- [ ] Aplicar un filtro y exportar: el archivo abre en Excel o LibreOffice sin advertencias de reparación.

- [ ] Comparar resumen exportado y dashboard con el mismo filtro: las cifras coinciden.

- [ ] Revisar un carnet con ceros: se conserva como texto; las horas permiten sumarse como números.

- [ ] Exportar un nombre que inicia con =: se ve como texto y no ejecuta una fórmula.

- [ ] Probar sin sesión y con exceso del límite: se rechaza con un resultado controlado, sin congelar el portal.

Versión al aceptar: Git v0.14.0; imágenes 0.14.0.

## Etapa 15 Pruebas integrales rendimiento y recuperación

Estimación 16 horas  |  Requiere Etapa 14

### Qué vamos a hacer y para qué

Detectar fallos importantes antes de publicar y comprobar en el equipo disponible que el sistema maneja un volumen de prueba acordado.

Producto de la etapa: Suite crítica automatizada, informe de rendimiento y ensayo de restauración.

### Orden de trabajo

- Probar autenticación, horas, abandono, cierres, idempotencia y concurrencia.

- Medir listados, dashboard y exportación con datos sintéticos aislados.

- Restaurar un backup y verificar datos; corregir errores prioritarios.

### Prompt para la IA

```text
Implementa solo la etapa 15. Ejecuta pruebas Go y de integración con PostgreSQL real, más recorridos Playwright dockerizados para login, estudiante, proyecto y horas. Usa un ambiente exclusivo hs-test sin datos reales. Genera datos sintéticos repetibles y registra hardware, tamaño y latencias. Propón como referencia 10000 estudiantes, 1000 proyectos y 200000 movimientos; valida que sea adecuado al equipo. Comprueba transacciones concurrentes de cierre y acreditación, idempotencia y no duplicación en reportes. Mide objetivos de la tabla de calidad y ensaya backup más restore en otro volumen. Entrega comandos, resultados reproducibles y fallos pendientes; no declares pruebas no ejecutadas.
```

### Criterios que validaré manualmente

- [ ] Ejecutar el comando de pruebas: termina correctamente y el informe indica qué escenarios corrieron.

- [ ] Completar manualmente el recorrido login → institución → proyecto → estudiante → asignación → horas → exportación.

- [ ] Leer el informe de carga: identifica hardware y datos; compara resultados con los objetivos acordados.

- [ ] Restaurar en un volumen nuevo: coinciden carnets, participaciones y acumulados de la muestra.

- [ ] Verificar que no quedan fallos críticos o altos abiertos antes de continuar con la publicación.

Versión al aceptar: Git v0.15.0; imágenes 0.15.0. Incluye 10 h para correcciones prioritarias.

## Etapa 16 Publicación de imágenes y versiones reproducibles

Estimación 12 horas  |  Requiere Etapa 15 y acceso a GitHub

### Qué vamos a hacer y para qué

Automatizar los controles y guardar las imágenes aprobadas para instalar exactamente la misma versión sin recompilar en la laptop.

Producto de la etapa: GitHub Actions, imágenes en GHCR y manifiesto de cada release.

### Orden de trabajo

- Automatizar compilación, pruebas, migraciones de prueba y revisión de dependencias.

- Publicar api, web y migrator desde tags Git aprobados.

- Registrar commit, tags, digests, plataforma y compatibilidad del esquema.

### Prompt para la IA

```text
Implementa solo la etapa 16. Configura GitHub Actions para verificar Go, TypeScript, pruebas críticas, migraciones y build de imágenes. En pull requests de fuentes no confiables no publiques imágenes ni expongas secretos. Publica en GHCR únicamente al recibir un tag de release aprobado; usa GITHUB_TOKEN con los permisos mínimos. Crea etiquetas SemVer y sha del commit, con etiquetas OCI y digest registrado. Nunca sobrescribas un tag de release. Mantén api, web y migrator coordinados por versión y crea deploy/releases/<version>.env y manifest.json sin secretos. Promueve los mismos digests probados; no reconstruyas en producción. Fija acciones y dependencias y comprueba acceso de lectura desde el servidor.
```

### Criterios que validaré manualmente

- [ ] Abrir una PR de prueba con un fallo de compilación: la comprobación falla y no publica imágenes.

- [ ] Crear un tag aprobado: aparecen api, web y migrator con la versión esperada en GHCR.

- [ ] Comparar manifest.json: commit, digest y versión del esquema corresponden al release.

- [ ] Instalar las imágenes en un ambiente limpio: funciona sin compilar en ese equipo.

- [ ] Comprobar que el servidor puede descargar paquetes privados sin usar un token con permiso de escritura.

Versión al aceptar: Git v0.16.0; imágenes 0.16.0 en GHCR.

## Etapa 17 Preparación de Ubuntu Server y administración

Estimación 12 horas  |  Requiere Etapas 02 y 16; laptop disponible

### Qué vamos a hacer y para qué

Preparar el equipo físico para operar de forma continua, con administración remota protegida y conservación de los datos.

Producto de la etapa: Ubuntu Server, Docker Engine, Portainer, Cockpit y respaldo fuera del equipo.

### Orden de trabajo

- Documentar instalación, disco, energía, red, actualizaciones y arranque automático.

- Configurar Portainer en Docker y Cockpit como servicio del host.

- Establecer firewall, credenciales técnicas, monitoreo básico y respaldo programado.

### Prompt para la IA

```text
Implementa solo la etapa 17 y entrega un procedimiento verificable antes de aplicarlo al equipo. Instala Ubuntu Server LTS y Docker Engine con Compose desde fuentes oficiales. La aplicación, PostgreSQL, herramientas y Portainer van en contenedores; Cockpit administra Ubuntu y se instala en el host. Configura usuario técnico, SSH con claves, firewall, actualizaciones planificadas y pruebas de reinicio. Comprueba disco, temperatura, conexión y política al cerrar la tapa para evitar suspensión. Protege Portainer y Cockpit mediante LAN o acceso remoto autenticado; no los publiques abiertos. Define backups diarios ejecutados en contenedor y copia cifrada fuera de la laptop, retención y alertas de fallo. Usa Compose versionado como fuente de configuración.
```

### Criterios que validaré manualmente

- [ ] Reiniciar la laptop: Docker y los servicios previstos vuelven a iniciar y los datos permanecen.

- [ ] Revisar la política de energía: cerrar la tapa según la prueba acordada no suspende el servicio.

- [ ] Ingresar a Portainer y Cockpit con credenciales técnicas: muestran su ámbito correcto.

- [ ] Comprobar que PostgreSQL, Portainer y Cockpit no son accesibles públicamente sin el control definido.

- [ ] Localizar un backup reciente fuera de la laptop y verificar el registro de ejecución y espacio disponible.

Versión al aceptar: Git v0.17.0; imágenes de aplicación 0.17.0 y herramientas fijadas por versión/digest.

## Etapa 18 Ensayo de producción dominio y carga inicial

Estimación 14 horas  |  Requiere Etapas 16 y 17; dominio y decisiones funcionales cerrados

### Qué vamos a hacer y para qué

Validar el despliegue real con una versión candidata, comprobar el túnel y cargar la información inicial sin duplicar horas existentes.

Producto de la etapa: Versión candidata, dominio HTTPS, carga inicial conciliada y ensayo de reversión.

### Orden de trabajo

- Conectar cloudflared dockerizado a web:8080; proteger paneles técnicos aparte.

- Ejecutar el procedimiento de migración y pruebas sobre un entorno aislado.

- Conciliar saldos previos y ensayar actualización, backup y recuperación.

### Prompt para la IA

```text
Implementa solo la etapa 18. Despliega primero la candidata con imágenes por digest y volumen independiente, por ejemplo hs-staging. Ejecuta cloudflared como contenedor conectado a edge y usa http://web:8080 como origen; localhost dentro del túnel no es el host ni otro contenedor. Configura dominio y HTTPS sin exponer db. Revisa cookies, cabeceras, no-cache en datos privados y acceso protegido. Ensaya backup, migración de una sola ejecución, readiness y smoke tests. Prepara carga inicial manual o por archivo controlado solo si se autoriza: conserva fuentes, saldos y fecha de corte sin inventar historia. Ensaya rollback compatible; no uses down -v ni migres automáticamente al iniciar cada réplica.
```

### Criterios que validaré manualmente

- [ ] Abrir el dominio desde otra red: HTTPS funciona y solo muestra acceso hasta iniciar sesión.

- [ ] Inspeccionar cookies y navegar por rutas internas: sesión, CSRF y recarga de página funcionan en el dominio.

- [ ] Comparar una muestra de alumnos con la fuente anterior: internas, externas y proyectos coinciden, sin doble carga.

- [ ] Aplicar el procedimiento de actualización y verificar que ejecuta migraciones antes de habilitar la nueva API.

- [ ] Volver a una imagen anterior compatible en el ensayo: los datos posteriores permanecen; el resultado queda registrado.

Versión al aceptar: Git v1.0.0-rc.1; imágenes 1.0.0-rc.1. Incluye 6 h para ajustes de despliegue.

## Etapa 19 Entrega operación y primera versión estable

Estimación 6 horas  |  Requiere Etapa 18 aceptada

### Qué vamos a hacer y para qué

Entregar una versión que el administrador pueda usar y que tú puedas mantener, actualizar y recuperar con instrucciones claras.

Producto de la etapa: Release 1.0.0, manuales, capacitación y acta de aceptación.

### Orden de trabajo

- Cerrar la lista de aceptación con el administrador y documentar pendientes futuros.

- Promover la candidata validada sin recompilar y registrar su manifiesto.

- Entregar operación, respaldo, recuperación, actualización y control de horas del desarrollo.

### Prompt para la IA

```text
Implementa solo la etapa 19. Comprueba todos los criterios de las etapas y prepara el release 1.0.0 con los mismos digests ya validados en la candidata, sin cambios ocultos. Entrega manual de usuario, manual técnico, inventario de versiones y procedimiento de restauración y actualización. Documenta responsables, ubicación de respaldos y cómo recuperar acceso sin compartir secretos en el repositorio. Prepara una capacitación con crear proyecto, registrar alumno, agregar horas y exportar. Incluye acta de aceptación con resultados, limitaciones conocidas y funciones futuras fuera de alcance. Registra horas reales; no declares terminadas etapas que todavía no se validaron.
```

### Criterios que validaré manualmente

- [ ] El administrador realiza el recorrido principal con el manual sin depender de la IA.

- [ ] El acta identifica la versión, las validaciones realizadas y los pendientes aceptados.

- [ ] El release 1.0.0 contiene los mismos digests de la candidata aprobada.

- [ ] Los manuales permiten localizar un respaldo y seguir la restauración ensayada.

- [ ] Git, imágenes, esquema y registro de horas quedan identificados; el total documentado refleja trabajo real.

Versión al aceptar: Git v1.0.0; imágenes 1.0.0 con los digests aceptados de la candidata.

## Comandos de uso diario y verificación

La etapa 01 y las herramientas de la etapa 02 deben crear estos nombres y scripts. Son el contrato de operación del proyecto; al implementar se comprueban y se documenta cualquier ajuste. Los comandos usan Bash en Ubuntu, WSL2 o una terminal equivalente; no requieren Go ni npm instalados fuera de Docker.

```sh
dcdev() {
  docker compose -p hs-dev \
    -f compose.yaml -f compose.dev.yaml "$@"
}
dcdev up -d --build
dcdev ps
dcdev logs --tail=100 api
dcdev --profile tools run --rm go-tools go test ./...
dcdev --profile tools run --rm node-tools npm run check
dcdev down
```

go-tools trabaja en /workspace/apps/api. node-tools trabaja en /workspace/apps/web; npm run check debe ejecutar TypeScript y lint. Las pruebas Go con PostgreSQL y Playwright usarán hs-test y una base de datos desechable identificada; no la base de hs-dev ni hs-prod.

### Resultados esperados

- up inicia el ambiente; ps muestra servicios y salud.

- logs muestra eventos recientes sin contraseñas, tokens o información innecesaria del estudiante.

- Las comprobaciones salen con código 0. Si fallan, el informe conserva el error y el criterio queda pendiente.

- down retira contenedores del ambiente pero conserva el volumen. No agregar -v para resolver problemas.

### Cuando una etapa falla

```text
Corrige únicamente el criterio [identificador] de la etapa [XX]. Resultado esperado: [...]. Resultado observado: [...]. Comando o pasos: [...]. Revisa la causa antes de cambiar código. No reinicies la base ni borres datos. Explica el cambio mínimo, aplica la corrección y repite las comprobaciones relacionadas. Entrega evidencia y conserva los demás comportamientos aceptados.
```

## Actualización respaldo y vuelta a una versión anterior

El procedimiento real se implementa y ensaya en las etapas 16 a 18. Antes de actualizar se verifica que todas las imágenes están descargadas, existe respaldo recuperable y se conocen las migraciones y su compatibilidad.

### Secuencia de actualización

- Entrar en mantenimiento y bloquear nuevas escrituras; dejar terminar operaciones en curso.

- Validar configuración sin imprimir secretos, descargar imágenes exactas y tomar un respaldo consistente.

- Aplicar migraciones con un único proceso; si falla, no habilitar la nueva API.

- Iniciar web y API, revisar readiness y comprobar login, lectura y una operación controlada.

- Salir de mantenimiento y registrar manifiesto, esquema, respaldo y resultado.

```sh
dcprod() {
  docker compose -p hs-prod \
    --env-file deploy/releases/1.0.0.env \
    -f compose.yaml -f compose.prod.yaml "$@"
}
dcprod config --quiet
dcprod pull web api migrator
dcprod --profile ops run --rm backup
dcprod --profile ops run --rm migrator up
dcprod up -d web api
dcprod ps
```

Este bloque ilustra la secuencia de Compose una vez implementados los servicios. El modo mantenimiento, la verificación del respaldo y los smoke tests se gestionarán en scripts de operación; no ejecutar solo el bloque sin esos pasos.

### Cómo recuperar el servicio

Si el esquema sigue siendo compatible, volver a los digests anteriores y comprobar el servicio, conservando las nuevas escrituras de la base. No ejecutar migraciones down automáticamente. Priorizar migraciones aditivas: agregar primero, adaptar el código y retirar columnas solo en una versión posterior verificada.

Si una migración impide volver, aplicar una corrección hacia adelante. Restaurar un respaldo es recuperación de datos, requiere detener escrituras y puede perder cambios desde el respaldo. Se decide y documenta expresamente; una restauración no se presenta como rollback sin pérdida.

## Objetivos de calidad e hitos de aceptación

| Aspecto | Objetivo propuesto |
|---|---|
| Datos de prueba | Referencia propuesta: 10000 estudiantes, 1000 proyectos y 200000 movimientos, solo sintéticos. |
| Carga de referencia | Cinco clientes concurrentes durante un ensayo controlado; no representa la demanda real conocida. |
| Listados | Objetivo inicial p95 menor de 500 ms para páginas de 25 filas, medido dentro de la red local. |
| Dashboard | Objetivo inicial p95 menor de 1500 ms con filtros; medir SQL e índices si no se alcanza. |
| Exportación | Hasta 50000 filas: objetivo menor de 30 s; registrar consumo y limitar memoria según el equipo. |
| Recuperación | Objetivo inicial RPO 24 h y RTO 4 h. Acordar con el administrador y comprobar mediante restauración. |
| Salida a producción | Sin defectos críticos o altos abiertos; respaldo restaurado y recorrido principal aceptado. |

p95 es el tiempo por debajo del cual queda el 95 % de las respuestas medidas. RPO es la pérdida de datos máxima objetivo; RTO es el tiempo objetivo para recuperar el servicio. Son metas para acordar, no garantías de la laptop; medir antes de aprobar y dejar constancia si se ajustan.

### Hitos que conviene demostrar

- Después de 02: ambiente dockerizado que reinicia y conserva datos.

- Después de 06: base real y acceso protegido con interfaz Tailwind.

- Después de 12: ciclo completo de estudiante, proyecto, participación y acumulación de horas.

- Después de 15: reportes y Excel conciliados, pruebas críticas y respaldo restaurado.

- Después de 19: dominio en operación, release estable, capacitación y entrega aceptada.

Escalabilidad inicial significa mantener estructura modular, SQL correcto, paginación, índices y capacidad de medir. La laptop seguirá siendo un único punto de operación: los respaldos fuera del equipo y el procedimiento de recuperación forman parte de la entrega.

## Fuentes y control del plan

Base funcional: Modelo Preliminar Horas Sociales versión 0.2 del 23 de octubre de 2026 y respuestas de la entrevista. La elección de Go, las herramientas y la secuencia son la propuesta técnica de este plan; las decisiones funcionales pendientes se registran antes de su implementación.

Documentación técnica oficial consultada el 23 de octubre de 2026. Verificar compatibilidad y fijar versiones exactas al iniciar. Los prompts describen trabajo que deberá ejecutarse y validarse en el repositorio; esta guía no afirma que las etapas ya estén implementadas.

Go HTTP  https://pkg.go.dev/net/http

Router chi  https://github.com/go-chi/chi

Go y transacciones  https://go.dev/doc/database/execute-transactions

pgx y pool  https://github.com/jackc/pgx

sqlc con pgx  https://docs.sqlc.dev/en/latest/guides/using-go-and-pgx.html

Goose  https://pressly.github.io/goose/

Tailwind CSS con Vite  https://tailwindcss.com/docs/installation/using-vite

Compose en producción  https://docs.docker.com/compose/how-tos/production/

PostgreSQL en Docker  https://docs.docker.com/guides/postgresql/

GitHub Container Registry  https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry

Digests de imágenes  https://docs.docker.com/dhi/explore/security-concepts/digests/

Cloudflare Tunnel  https://developers.cloudflare.com/cloudflare-one/networks/connectors/cloudflare-tunnel/

Sesiones seguras OWASP  https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html

Contraseñas OWASP  https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html

Excelize  https://xuri.me/excelize/en/

### Registro por etapa

Etapa y versión: __________   Fecha: __________   Horas reales: __________

Commit: __________   Imágenes y digests: __________   Migración: __________

Criterios aprobados: __________   Pendientes: __________   Validado por: __________

Mantener este registro en docs/validaciones y actualizar el presupuesto restante. Si cambia el alcance, ajustar primero reglas, plan y criterios; luego solicitar a la IA el cambio concreto.
