# Portal de Horas Sociales: lógica de negocio y contexto del proyecto

> **Publicación en el repositorio: 29 de octubre de 2026, issue #117, solicitada por el propietario.** Se conserva esta consolidación con su corte histórico del 25 de octubre; las afirmaciones de pendientes/avance de su cuerpo no describen automáticamente el estado actual. La publicación no aprueba nuevas tablas, políticas ni propuestas antiguas.
>
> Para continuar, prevalecen [decisiones vigentes](DECISIONES_VIGENTES.md), [contrato de acceso aceptado](CONTRATO_ACCESO.md), [guía de relevo](RELEVO_OTRA_IA.md), [identidad Archivo](IDENTIDAD_VISUAL.md) y el código/estado actualizado de GitHub. En particular: JWT de dos horas, sin refresh/recuperación/revocación; un tipo y un área por proyecto; modalidad es un solo concepto seleccionable/creable; coordinador simple según el acuerdo registrado en el relevo. No ejecutar las referencias históricas a sesiones revocables ni recuperar obligatoriamente el modelo v0.2: se autorizó diseñar un modelo nuevo por bloques, con M10 aceptado y los demás sujetos a revisión.
>
> Se adaptaron enlaces locales para que funcionen fuera del equipo original. Las fuentes antiguas ausentes se identifican como tales; no se afirma haberlas incorporado.

**Proyecto:** ProyectoSocial Nativi / Portal de Horas Sociales

**Documento:** Logica_Negocio.md

**Versión de esta consolidación:** 1.0

**Fecha de corte:** 25 de octubre de 2026, America/El_Salvador

**Propósito:** reunir el contexto funcional, las decisiones, la planificación y los pendientes recuperados para que una persona o una IA pueda continuar el proyecto sin depender del historial de conversaciones.

> Este documento consolida las fuentes disponibles. No sustituye un modelo de datos que no se haya recuperado ni convierte una propuesta técnica en un acuerdo del usuario. El modelo preliminar v0.2 y parte de las respuestas originales a la entrevista se mencionan en el plan, pero no se encontraron completos entre las fuentes consultadas. Sus campos, tablas y aprobaciones no se reconstruyen por intuición.

## Índice

1. [Cómo interpretar los acuerdos y las fuentes](#1-cómo-interpretar-los-acuerdos-y-las-fuentes)
2. [Propósito, alcance y exclusiones](#2-propósito-alcance-y-exclusiones)
3. [Personas, acceso y sesiones](#3-personas-acceso-y-sesiones)
4. [Glosario del negocio](#4-glosario-del-negocio)
5. [Catálogos, estructura académica y territorio](#5-catálogos-estructura-académica-y-territorio)
6. [Instituciones y contactos](#6-instituciones-y-contactos)
7. [Proyectos y ciclo de vida](#7-proyectos-y-ciclo-de-vida)
8. [Estudiantes y expediente académico](#8-estudiantes-y-expediente-académico)
9. [Participaciones y abandono](#9-participaciones-y-abandono)
10. [Libro de horas, saldos y cumplimiento](#10-libro-de-horas-saldos-y-cumplimiento)
11. [Historia y auditoría](#11-historia-y-auditoría)
12. [Dashboard, reportes y Excel](#12-dashboard-reportes-y-excel)
13. [Modelo conceptual y diccionario recuperable](#13-modelo-conceptual-y-diccionario-recuperable)
14. [Casos de ejemplo y aceptación funcional](#14-casos-de-ejemplo-y-aceptación-funcional)
15. [Tecnología, arquitectura y operación](#15-tecnología-arquitectura-y-operación)
16. [Plan de ejecución y flujo de trabajo](#16-plan-de-ejecución-y-flujo-de-trabajo)
17. [Estado recuperado del desarrollo](#17-estado-recuperado-del-desarrollo)
18. [Decisiones pendientes y materiales faltantes](#18-decisiones-pendientes-y-materiales-faltantes)
19. [Cambios y contradicciones entre fuentes](#19-cambios-y-contradicciones-entre-fuentes)
20. [Registro de fuentes y continuidad](#20-registro-de-fuentes-y-continuidad)

## 1. Cómo interpretar los acuerdos y las fuentes

Cada regla identifica su procedencia. Estos estados no son equivalentes:

| Estado | Significado | Cómo usarlo |
|---|---|---|
| **CONFIRMADO POR EL USUARIO** | Aparece en una instrucción o aclaración humana recuperada. | Tratarlo como requisito dentro de su alcance. |
| **CONFIRMADO SEGÚN EL PLAN** | El plan recibido lo incluye expresamente como regla confirmada, pero no se recuperó la respuesta original de la entrevista. | Conservarlo con esa procedencia; contrastarlo al incorporar el modelo v0.2. |
| **ESPECIFICADO EN EL PLAN** | Es una instrucción o criterio del plan, sin confirmación humana independiente recuperada para ese detalle. | Utilizarlo para preparar tareas; resolver contradicciones antes de convertirlo en esquema o política irreversible. |
| **PROPUESTA** | Alternativa planteada por la IA o meta orientativa de planificación. | No presentarla como requisito aprobado. |
| **PENDIENTE** | No existe respuesta suficiente, falta un material o hay contradicción. | Bloquea únicamente el comportamiento que depende de ella. |

**Orden de precedencia:** aclaración humana más reciente → acuerdo humano anterior que siga vigente → documentación que declara reglas confirmadas → instrucciones del plan → propuestas. Una tarea creada en GitHub no aprueba por sí misma una decisión de negocio.

Identificadores de fuentes utilizados: **F01** primer avance PDF; **F02** entrevista inicial recuperada; **F03** plan de programación del 23 de octubre; **F04** instrucciones y aclaraciones humanas del chat de implementación; **F05** aclaraciones vigentes; **F06** contexto e índice de planeación; **F07** documentación de validación e inicialización; **F08** flujo GitFlow. Las referencias completas aparecen al final.

## 2. Propósito, alcance y exclusiones

### 2.1 Problema que resuelve

**CONFIRMADO POR EL USUARIO / F02.** Centralizar la información de instituciones, proyectos, estudiantes y horas sociales para conocer:

- Qué proyectos existen, dónde se ejecutan y cuál es su estado.
- Qué institución y qué responsables están asociados.
- Qué estudiantes participan y a qué facultad y carrera pertenecen.
- Cuántas horas acumula cada estudiante en uno o varios proyectos.
- Qué estudiantes han cumplido el requisito y cuáles continúan pendientes.
- Qué trabajo se ha realizado por facultad, género y territorio.
- Qué estudiantes abandonaron una participación, conservando ese antecedente.

La plataforma es administrativa. Registrar una persona o una institución no le otorga acceso al portal.

### 2.2 Módulos contemplados

| Módulo | Resultado esperado | Procedencia |
|---|---|---|
| Administradores y sesiones | Inicio y cierre de sesión, validación JWT y cuentas ADMIN. | F04; seguridad ampliada en F03. |
| Catálogos | Facultades, carreras, género, clasificación, responsables y selección territorial. | F02/F03. |
| Instituciones | Registro, búsqueda, edición y consulta de sus proyectos y contactos. | F02/F03. |
| Proyectos | Alta, edición, filtros, ubicación, clasificación, estado y cierre. | F02/F03. |
| Estudiantes | Expediente único, búsqueda por carnet y consulta académica. | F02/F03. |
| Participaciones | Vincular estudiantes y proyectos, conservar estados e historial. | F02/F03. |
| Horas | Movimientos, acumulados internos/externos, saldos previos y cumplimiento. | F02/F03. |
| Dashboard y reportes | Indicadores con filtros y distinción entre personas, participaciones y horas. | F02/F03. |
| Excel | Exportación XLSX consistente con los reportes. | F03. |
| Operación y entrega | Instalación, seguridad, respaldos, recuperación, manuales y capacitación. | F01/F03. |

### 2.3 Exclusiones y alcance condicionado

**CONFIRMADO SEGÚN EL PLAN / F03, compatible con F04:** sin cuentas estudiantiles, acceso de coordinadores, registro público ni catálogo público de proyectos.

**ESPECIFICADO EN EL PLAN / F03:** Google Identity y Gmail quedan fuera de la entrega inicial. El backend se organiza como una aplicación modular; no se requieren microservicios, Kubernetes, Redis ni colas sin una necesidad medida.

**PENDIENTE / F01:** carga de documentos, formularios y evidencias. El PDF los considera únicamente si se confirman. No asumir que existe un módulo documental aprobado por el hecho de que aparezca en la propuesta inicial.

**FUERA DEL ALCANCE INICIAL SEGÚN F01:** firma electrónica, trabajo sin conexión, integraciones universitarias, notificaciones y nuevos perfiles. La importación de estudiantes tampoco se considera una integración automática aprobada.

## 3. Personas, acceso y sesiones

### 3.1 Regla vigente de administradores

**CONFIRMADO POR EL USUARIO / F04.** Existe **un solo rol ADMIN**, pueden existir **múltiples cuentas administradoras**, todas tienen **acceso completo** y deben validarse mediante **JWT**.

Esta aclaración sustituye las referencias anteriores a una cuenta única. No implementar una restricción que rechace al segundo administrador ni crear jerarquías de administradores que no se hayan solicitado.

| Persona | ¿Es un dato del negocio? | ¿Tiene acceso previsto? |
|---|---|---|
| Administrador | Sí, identidad propia de quien opera el sistema. | Sí, con autenticación y rol ADMIN. |
| Estudiante | Sí, expediente y participaciones. | No. |
| Coordinador académico | Sí, responsable académico. | No por estar registrado. |
| Contacto institucional | Sí, información de contacto. | No por estar registrado. |
| Responsable de proyecto | Sí, relación por definir con mayor precisión. | No por estar registrado. |
| Administrador técnico del servidor | Opera infraestructura. | Usa accesos técnicos independientes. |

### 3.2 Sesiones y seguridad de acceso

**ESPECIFICADO EN EL PLAN / F03:** contraseñas con Argon2id; JWT de corta duración; sesión revocable en PostgreSQL; cookies HttpOnly, Secure y SameSite en su configuración apropiada; protección CSRF y validación de origen para operaciones que modifican datos.

El plan contempla login, renovación, logout, alta inicial controlada y recuperación por consola. La API debe comprobar la autenticación en las operaciones protegidas. Cerrar sesión, recuperar credenciales o desactivar una cuenta debe invalidar las sesiones afectadas según el contrato que se acuerde.

**PENDIENTE:** identificador exacto para iniciar sesión, expiraciones, política de contraseñas, alta de administradores adicionales, cambios de credenciales, recuperación, desactivación y comportamiento ante la desactivación del último ADMIN. La aclaración de múltiples cuentas no resuelve estas políticas.

La cuenta del portal no entrega permisos sobre Ubuntu, Portainer, Cockpit o Docker. Los accesos técnicos tienen sus propias credenciales.

## 4. Glosario del negocio

| Término | Definición usada en esta consolidación |
|---|---|
| Institución / organización | Entidad vinculada a proyectos. F03 utiliza ambos términos para el mismo módulo; falta ratificar el nombre visible. |
| Proyecto | Iniciativa de horas sociales asociada a una institución y un lugar de ejecución. |
| Estudiante | Persona con expediente académico y carnet. Su expediente es independiente de la cantidad de proyectos. |
| Participación / asignación | Vínculo de un estudiante con un proyecto; permite registrar su situación individual. |
| Naturaleza | Clasificación INTERNA o EXTERNA que determina la categoría de las horas nuevas. |
| Tipo de proyecto | Clasificación solicitada en la entrevista; su lista y cardinalidad final requieren fuente. |
| Área de proyecto | Clasificación distinta del tipo; su relación con el tipo sigue pendiente. |
| Modalidad | Término mencionado sin definición recuperada suficiente. |
| Movimiento de horas | Registro individual que aporta horas al acumulado y conserva su procedencia. |
| Saldo inicial / previo | Horas anteriores al uso del sistema cuya historia puede estar incompleta. |
| fecha_horas | Fecha informada por el administrador para la realización de las horas. |
| registrado_en | Instante técnico en que se capturó el registro. |
| Cumplimiento | Alcanzar por separado la meta interna y la meta externa. |
| Abandono individual | Salida de un estudiante de un proyecto, conservando su participación e historia. |
| Cierre de proyecto | Paso de ACTIVO a un estado terminal que impide nuevas asignaciones y acreditaciones ordinarias. |
| Estudiante único | Persona contada una vez dentro del conjunto de un reporte. |
| Participaciones | Vínculos estudiante–proyecto; una persona puede aportar varios. |
| Idempotencia | Un reintento de la misma operación no genera una segunda acreditación. |

## 5. Catálogos, estructura académica y territorio

### 5.1 Estructura académica

**CONFIRMADO POR EL USUARIO / F02:** se requiere información de carrera, filtros por facultad, género del estudiante y registro de coordinadores o responsables académicos.

**ESPECIFICADO EN EL PLAN / F03:** una carrera identifica su facultad; la facultad actual del estudiante se obtiene de la carrera. Contemplar mantenimiento de facultades, carreras, género, coordinadores, tipos y áreas.

El plan propone crear, editar y desactivar opciones, impedir duplicados que solo cambien mayúsculas y conservar referencias históricas. Una opción inactiva debe seguir visible en registros anteriores y dejar de ofrecerse en altas nuevas.

**PENDIENTE:** obligatoriedad y valores de género; cardinalidad exacta carrera–coordinador; si una persona coordina varias carreras; coordinador por facultad frente a coordinador por carrera; vigencia temporal de las asociaciones.

### 5.2 Tipos, áreas, modalidad y habilidades blandas

**CONFIRMADO POR EL USUARIO / F02:** se necesitan tipos de proyecto administrables y un campo de áreas. Se mencionaron modalidad y desarrollo de habilidades blandas, sin definición suficiente.

**PENDIENTE / F03:** confirmar un tipo y varias áreas por proyecto y si las áreas son independientes del tipo. El plan menciona **12 áreas y 26 tipos**, pero los listados y sus etiquetas no se recuperaron. Esas cantidades no permiten inventar sus valores.

No equiparar tipo, área, modalidad y naturaleza. No crear un campo de habilidades blandas sin aclarar si representa categoría, actividad o resultado.

### 5.3 Ubicación territorial

**CONFIRMADO POR EL USUARIO / F02:** se necesita conocer la ubicación de ejecución y reportar el trabajo por departamento y localidad. Las notas usan «14 departamentos» y «262 municipios», además del ejemplo «Chalatenango Norte».

**PROPUESTA RECUPERADA / F02; ESPECIFICADO EN EL PLAN / F03:** utilizar selección departamento → municipio → distrito y una dirección o comunidad específica. En el chat se propuso reinterpretar las 262 localidades como distritos y distinguirlas de los municipios.

La jerarquía se documenta como propuesta/diseño recuperado, no como verificación administrativa actual realizada en esta consolidación. Antes de cargar el catálogo deben verificarse nombres, códigos y relaciones con una fuente oficial. No producir el catálogo a partir de cantidades recordadas.

Separar la ubicación de la sede institucional de la ubicación física del proyecto: el trabajo territorial se atribuye al lugar de ejecución.

## 6. Instituciones y contactos

**CONFIRMADO POR EL USUARIO / F02:** registrar instituciones y asociarles varios proyectos; guardar encargado, correo, teléfonos, dirección y ubicación.

**CONFIRMADO SEGÚN EL PLAN / F03:** cada proyecto tiene una sola institución; una institución puede tener varios proyectos. Cada proyecto tiene una sola ubicación física de ejecución.

**ESPECIFICADO EN EL PLAN / F03:** listado, búsqueda, paginación, alta, edición, detalle y desactivación; teléfonos y extensiones como texto; validar correo en el backend; conservar las relaciones al desactivar o intentar eliminar una institución usada.

| Información | Tratamiento recuperado | Pendiente |
|---|---|---|
| Nombre institucional | Identificar y buscar la institución. | Unicidad, longitud y obligatoriedad formal. |
| Dirección y ubicación de sede | Independientes de la ejecución de proyectos. | Desglose y obligatoriedad de campos. |
| Encargado/contacto | Guardar identidad y medios de contacto. | Uno o varios; contactos reutilizables; cargo. |
| Correo | Validación de formato prevista. | Si es obligatorio y si admite varios. |
| Teléfonos/extensiones | Texto, conservando formato y ceros. | Cantidad y obligatoriedad. |
| Responsable del proyecto | Distinto del contacto institucional. | Asociación exacta y número de responsables. |

El prompt del plan usa «un responsable principal, salvo que se definan varios», mientras su apartado de decisiones reconoce que la multiplicidad no está cerrada. Por eso no se toma una relación de contacto único como acuerdo definitivo.

## 7. Proyectos y ciclo de vida

### 7.1 Datos funcionales

**CONFIRMADO POR EL USUARIO / F02:** nombre, institución, ubicación, clasificación, estado y fecha de finalización al cerrar. Debe existir edición de información.

**ESPECIFICADO EN EL PLAN / F03:** descripción, contacto específico, cupo orientativo y naturaleza. Alta sin fecha final, estado inicial ACTIVO y naturaleza EXTERNA por defecto. No existe una acción «Publicar» dentro del flujo administrativo definido.

La fecha «15 de diciembre» aparece aislada en la entrevista inicial. **PENDIENTE:** determinar a qué proyecto o evento correspondía. No confundirla con la fecha de entrega del software.

### 7.2 Estados y transiciones

**CONFIRMADO POR EL USUARIO / F02:** los estados solicitados son Abandonado, Activo, Cancelado y Finalizado, presentados en orden alfabético.

**ESPECIFICADO EN EL PLAN / F03:** solo se permite salir de ACTIVO hacia ABANDONADO, CANCELADO o FINALIZADO. No se permite reapertura.

| Acción | Condición previa | Efecto | Fecha e historia | Actor / procedencia |
|---|---|---|---|---|
| Crear | Datos y relaciones válidos. | ACTIVO; EXTERNA por defecto. | Registrar creación según contrato. | ADMIN; F03. |
| Editar datos | Proyecto existente. | Actualiza campos autorizados. | Auditoría de cambios relevantes. | ADMIN; F02/F03. Edición tras cierre pendiente. |
| Cambiar naturaleza | Cambio autorizado. | Solo afecta horas nuevas. | Las horas anteriores conservan categoría. | ADMIN; regla confirmada según F03. |
| Finalizar | Proyecto ACTIVO. | FINALIZADO; bloquea nuevas asignaciones y horas ordinarias. | Fecha final válida e historia con responsable. | ADMIN; F02/F03. |
| Cancelar | Proyecto ACTIVO. | CANCELADO; bloquea nuevas asignaciones y horas ordinarias. | Evento de cierre; campos exactos de motivo/fecha pendientes. | ADMIN; F03. |
| Abandonar por abandono total | Al menos una participación y todas ABANDONADAS. | Proyecto ABANDONADO en la misma operación. | Conservar historia del proyecto y de las participaciones. | Operación administrativa que produce el último abandono; F03. |
| Abandonar manualmente el proyecto | No existe procedimiento completo recuperado. | PENDIENTE. | PENDIENTE. | No inferir desde la existencia del estado. |
| Reabrir | Cualquier estado terminal. | Rechazar según el plan. | Conservar cierre anterior. | F03. |

**PENDIENTE:** qué estados adquieren las participaciones al finalizar o cancelar; tratamiento de combinaciones de estudiantes finalizados y abandonados; fecha final en abandono/cancelación; correcciones posteriores al cierre; cambios permitidos sobre datos descriptivos de proyectos cerrados.

### 7.3 Consistencia del cierre

**ESPECIFICADO EN EL PLAN / F03:** estado e historial se guardan en una misma transacción. El cierre, la asignación y la acreditación coordinan el acceso al proyecto para que no se guarden horas o participaciones basadas en un estado anterior.

Las nuevas acreditaciones se rechazan después del cierre aunque se intente informar una fecha de realización anterior. El procedimiento excepcional para corregir datos de un proyecto cerrado sigue pendiente; no debe convertirse en una forma de ingresar horas tardías.

## 8. Estudiantes y expediente académico

**CONFIRMADO POR EL USUARIO / F02:** expediente por estudiante, búsqueda por carnet, carrera, facultad, género y consulta de horas acumuladas en distintos proyectos.

**ESPECIFICADO EN EL PLAN / F03:** alta y edición manual con carnet único como texto, nombres, apellidos, carrera y género conforme al modelo aprobado. Búsqueda por carnet y nombre; filtros por facultad/carrera; resultados paginados.

Reglas previstas:

- Conservar ceros iniciales del carnet; no convertirlo a número.
- Rechazar un carnet duplicado sin sobrescribir el expediente existente.
- Obtener la facultad actual de la carrera seleccionada.
- Mostrar participaciones y detalle por proyecto, horas internas, externas y pendientes.
- Calcular el acumulado desde el libro de horas y saldos; no guardar un total que el administrador pueda sobrescribir directamente.
- Mostrar estados vacíos cuando no existan participaciones o movimientos.

**PENDIENTE:** campos de contacto, obligatoriedad completa, normalización del carnet, identificadores alternativos, cambios de carrera y conservación de su historia.

La entrevista indica que se compartirá información académica, pero no define formato, periodicidad ni mecanismo de importación. No crear sincronización institucional por deducción.

## 9. Participaciones y abandono

### 9.1 Asignación

**CONFIRMADO POR EL USUARIO / F02:** un proyecto puede tener muchos estudiantes y un estudiante puede participar en varios proyectos.

**ESPECIFICADO EN EL PLAN / F03:** una sola participación por pareja estudiante–proyecto; altas únicamente en proyectos ACTIVOS; no reincorporación después del abandono. La participación inicial se maneja como ACTIVA en el flujo de horas.

**CONFIRMADO SEGÚN EL PLAN / F03:** el cupo solo advierte. El campo `alumnos_solicitados` no impone un límite duro.

Ejemplo: si se solicitan cinco estudiantes, asignar una sexta persona muestra una advertencia y permite continuar. La facultad del estudiante no debe usarse como restricción de asignación solo porque exista un filtro por facultad.

### 9.2 Abandono individual y total

**CONFIRMADO POR EL USUARIO / F02:** debe quedar registro de quienes se inscribieron y no realizaron el trabajo.

**CONFIRMADO SEGÚN EL PLAN / F03:** abandonar conserva horas e historial, no permite reincorporación y puede producir abandono total del proyecto.

| Situación | Resultado previsto |
|---|---|
| Un estudiante abandona y quedan otras participaciones activas. | Se conserva su participación ABANDONADA; el proyecto continúa ACTIVO. |
| Todas las participaciones existentes quedan ABANDONADAS y hay al menos una. | El proyecto pasa a ABANDONADO. |
| El proyecto no tiene estudiantes. | No aplicar abandono total por evaluar un conjunto vacío. |
| No quedan activos, pero existen participaciones finalizadas. | No equivale a «todos abandonaron»; cierre pendiente de definición. |
| Se intenta duplicar una participación. | Rechazar sin modificar el historial previo. |
| Se intenta reinscribir a quien abandonó ese proyecto. | Rechazar según el plan. |

**PENDIENTE:** catálogo completo de estados individuales, finalización de una participación, cancelación individual, motivos/fechas obligatorios y efecto del cierre global. Solo ACTIVA y ABANDONADA cuentan con comportamiento suficiente recuperado; no inventar el resto de la máquina de estados.

## 10. Libro de horas, saldos y cumplimiento

### 10.1 Metas y categorías

**CONFIRMADO SEGÚN EL PLAN / F03:** cada estudiante debe acumular **300 horas internas y 300 externas**. Son metas independientes. Las horas se expresan como enteros y se agregan mediante movimientos. Se admiten saldos iniciales sujetos a definir su tratamiento.

No confundir:

- **300 horas reloj:** presupuesto total de desarrollo y entrega del software.
- **600 horas sociales:** meta estudiantil compuesta por 300 internas y 300 externas.

La pregunta inicial sobre el significado de «600» queda orientada por la regla posterior del plan; la nota «204 proyectos / 825 estudiantes» es un ejemplo de entrevista, no un inventario validado.

Para una vista de avance actual, con `I` y `E` como acumulados válidos:

```text
total_registrado = I + E
pendientes_internas = max(0, 300 - I)
pendientes_externas = max(0, 300 - E)
pendientes_totales = pendientes_internas + pendientes_externas
cumplio = (I >= 300) Y (E >= 300)
```

Estas fórmulas expresan la regla del plan. No suponen que se haya aprobado un máximo de registro por categoría. Si se muestra un porcentaje o una barra limitada al 100 %, su presentación debe acordarse; no alterar los movimientos para ajustarla.

### 10.2 Registro de movimientos

**ESPECIFICADO EN EL PLAN / F03:**

| Regla | Comportamiento previsto |
|---|---|
| Cantidad | Entero positivo en una acreditación ordinaria; rechazar decimales antes de convertir. |
| Proyecto | Debe estar ACTIVO al guardar. |
| Participación | Debe estar ACTIVA al guardar. |
| Fecha de realización | `fecha_horas`, informada por el administrador. |
| Fecha de captura | `registrado_en`, conservada por el sistema. |
| Categoría | Copiar la naturaleza vigente del proyecto en el movimiento. |
| Acumulado | Sumar movimientos válidos y saldos reconocidos por categoría. |
| Historia | Conservar participación, categoría y datos históricos que finalmente se aprueben. |
| Cierre | Rechazar acreditaciones ordinarias nuevas tras cerrar proyecto o participación. |
| Reintentos | Identificar la operación para no acreditar dos veces. |
| Cambios | No exponer edición, borrado ni reversos mientras su política siga pendiente. |

**CONFIRMADO SEGÚN EL PLAN / F03:** cambiar un proyecto de EXTERNA a INTERNA no reclasifica movimientos anteriores. Por ejemplo, 40 horas externas seguidas de 10 internas se mantienen como 40 externas y 10 internas.

**PENDIENTE:** si se permiten fechas futuras, fechas anteriores al inicio o a la asignación, rango máximo por movimiento, detalle de actividad/evidencia y obligatoriedad del motivo. La existencia de `fecha_horas` no define esas validaciones.

### 10.3 Idempotencia

**ESPECIFICADO EN EL PLAN / F03:** `solicitud_id` UUID único identifica una solicitud de acreditación. El cliente reutiliza la misma clave al reintentar.

- Misma clave y mismo contenido: un solo movimiento, sin sumar nuevamente.
- Misma clave y contenido diferente: conflicto, sin alterar la acreditación anterior.
- Un fallo de red no debe obligar al usuario a adivinar si se duplicaron horas.

**PENDIENTE TÉCNICO:** ámbito exacto de unicidad, comparación de contenido y respuesta HTTP. Es una ampliación propuesta en el plan; no evidencia de que ese campo exista en el modelo v0.2.

### 10.4 Saldos iniciales y fecha de corte

**CONFIRMADO SEGÚN EL PLAN / F03:** pueden existir horas previas. Deben mantenerse separadas por categoría y no sobrescribir el libro de movimientos.

**ESPECIFICADO EN EL PLAN / F03:** un saldo sin historia puede aumentar el acumulado actual; no se atribuye automáticamente a un año, institución, territorio o proyecto desconocidos.

**PENDIENTE:** fecha de corte, documento de respaldo, captura/aprobación, asociación opcional con proyectos, corrección del saldo y prevención de doble conteo al cargar después movimientos del mismo período.

No crear proyectos ficticios ni fechas inventadas para completar antecedentes desconocidos.

### 10.5 Correcciones, reversos y operaciones posteriores al cierre

**PROPUESTA / F03:** corregir mediante reversos o movimientos compensatorios para conservar el original.

**PENDIENTE:** aprobación del mecanismo, actor que valida, motivo, enlace al movimiento corregido, fecha contable, límites, efecto en reportes y procedimiento administrativo posterior al cierre.

Hasta resolverlo no habilitar edición, borrado ni reversos. Tampoco asumir que una corrección posterior al cierre equivale a una acreditación ordinaria permitida.

## 11. Historia y auditoría

**CONFIRMADO POR EL USUARIO / F02:** conservar el antecedente de abandono y consultar el acumulado de una persona en varios proyectos.

**CONFIRMADO SEGÚN EL PLAN / F03:** conservar horas tras abandono y mantener la categoría original de los movimientos aunque cambie la naturaleza del proyecto.

**ESPECIFICADO EN EL PLAN / F03:** registrar eventos relevantes con fecha y actor; guardar conjuntamente el estado del proyecto y su historia; conservar referencias de catálogos usados.

| Cambio | Tratamiento recuperado | Estado |
|---|---|---|
| Naturaleza del proyecto | No modifica la categoría de horas anteriores. | Confirmado según F03. |
| Abandono | Conserva participación, horas e historia. | Confirmado según F03. |
| Cambio de carrera del estudiante | Facultad actual deriva de nueva carrera; efecto histórico por definir. | Parcialmente especificado / pendiente. |
| Cambio de facultad de una carrera | Efecto sobre reportes pasados y vigencia no definido. | Pendiente. |
| Cambio de tipo o áreas | No hay política completa de clasificación histórica recuperada. | Pendiente. |
| Cambio del contacto institucional | No sustituye automáticamente al responsable particular del proyecto. | Especificado en F03. |
| Desactivación de un catálogo | Sigue visible en registros antiguos; no se ofrece en altas nuevas. | Especificado en F03. |
| Corrección de horas | Mecanismo e impacto histórico por aprobar. | Pendiente. |

No afirmar que el sistema ya conserva una instantánea académica o de clasificación sin el acuerdo y las columnas que la soporten. Las operaciones de administradores distintos deben poder atribuirse a su cuenta cuando la regla requiera trazabilidad.

## 12. Dashboard, reportes y Excel

### 12.1 Preguntas del negocio

**CONFIRMADO POR EL USUARIO / F02:** conocer proyectos, estudiantes participantes, cumplimiento, hombres y mujeres por facultad, distribución académica, acumulados por persona/proyecto y trabajo por territorio.

**CONFIRMADO SEGÚN EL PLAN / F03:** dashboard y XLSX deben compartir filtros y cifras; estudiantes únicos se cuentan aparte de participaciones.

### 12.2 Definiciones de indicadores

| Indicador | Definición recuperada / límite |
|---|---|
| Estudiantes registrados | Expedientes existentes; no equivale a participantes del período. |
| Estudiantes asignados | Personas con participaciones; distinguirlos de quienes registraron horas. |
| Participaron en el período | Según F03, estudiantes únicos con horas realizadas dentro del período. |
| Participaciones | Vínculos estudiante–proyecto; su filtro temporal exacto debe describirse por reporte. |
| Horas del período | Movimientos filtrados por `fecha_horas`, no por `registrado_en`. |
| Acumulado actual | Horas válidas conocidas más saldos reconocidos, separadas por naturaleza. |
| Cumplieron actualmente | Estudiantes con al menos 300 internas y 300 externas. |
| Completaron en un año | Pendiente: requiere regla para la fecha de cumplimiento y correcciones. |
| Proyectos por estado | Conteo de proyectos; no contar filas multiplicadas por estudiantes o áreas. |
| Trabajo territorial | Usar ubicación de ejecución conocida. No atribuir antecedentes sin ubicación. |

Una persona con horas en dos proyectos equivale a **un estudiante único y dos participaciones**. Si los subtotales por área se superponen, su suma no debe presentarse como un total de personas distintas.

### 12.3 Filtros y coherencia

**ESPECIFICADO EN EL PLAN / F03:** período, facultad, carrera, género, institución, ubicación, tipo y área; filtros de estado para listados de proyectos. Definir expresamente qué fecha y qué versión histórica utiliza cada consulta.

No multiplicar movimientos al relacionar varias áreas de un proyecto. Un saldo inicial sin historia no alimenta actividad anual o territorial inventada. No presentar «completados en el año» tomando solamente a quienes cumplen hoy.

**PENDIENTE:** límites inclusivos de fechas, períodos académicos frente a intervalos libres, catálogo de género y tratamiento de desconocidos; criterio histórico o actual para carrera, facultad y clasificación; fecha de cumplimiento; layout final de las consultas.

### 12.4 Exportación Excel

**ESPECIFICADO EN EL PLAN / F03:** exportación XLSX autenticada desde el backend, mediante Excelize; mismas consultas y filtros que el dashboard; hojas de resumen, detalle y filtros, con fecha de generación.

- Carnet como texto, horas como números.
- Texto introducido por usuarios no se transforma en fórmulas.
- Límite inicial documentado de 50 000 filas; si se excede, solicitar reducir filtros.
- Consultas y generación cancelables si se desconecta el cliente.
- Evitar descargar toda la base al navegador para construir la exportación.

El límite es una especificación técnica del plan y debe validarse con el equipo. La exportación de reportes a PDF se preguntó en la entrevista inicial, pero no se recuperó su aprobación; los manuales PDF son un producto distinto.

## 13. Modelo conceptual y diccionario recuperable

### 13.1 Límite de recuperación del modelo v0.2

F03 cita el **Modelo Preliminar Horas Sociales v0.2 del 23 de octubre de 2026** y menciona **19 tablas de negocio**. No se recuperaron ese archivo, su diagrama ni sus tablas completas. Por tanto:

- No se enumeran 19 tablas inventadas.
- No se presenta un diccionario SQL como si estuviera aprobado.
- No se afirma conocer longitudes, nulabilidad, PK/FK, nombres físicos o índices ausentes.
- No se crea un diagrama entidad–relación que pueda confundirse con el modelo original.

La siguiente relación conceptual resume reglas recuperadas, no nombres físicos de tablas:

| Relación | Cardinalidad recuperada | Procedencia / limitación |
|---|---|---|
| Institución → proyecto | Una institución, varios proyectos; una institución por proyecto. | F02/F03. |
| Estudiante ↔ proyecto | Muchos a muchos mediante participación. | F02; unicidad del vínculo especificada en F03. |
| Proyecto → ubicación de ejecución | Una ubicación física por proyecto. | Confirmado según F03. |
| Carrera → facultad | Una facultad por carrera en el diseño descrito. | Especificado en F03; temporalidad pendiente. |
| Participación → movimiento | Varios movimientos asociados al vínculo. | Especificado en F03. |
| Proyecto → tipo / área | Un tipo y varias áreas es una opción pendiente de ratificación. | No fijar cardinalidad física. |
| Carrera ↔ coordinador | No consolidada. | El plan contiene instrucciones y pendientes contradictorios. |
| Institución / proyecto → contacto | Multiplicidad y reutilización no consolidadas. | Pendiente. |
| Administrador → sesión | Sesiones revocables previstas. | Diseño técnico del plan; esquema pendiente. |

### 13.2 Datos y restricciones recuperados

Los nombres entre código solo se conservan cuando aparecen en las fuentes. «No recuperado» significa que no existe base para fijar el dato físico en esta consolidación.

| Concepto / dato | Nombre recuperado | Tipo recuperado o tratamiento previsto | Obligatoriedad / restricción recuperada | Estado / fuente |
|---|---|---|---|---|
| Identificadores | No se recuperaron nombres físicos. | UUID propuesto en etapa 03. | PK concretas no recuperadas. | Especificado F03. |
| Administrador | No recuperado. | Identidad individual; rol ADMIN. | Varias cuentas permitidas; acceso JWT. | Confirmado usuario F04. |
| Sesión administradora | `sesion_admin` citado. | Tabla técnica adicional; campos no recuperados. | Revocación prevista. | Diseño F03. |
| Facultad | No recuperado. | Catálogo. | Duplicados/inactivación según contrato. | F02/F03. |
| Carrera | No recuperado. | Catálogo relacionado con facultad. | FK física y nombres no recuperados. | F02/F03. |
| Coordinador | No recuperado. | Persona académica. | Cardinalidades pendientes. | F02/F03. |
| Género | No recuperado. | Catálogo previsto. | Valores y obligatoriedad pendientes. | F02/F03. |
| Institución | No recuperado. | Nombre, sede y contacto. | Unicidad/obligatoriedad completas pendientes. | F02/F03. |
| Teléfono / extensión | No recuperado. | Texto. | No convertir a número. | Especificado F03. |
| Correo | No recuperado. | Texto con validación de formato. | Si es obligatorio, pendiente. | F02/F03. |
| Proyecto | No recuperado. | Registro administrativo. | Una institución y una ejecución física. | Confirmado según F03. |
| Nombre de proyecto | No recuperado. | Texto; longitud no recuperada. | Obligatoriedad formal y duplicados pendientes. | F02. |
| Descripción | No recuperado. | Texto; longitud no recuperada. | Pendiente. | Especificado F03. |
| Estado de proyecto | No recuperado. | ACTIVO, ABANDONADO, CANCELADO, FINALIZADO. | Transiciones de sección 7. | F02/F03. |
| Naturaleza | No recuperado. | INTERNA / EXTERNA. | EXTERNA inicial; categoría histórica por movimiento. | Confirmado según F03. |
| Cupo orientativo | `alumnos_solicitados`. | Numérico; tipo físico no recuperado. | Advierte, no bloquea. Rango pendiente. | F03. |
| Fecha final | Nombre físico no recuperado. | Fecha; tipo SQL exacto no recuperado. | No exigida al alta; se captura al finalizar. | F02/F03. |
| Tipo y áreas | Nombres físicos no recuperados. | Catálogos separados. | Cardinalidad y etiquetas pendientes. | F02/F03. |
| Carnet | Nombre físico no recuperado. | Texto. | Único; conservar ceros; normalización pendiente. | Especificado F03. |
| Nombres / apellidos | No recuperados. | Texto. | Longitudes/obligatoriedad completas pendientes. | Especificado F03. |
| Carrera actual del estudiante | No recuperado. | Relación académica. | Facultad derivada; historia pendiente. | F03. |
| Participación | Nombre físico no recuperado. | Vínculo estudiante–proyecto. | Pareja única; no reincorporar tras abandono. | Especificado F03. |
| Estado individual | No recuperado. | ACTIVA / ABANDONADA en reglas recuperadas. | Resto del catálogo y cierres pendientes. | F03. |
| Horas de acreditación | Nombre físico no recuperado. | INTEGER. | Entero positivo para movimiento ordinario. | F03. |
| Fecha de horas | `fecha_horas`. | Fecha de realización; tipo SQL exacto no recuperado. | Informada por ADMIN; límites pendientes. | Confirmado según F03. |
| Captura técnica | `registrado_en`. | Instante; tipo SQL/zona exactos no recuperados. | No sustituye fecha de realización en reportes. | F03. |
| Clave de reintento | `solicitud_id`. | UUID. | Única; misma clave con cuerpo distinto es conflicto. | Ampliación técnica especificada en F03. |
| Saldo previo | No recuperado. | Horas enteras por categoría. | Corte, historia y doble conteo pendientes. | F03. |
| Actor / evento de historia | No recuperado. | Referencia a quien ejecutó y evento. | Campos, alcance y retención pendientes. | F01/F03/F04. |

### 13.3 Unicidad, relaciones e índices

**ESPECIFICADO EN EL PLAN / F03:** carnet único, pareja estudiante–proyecto única, `solicitud_id` único y claves foráneas válidas. Evitar cascadas destructivas que borren antecedentes. Catálogos y datos de demostración se cargan por separado.

**PROPUESTA TÉCNICA:** evaluar índices para carnet, referencias académicas, institución/estado del proyecto, participación y fechas de movimientos. No se recuperó un inventario de índices aprobados ni sus nombres; deben derivarse del esquema y de consultas medidas.

## 14. Casos de ejemplo y aceptación funcional

Todos los datos de esta sección son sintéticos. Los resultados siguen las reglas con la procedencia indicada; no constituyen nuevas decisiones.

| Caso | Entrada y acción | Resultado esperado | Base |
|---|---|---|---|
| C01 Múltiples ADMIN | Crear una segunda cuenta por el procedimiento que se acuerde. | No rechazarla por existir ya otra; ambas usan rol ADMIN y JWT. | Usuario F04. |
| C02 Sin sesión | Consultar expedientes, operaciones o XLSX sin sesión válida. | Acceso denegado. | F03/F04. |
| C03 Carnet | Registrar `00001234` y consultar/exportar. | Se conserva exactamente como texto. Duplicarlo se rechaza. | F03. |
| C04 Varios proyectos | Una persona participa en A y B. | Un expediente y dos participaciones. | F02/F03. |
| C05 Cupo | Proyecto solicita cinco; se asigna una sexta persona. | Advertencia sin bloqueo por cupo. | F03. |
| C06 Abandono parcial | Dos participaciones activas; una abandona. | Esa participación queda abandonada; se conservan sus horas; proyecto activo. | F03. |
| C07 Abandono total | La única participación restante abandona y todas son abandonadas. | Proyecto abandonado en la misma operación; no nuevas asignaciones. | F03. |
| C08 Conjunto vacío | Proyecto sin participaciones. | No marcar abandono total por ausencia de activos. | F03. |
| C09 Estados mezclados | Una participación finalizada y otra abandonada. | No inferir abandono total; regla general de cierre pendiente. | F03. |
| C10 Reincorporación | Reasignar a una persona que abandonó ese proyecto. | Rechazo; mantener historia original. | F03. |
| C11 Suma | Registrar 20 y luego 15 horas. | Dos movimientos, 35 horas. | F03. |
| C12 Meta | 180 internas y 225 externas. | 405 registradas; faltan 120 internas y 75 externas: 195 en total. | F03. |
| C13 Sin compensación | 350 internas y 250 externas. | 600 registradas; no cumplió; faltan 50 externas. | F03. |
| C14 Categoría histórica | 40 externas; cambiar proyecto a interno; agregar 10. | 40 externas y 10 internas. | F03. |
| C15 Decimales | Intentar registrar 1.5 horas. | Rechazar, sin redondear ni truncar. | F03. |
| C16 Reintento | Enviar dos veces la misma clave y contenido. | Una acreditación. | F03. |
| C17 Clave conflictiva | Misma clave; cambiar 10 horas por 20. | Conflicto, sin sobrescribir ni duplicar. | F03. |
| C18 Cierre | Finalizar y luego intentar acreditar horas o asignar estudiante. | Rechazo. | F03. |
| C19 Carrera | Cambiar carrera actual. | Facultad actual derivada; impacto histórico requiere decisión previa. | F03. |
| C20 Saldo desconocido | Cargar saldo externo reconocido sin año ni territorio. | Aumenta acumulado; no inventa actividad anual o territorial. Carga depende de política de corte. | F03. |
| C21 Reporte único | Una persona con horas en dos proyectos del período. | Un estudiante único, dos participaciones; horas sin duplicación. | F03. |
| C22 Varias áreas | Un movimiento de 10 horas en un proyecto con dos áreas. | Total general de 10, no 20; cardinalidad de áreas debe aprobarse primero. | F03. |
| C23 Fechas distintas | Horas realizadas en septiembre y capturadas en octubre. | Actividad de septiembre según `fecha_horas`. | F03. |
| C24 Excel | Exportar con los filtros activos. | Mismas cifras del dashboard; carnet texto y horas numéricas. | F03. |
| C25 Texto parecido a fórmula | Nombre comienza con `=`. | Excel muestra texto y no ejecuta fórmula. | F03. |
| C26 Corrección cerrada | Se detecta error en un proyecto finalizado. | No habilitar edición/reverso improvisado; aplicar política cuando se apruebe. | Pendiente F03. |

## 15. Tecnología, arquitectura y operación

### 15.1 Elección vigente del desarrollo

**CONFIRMADO POR EL USUARIO / instrucciones de inicialización recuperadas en F04 y chat tecnológico:** Go, net/http, chi, pgx/v5 y pgxpool; React, TypeScript y Vite; Tailwind con integración oficial y sin CDN; PostgreSQL 18; Docker/Compose; migraciones Goose; Git.

El PDF inicial proponía Node.js/Fastify/Drizzle para el backend. Las instrucciones posteriores de Go sustituyen esa selección. Node puede seguir formando parte de las herramientas de construcción del frontend; eso no convierte a Node en el backend de negocio.

**ESPECIFICADO EN EL PLAN / F03:** SQL parametrizado, sqlc, OpenAPI, React Router, TanStack Query, Excelize, Go testing y Playwright. Estas herramientas adicionales no se consideran instaladas solo por figurar en el plan.

### 15.2 Flujo de la aplicación

```text
Navegador
  → web (React servido por Nginx en producción)
    → /api → backend Go
      → PostgreSQL
```

En desarrollo, Vite hace de proxy para `/api`. El navegador no accede directamente a PostgreSQL. En la publicación prevista, Cloudflare Tunnel conduce el dominio hacia web; el plan posterior ubica `cloudflared` en un contenedor.

**CONFIRMADO POR EL USUARIO / inicialización F04:** dependencias, compilación y ejecución dentro de contenedores; no exigir instalaciones locales de Go, Node/npm o PostgreSQL. El equipo de desarrollo necesita las herramientas anfitrionas indicadas en el proyecto, como Docker, Git y editor.

### 15.3 Estructura acordada

```text
apps/api/cmd/api/                 inicio de la API
apps/api/internal/config/        configuración
apps/api/internal/platform/      base de datos, HTTP y servicios comunes
apps/api/internal/modules/       módulos de negocio
apps/api/db/migrations/          migraciones numeradas
apps/api/db/queries/             consultas
apps/web/src/app/                estructura y navegación
apps/web/src/components/         componentes comunes
apps/web/src/features/           módulos de interfaz
apps/web/src/lib/                HTTP y utilidades
deploy/docker/                  imágenes
deploy/nginx/                   servidor web
scripts/                        operación y comprobaciones
docs/                           documentación
api/                            contrato HTTP
tests/                          integración
```

El plan amplía la estructura con un comando de administración y código generado por sqlc cuando se necesiten. Evitar implementaciones ficticias y capas vacías.

### 15.4 Persistencia y seguridad operativa

**ESPECIFICADO EN LAS INSTRUCCIONES Y EL PLAN / F03/F04:**

- Servicios `web`, `api`, `db` y `migrator` temporal.
- Publicación local de web en `127.0.0.1:8080`; API y base sin puertos públicos directos.
- API conectada a `db:5432`, no a localhost dentro de su contenedor.
- Volumen persistente de PostgreSQL 18 en `/var/lib/postgresql`.
- Entornos separados: desarrollo, pruebas, ensayo y producción, con datos y credenciales independientes.
- Versiones fijadas, lockfiles y migraciones numeradas; no imágenes `latest`.
- Secretos fuera de Git, imágenes y logs; no imprimir contraseñas.
- No borrar volúmenes ni reiniciar la base para corregir errores.
- Separar privilegios de migración y aplicación antes del negocio.
- Runtime de web/API sin root; salud real y cierre ordenado de procesos.

Rutas iniciales: `/api/health/live` indica funcionamiento del proceso; `/api/health/ready` verifica PostgreSQL y responde indisponibilidad cuando falla. El plan contempla `/api/v1` para negocio; no confundir ese prefijo con las rutas de salud existentes.

### 15.5 Producción y recuperación

**CONTEMPLADO EN F01/F03:** laptop institucional con Ubuntu Server, Docker/Compose, dominio, HTTPS y Cloudflare Tunnel; Portainer para Docker y Cockpit en el sistema anfitrión. Paneles técnicos con acceso restringido y credenciales independientes.

**PROPUESTAS QUE REQUIEREN MEDICIÓN Y ACUERDO:** copias diarias cifradas fuera de la laptop, pérdida máxima objetivo de 24 h y recuperación en 4 h según el plan posterior. El PDF hablaba de una jornada y proponía retener 14 copias diarias y 8 semanales. Falta acordar política final, destino, responsables y presupuesto.

Respaldar y restaurar deben probarse en un entorno separado. Volver a una imagen anterior no revierte automáticamente las migraciones ni los datos. La laptop constituye un único punto de operación sujeto a energía, Internet y estado del equipo.

## 16. Plan de ejecución y flujo de trabajo

### 16.1 Tiempo y entrega

**PLANIFICACIÓN RECUPERADA / F01/F03:** período del 1 de septiembre al 10 de diciembre de 2026, entrega objetivo el **10 de diciembre de 2026**, esfuerzo total previsto de **300 horas reloj**, equivalente a 18 000 minutos.

Son estimaciones, no constancia de trabajo ejecutado ni acreditación automática. Reutilizar y validar el avance previo; no agregar otras 300 horas ni contabilizar dos veces actividades solapadas. La bitácora registra fecha, actividad, duración real y evidencia.

El PDF proponía hitos de módulos el 30 de octubre, consultas el 10 de noviembre, laptop el 17, dominio el 20, pruebas el 1 de diciembre, piloto/publicación condicionada el 4 y cierre el 10. Se conservan como cronograma histórico que debe conciliarse con avance y disponibilidad reales, no como hitos ya cumplidos.

### 16.2 Etapas del plan posterior

| Etapa | Entrega | Horas previstas |
|---|---|---:|
| 00 | Alcance y decisiones de negocio | 10 |
| 01 | Inicialización y primer arranque Docker | 10 |
| 02 | Redes, persistencia y herramientas Docker | 18 |
| 03 | Esquema relacional y migraciones | 20 |
| 04 | Base de API y contrato | 14 |
| 05 | Autenticación y administración | 20 |
| 06 | Estructura frontend con Tailwind | 18 |
| 07 | Catálogos y configuración académica | 14 |
| 08 | Instituciones y contactos | 12 |
| 09 | Proyectos y ciclo de vida | 18 |
| 10 | Expediente y búsqueda de estudiantes | 16 |
| 11 | Asignaciones y abandono | 16 |
| 12 | Movimientos de horas y saldos | 26 |
| 13 | Dashboard y estadísticas | 18 |
| 14 | Excel | 10 |
| 15 | Pruebas, rendimiento y recuperación | 16 |
| 16 | Imágenes y versiones reproducibles | 12 |
| 17 | Ubuntu Server y administración | 12 |
| 18 | Ensayo de producción, dominio y carga | 14 |
| 19 | Entrega y versión estable | 6 |
| **Total** | **Proyecto completo** | **300** |

**CONFIRMADO POR EL USUARIO / F04:** dividir el desarrollo en issues pequeños, con instrucciones utilizables como prompts, criterios de aceptación y dependencias. Priorizar base de identidad/autenticación y continuar módulo por módulo.

El orden ajustado en F06 contempla persistencia mínima de identidad antes de JWT y sesiones; no exige construir todas las tablas de negocio para autenticar. Catálogos, instituciones y estudiantes pueden avanzar cuando sus dependencias concretas estén resueltas. Una duda debe bloquear solo lo que depende de su respuesta.

### 16.3 GitFlow y versiones

**CONFIRMADO POR EL USUARIO / F04; DOCUMENTADO EN F05/F08:**

- Repositorio: [ServicioHorasSociales](https://github.com/Luisnativii/ServicioHorasSociales).
- `main` conserva únicamente README por ahora.
- `develop` integra el desarrollo.
- Una rama por issue y PR hacia `develop`, con comprobaciones.
- Automatizar nomenclatura de ramas y aumento de versión.
- Base de versión `V0.0.1`; convención documentada `VMAJOR.MINOR.PATCH`.

Los ejemplos antiguos `v0.12.0` o una versión por etapa no prevalecen sobre GitFlow vigente ni autorizan crear tags o desplegar por cada tarea. Las reglas exactas de ramas, etiquetas y automatización se consultan en F08.

**DETALLE DOCUMENTADO EN F08:** la rama usa `tipo/NUMERO-descripcion`; el issue pasa de `status:backlog` a `status:ready` para iniciar su rama y luego a `status:in-progress`. La creación de un issue por sí sola no crea una rama. Los tipos contemplados son feature, fix, chore, docs, refactor, test y ci, con el prefijo de PR correspondiente.

El PR identifica su issue en el título, incluye `Closes #NUMERO` y lleva exactamente una etiqueta `version:patch`, `version:minor` o `version:major`. Requiere los checks GitFlow y Docker; se integra mediante squash. La pipeline etiqueta la integración vigente de `develop` cuando pasan sus comprobaciones, sin incrementar versión en cada commit ni sobrescribir tags. La automatización actual de tags no publica imágenes ni despliega; esas acciones pertenecen a una etapa posterior.

### 16.4 Criterios de entrega

**CONTEMPLADOS EN F01/F03:** flujos administrativos aceptados; acceso privado validado con varias cuentas ADMIN según la corrección vigente; reportes/Excel conciliados; integridad de datos; manual de usuario y operación; OpenAPI; modelo/diccionario; pruebas; copia externa restaurada; capacitación; bitácora y transferencia de accesos por canal seguro.

**PROPUESTAS DE CALIDAD / F03:** 10 000 estudiantes, 1 000 proyectos y 200 000 movimientos sintéticos; cinco clientes concurrentes; p95 menor de 500 ms en listados y 1 500 ms en dashboard; exportación de 50 000 filas en menos de 30 s. Son referencias de ensayo, no volúmenes reales ni resultados medidos.

El PDF inicial usaba una muestra menor y consultas p95 inferiores a dos segundos; estos objetivos deben reconciliarse y medirse en el equipo real. No publicar con defectos críticos/altos abiertos ni afirmar recuperación sin probarla.

## 17. Estado recuperado del desarrollo

Esta sección describe evidencia documental y de conversación al corte; esta tarea no volvió a ejecutar los contenedores ni auditó en vivo todo GitHub.

| Área | Situación recuperada |
|---|---|
| Copia `horas-sociales` dentro de este proyecto documental | Inicialización antigua; README indica validaciones Docker pendientes y ausencia de negocio/autenticación. |
| Repositorio de trabajo en `/Users/luisnativii/Desktop/horas-sociales` | El chat y F05/F07 documentan inicialización ejecutada, salud real, PostgreSQL, persistencia y comprobaciones. Es distinto de la copia antigua. |
| Versión base | F05 acredita V0.0.1 e issue #1 para no repetir la inicialización. |
| Validación manual | Referenciada en issue #2; no inferir aceptación final por pruebas automáticas. |
| Planeación | F06 y el chat registran 95 tareas nuevas y 10 hitos, además de tareas previas e índice; incluye dependencias. El conteo abierto puede cambiar. |
| Revisión de planeación | El chat recuperado registra PR #100; su estado posterior debe consultarse al continuar. |
| Autenticación y negocio | La documentación recuperada todavía los sitúa como siguientes etapas; existencia de issues no equivale a implementación. |
| Modelo completo | v0.2 pendiente de incorporar y contrastar. |

Este documento se guarda en la raíz del proyecto documental. Puede incorporarse después al repositorio como `docs/Planeacion/Logica_Negocio.md` mediante el flujo de consolidación; no se modificó aquí el esquema ni se publicaron cambios remotos.

## 18. Decisiones pendientes y materiales faltantes

### 18.1 Preguntas concretas

| ID | Pregunta pendiente | Módulo / qué bloquea |
|---|---|---|
| D01 | ¿Dónde está el modelo v0.2 completo y cuáles son sus 19 tablas, campos y restricciones aprobadas? | Diccionario, migraciones y correspondencia del esquema. |
| D02 | ¿Qué respuestas originales respaldan las reglas que F03 declara confirmadas? | Ratificación de reglas y resolución de divergencias. |
| D03 | ¿Organización e institución serán el mismo nombre visible o existe una diferencia de negocio? | Terminología final; no crear dos módulos por intuición. |
| D04 | ¿Cada proyecto tiene un tipo y varias áreas? ¿Las áreas dependen del tipo? | Clasificación, relaciones y filtros. |
| D05 | ¿Cuáles son las etiquetas exactas de las 12 áreas y 26 tipos mencionados? | Catálogos iniciales y selección en formularios. |
| D06 | ¿Qué significan modalidad y habilidades blandas? | Campos y reportes relacionados. |
| D07 | ¿Un coordinador puede tener varias carreras y una carrera varios coordinadores? ¿Existe vínculo por facultad? | Modelo académico y responsables. |
| D08 | ¿Coordinadores y contactos comparten una entidad de persona? ¿Tienen relación directa con proyectos? | Modelo y reutilización de datos. |
| D09 | ¿Cuántos contactos admite institución/proyecto y qué campos son obligatorios? | Contactos y validaciones definitivas. |
| D10 | ¿Qué estados completos tiene una participación y cómo se finaliza individualmente? | Ciclo de participación y acreditación. |
| D11 | ¿Qué pasa con cada participación al cancelar o finalizar un proyecto? | Cierre coherente y reportes. |
| D12 | ¿Cómo cerrar un proyecto con participaciones finalizadas y abandonadas? | Cierre mixto, sin inferir abandono total. |
| D13 | ¿Existe abandono manual global? ¿Qué fecha y motivo exige? | Transición a ABANDONADO fuera del caso automático. |
| D14 | ¿Qué fecha final se registra al cancelar/abandonar y cómo se valida? | Fechas de cierre. |
| D15 | ¿Qué datos pueden editarse después del cierre? | Correcciones descriptivas y auditoría. |
| D16 | ¿Se aprueban reversos? ¿Cómo se autoriza y registra una corrección tras cierre? | Correcciones de horas; conservar originales. |
| D17 | ¿Cuál es el corte de saldos iniciales y cómo impedir que se dupliquen con movimientos importados? | Carga histórica y acumulados confiables. |
| D18 | ¿Qué respaldo y metadatos mínimos exige un saldo previo? | Captura de antecedentes desconocidos. |
| D19 | ¿Reportes usan carrera/facultad del momento o actuales? ¿Cómo se conserva la historia? | Movimientos e informes académicos. |
| D20 | ¿Cambios de tipo/área deben conservar clasificación histórica? | Informes por clasificación. |
| D21 | ¿Qué datos de estudiante son obligatorios y cómo se normaliza el carnet? | Formularios, unicidad y carga inicial. |
| D22 | ¿Se permite fecha futura o anterior a asignación/inicio? ¿Hay máximo de horas por movimiento? | Validación temporal y cantidades. |
| D23 | ¿La meta 300+300 tiene excepciones por carrera o cohortes? | Confirmar alcance uniforme del cumplimiento descrito por F03. |
| D24 | ¿Cómo se determina la fecha en que alguien completó las horas, incluso con saldo previo y correcciones? | «Completaron en el año» y estado a una fecha. |
| D25 | ¿Qué períodos, límites de fecha y tratamiento de datos desconocidos necesitan los informes? | Contrato exacto de reportes. |
| D26 | ¿Se requiere exportar reportes a PDF además de Excel? | Exportación adicional, no manuales. |
| D27 | ¿Se confirma el módulo de documentos? ¿Qué tipos, tamaños y relación con registros? | Archivos y evidencias. |
| D28 | ¿Los datos académicos llegarán por archivo? ¿Formato y proceso autorizado? | Importación; no implica integración automática. |
| D29 | ¿Cuál será identificador de login, alta/recuperación y política del último ADMIN? | Contrato de administración de cuentas. |
| D30 | ¿Qué catálogo territorial oficial, códigos y versión se utilizarán? | Carga territorial y filtros confiables. |
| D31 | ¿A qué corresponde «15 de diciembre» en la entrevista? | Fecha aislada; no cambiar la entrega general. |
| D32 | ¿Dominio, responsable institucional, destino/retención de respaldos y horario de servicio? | Producción y continuidad. |
| D33 | ¿Hardware disponible y objetivos finales de rendimiento/recuperación? | Capacidad y criterios operativos. |
| D34 | ¿Cuántas horas reales se han ejecutado y qué hitos están aceptados? | Ajuste de cronograma sin doble conteo. |

### 18.2 Materiales mencionados que faltan

- Modelo Preliminar Horas Sociales v0.2 completo, del 23 de octubre de 2026.
- Respuestas posteriores de la entrevista que el plan cita como base de reglas confirmadas.
- Fotografías/listados de tipos y áreas mencionados por el usuario.
- Listado institucional de carreras, facultades, coordinadores y valores de género autorizados.
- Catálogo territorial verificado con sus relaciones y códigos.
- Archivo académico de estudiantes, si se decide utilizarlo, por un canal apropiado para datos personales.
- Formatos/documentos de horas y evidencias, si se confirma su carga.
- Bitácora real, inventario del equipo de producción y decisiones operativas.

La ausencia de un material no se completa con datos sintéticos fuera de pruebas. Las propuestas de implementación deben indicar qué decisión falta y qué parte sí puede avanzar.

## 19. Cambios y contradicciones entre fuentes

| Tema | Fuente anterior | Decisión o tratamiento vigente |
|---|---|---|
| Cantidad de cuentas | F01/F03 hablaban de cuenta única. | F04 confirma múltiples cuentas, un solo rol ADMIN, acceso completo y JWT. |
| Backend | F01 proponía Node/Fastify/Drizzle. | Instrucciones humanas posteriores usan Go/chi/pgx/Goose. |
| Territorio | Entrevista: «262 municipios». | Propuesta posterior distingue municipio/distrito; catálogo exacto pendiente de verificación. |
| Coordinadores | Entrevista menciona responsable por facultad; un criterio de F03 permite dos por carrera. | F03 también deja cardinalidades pendientes. Conservar la contradicción, no fijar relación final. |
| Contactos | Prompt F03 usa un principal por defecto. | El mismo plan deja multiplicidad pendiente; no tratar ese valor por defecto como aprobación. |
| Tipo y áreas | F03 propone un tipo y varias áreas. | F03 exige confirmar cardinalidad e independencia; continúa pendiente. |
| Libro de horas | PDF inicial dejaba horas/asignaciones por definir. | F02 las solicita; F03 concreta movimientos, meta 300+300 y políticas con distintos niveles de confirmación. |
| Modelo v0.2 | F03 dice basarse en 19 tablas de negocio. | Fuente completa no recuperada; este documento no la sustituye. |
| Sesiones e idempotencia | Modelo de negocio citado sin diccionario disponible. | `sesion_admin` y `solicitud_id` son adiciones técnicas descritas por el plan, no tablas/campos originales demostrados. |
| cloudflared | F01 propone servicio del anfitrión. | F03 propone contenedor; selección posterior del plan, aún sin despliegue demostrado aquí. |
| Versiones | Plan antiguo usa tags de ejemplo por etapa. | F04/F05/F08 establecen GitFlow y convención `VMAJOR.MINOR.PATCH`. |
| Evidencia local | Copia en esta carpeta indica Docker no ejecutado. | Otro chat y repositorio de trabajo documentan ejecución posterior; no mezclar copias. |
| Calidad y recuperación | F01 usa menor volumen, p95 2 s y recuperación en jornada. | F03 plantea metas más exigentes y RTO 4 h como propuestas; faltan acuerdo y mediciones finales. |
| Cronograma | Fechas estimadas para fases pasadas. | La fecha en un cronograma no demuestra ejecución; conciliar con bitácora. |

## 20. Registro de fuentes y continuidad

### 20.1 Fuentes consultadas

| ID | Fuente | Fecha / utilidad |
|---|---|---|
| F01 | Primer avance técnico (`Primer_avance_Horas_Sociales_Actualizado.pdf`, fuente no incluida en este repositorio) | Documento presentado el 15 de octubre de 2026; 14 páginas. Alcance inicial, infraestructura, seguridad, cronograma y entrega. |
| F02 | Chat **Organiza ideas de Horas Sociales**, ID `01a16654-a2f0-77a2-80bb-5df72aaab89b` | Lluvia de ideas humana, interpretación y preguntas de la IA. No contiene en la recuperación las respuestas completas al cuestionario. |
| F03 | [Plan de programación de 20 etapas](Plan_Programacion_IA_Horas_Sociales.md) | Versión 1.0 del 23 de octubre de 2026. Reglas declaradas confirmadas, propuestas, pendientes, prompts y presupuesto. |
| F04 | Chat **Inicializar portal de horas sociales**, ID `01a175c4-464b-7353-9aa2-e7a7c152ef53` | Instrucciones de inicialización, GitFlow, desglose de issues y aclaración humana sobre múltiples ADMIN. |
| F05 | [Aclaraciones vigentes](DECISIONES_VIGENTES.md) | 25 de octubre de 2026. Corrige cuenta única y reconoce fuentes faltantes. |
| F06 | [Índice de planeación](README.md), [contexto IA](CONTEXTO_IA.md), [índice de issues](INDICE_ISSUES.md) y [prompt de exportación](PROMPT_EXPORTAR_CONTEXTO.md) | Secuencia ajustada, tareas pequeñas y requerimientos para consolidar el contexto. |
| F07 | README y validación de la copia inicial (fuentes históricas no incluidas), [validación del repositorio de trabajo](../VALIDACION.md) y chat F04 | Distinguir inicialización preparada de ejecución posterior documentada. |
| F08 | [GitFlow](../GITFLOW.md) | Fuente operativa de ramas, PR y versiones. Consultar contenido vigente al trabajar. |
| F09 | Chat **Plan tecnológico inicial**, ID `6acd15f5-27fc-83e8-96b4-d74d22a4df0e` | Instrucciones de inicialización dockerizada y traslado del trabajo al entorno local. Recuperación limitada a los mensajes entregados por la herramienta. |

Los enlaces del repositorio se adaptaron a referencias relativas al publicarlo. Las fuentes antiguas no incluidas se mencionan sin enlaces locales inexistentes. No incluir secretos, sesiones, contraseñas ni expedientes reales al compartir esta consolidación.

### 20.2 Cómo continuar con este documento

1. Incorporar el modelo v0.2 y materiales faltantes con sus nombres originales.
2. Contrastar las reglas «confirmadas según el plan» con las respuestas originales o ratificación del responsable.
3. Resolver únicamente las preguntas que bloqueen la próxima tarea.
4. Actualizar este documento y el diccionario conservando historial de decisiones.
5. Revisar el issue, sus dependencias, criterios y estado actual del repositorio antes de programar.
6. Implementar una tarea pequeña y verificable; registrar evidencia y validación.

**Instrucción de continuidad:** este documento sirve como contexto consolidado. No autoriza a inventar campos, cambiar reglas pendientes, habilitar otros roles, reconstruir el proyecto desde cero ni declarar como implementado lo que solo está planeado.
