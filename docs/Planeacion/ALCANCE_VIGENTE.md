# Alcance vigente y avance del proyecto

Revisión 1 · 26 de octubre de 2026 · America/Guatemala · Issue [#6](https://github.com/Luisnativii/ServicioHorasSociales/issues/6).

Este documento delimita el trabajo administrativo previsto y distingue lo que ya existe de lo que todavía debe implementarse. Las [reglas consolidadas](REGLAS_CONSOLIDADAS.md) y las [decisiones vigentes](DECISIONES_VIGENTES.md) conservan sus estados y fuentes; el [modelo nuevo](MODELO_DATOS_PROPUESTO.md) continúa identificado como propuesta, con aprobación técnica por bloques.

## Usuarios, acceso y términos

| Elemento | Alcance confirmado y referencia |
| --- | --- |
| Administradores | Varias cuentas con un único rol ADMIN y acceso completo. Login por correo y contraseña; JWT y sesiones revocables. R01–R05; contrato restante en #20. |
| Alta y recuperación de cuentas | Herramienta de administración ejecutada en Docker. No hay formulario ni endpoint web para crear cuentas. Se permite desactivar la última cuenta activa. R03–R04. |
| Estudiantes | Expedientes administrados por ADMIN; no son cuentas de acceso ni disponen de portal propio. R01 y R16. |
| Coordinadores y contactos | Personas y vínculos del negocio; no obtienen acceso por estar registrados. Campos y asociaciones pendientes de #11. |
| Institución / organización | Dos nombres del mismo módulo de instituciones, según el alcance solicitado en #6. No crear dos catálogos o módulos duplicados. La sede institucional y la ubicación de ejecución de cada proyecto se distinguen. R09–R10. |
| Proyecto / participación | El proyecto organiza la actividad; la participación vincula un estudiante con un proyecto. Contar participaciones no equivale a contar estudiantes únicos. R18 y R29. |
| Horas / saldos previos | El libro acredita actividad por participación. Los saldos previos se tratan aparte, con su corte y procedencia pendientes de #10. R22–R26. |

No se incorpora acceso de estudiantes o coordinadores, catálogo público, registro público, roles adicionales ni acceso con Google/Gmail. Estas exclusiones prevalecen sobre cualquier deducción a partir de los nombres de las entidades.

## Funciones de la entrega planificada

Las filas siguientes son alcance futuro, no funciones declaradas como implementadas. Cada cambio se realiza en su issue y solo cuando sus decisiones y dependencias estén resueltas.

| Área | Resultado previsto | Issues de ejecución |
| --- | --- | --- |
| Identidad administrativa | Herramienta de cuentas, sesiones y acceso privado | #20–#33. |
| Catálogos | Academia, territorio, tipos/áreas y coordinadores | #34–#45; decisiones #7, #11 y #12. |
| Instituciones | Registro, consulta, sede y contactos independientes de proyectos | #46–#49; decisión #11. |
| Proyectos | Registro, clasificación, ubicación, cupo informativo e historial | #50–#54; decisiones #7, #8 y #12. |
| Estudiantes | Alta y edición desde la interfaz administrativa, carnet como texto, búsqueda y expediente | #55–#58; campos e historia aún sujetos a aprobación. |
| Participaciones | Asignación, abandono e integración de cierres | #59–#64; decisión #8. |
| Horas y saldos | Acreditación, idempotencia, categorías independientes y procedencia | #65–#71; decisiones #10 y #12. |
| Correcciones | Solo si se aprueba su mecanismo; no confundir corrección con hora tardía | #9 antes de #72–#73. |
| Reportes | Dashboard y exportación XLSX con las mismas reglas, filtros y cifras | #74–#79; contratos temporales #12 y #75. |
| Calidad y operación | Pruebas, recuperación, releases, servidor y entrega | #80–#99. Las operaciones reales conservan su autorización específica. |

### Carga inicial, operación diaria y exportación

**CONFIRMADO por el propietario el 26 de octubre de 2026 al trabajar #6:** la introducción de los datos iniciales la hará mediante consultas SQL; el administrador también debe poder introducir datos desde la aplicación durante su operación. Excel exportará los datos de los dashboards, conservando sus filtros y las mismas cifras. La carga inicial por SQL no sustituye los formularios administrativos previstos ni requiere un importador web.

La carga inicial SQL se concreta y ensaya en #95, después de aprobar el modelo y sus reglas. Conserva las fuentes, las categorías y las fechas históricas reales cuando exista ese detalle. `fecha_horas` y el instante técnico de incorporación no significan lo mismo. Si solo hay acumulados, #10 debe definir su representación y corte antes de cargarlos; no convertir un saldo sin historia en movimientos con fechas o proyectos inventados. El detalle disponible de la fuente se verificará en esos issues.

PDF, adjuntos y un importador masivo desde la interfaz no forman parte de las funciones solicitadas para esta entrega. Una ampliación requiere decisión y tarea específica. Modalidad y habilidades blandas conservan la decisión propia #13.

## Avance que se reutiliza

Corte: `develop` en **V0.0.3**, con #3 / PR #100 y #5 / PR #101 integrados. La evidencia de arranque original corresponde al equipo y fecha documentados en [VALIDACION.md](../VALIDACION.md); no se presenta como una nueva ejecución local durante #6.

| Entrega | Estado comprobado y límite |
| --- | --- |
| Base V0.0.1 / #1 | Inicialización Docker, Go/React/PostgreSQL, esquema vacío, salud real, persistencia y CI existentes. No se reconstruyen ni se contabilizan otra vez. |
| Aceptación manual / #2 | Issue abierto en su rama existente. Se conserva la evidencia automática, sin marcar realizadas las comprobaciones manuales del propietario. |
| Planeación / #3 | Integrada por PR #100; V0.0.2. Índice y prompts disponibles. |
| Consolidación / #5 | Integrada por PR #101; V0.0.3. El propietario marcó cumplida su aceptación documental. Las decisiones abiertas y propuestas de esquema no quedan aprobadas por arrastre. |
| Seguimiento / #4 | Índice general abierto durante la ejecución de los hitos; no es el siguiente módulo que deba implementarse entero. |
| Alcance / #6 | Esta entrega documental. No añade autenticación, tablas de negocio ni pantallas administrativas. |

La aplicación disponible sigue siendo la base técnica. Los catálogos de negocio, CRUD, login, acreditación de horas y reportes requieren sus implementaciones posteriores aunque ya exista su documentación.

## Tres cantidades diferentes

| Cantidad | Significado y tratamiento |
| --- | --- |
| 300 horas de desarrollo | Presupuesto orientativo del proyecto completo, incluido el trabajo previo y la validación. Las veinte etapas del plan suman 300; el desglose en issues no agrega otras 300 ni acredita horas consumidas. |
| Horas reales de desarrollo | Duraciones registradas y confirmadas por el propietario en [REGISTRO_HORAS.md](REGISTRO_HORAS.md). No se calculan a partir de commits, tiempo de conversación, CI o estimaciones de etapas. |
| 300 internas y 300 externas | Meta estudiantil de 600 horas sociales, por categorías independientes. No es el presupuesto de programación y el exceso de una categoría no compensa la otra. R25. |

No se afirma un porcentaje de avance ni un saldo de presupuesto mientras el histórico de horas reales esté incompleto. El registro distingue dato desconocido, estimación y duración real confirmada.

## Reglas transversales y límites de esta entrega

- Consultar el contrato aprobado del issue antes de implementar; conservar las propuestas y decisiones pendientes con su estado explícito.
- Operaciones de negocio privadas con el contrato de identidad previsto. No añadir usuarios, roles ni accesos públicos por deducción.
- Conservar historia: desactivación de catálogos usados, categoría original de horas, actores y cambios de estado según sus decisiones aprobadas. Evitar borrados destructivos y cambios retroactivos de totales.
- Diferenciar fecha de actividad e instante de captura, estudiante único y participación, actividad del período y acumulado. Los contratos precisos se acuerdan en #8, #12 y #75.
- Cupo informativo; horas enteras; reintentos sin duplicados; operaciones concurrentes coherentes. Implementar y probar estos comportamientos en sus tareas, no darlos por existentes aquí.
- Mantener SQL parametrizado y compilación/ejecución de Go, Node y PostgreSQL en Docker. Conservar secretos, `.env` y volúmenes; los datos reales quedan fuera de documentación y pruebas.
- Issue y rama propios, PR a `develop`, título convencional, `Closes #NUMERO`, una etiqueta de versión y checks requeridos. #6 no modifica CI ni las protecciones de integración. Preparar el PR sin esperar activamente a Docker no equivale a declararlo aprobado.

## Decisiones que siguen abiertas

Este issue documenta sus fronteras; no sustituye las decisiones específicas ni elimina sus dependencias.

| Decisión | Responsable antes de implementar |
| --- | --- |
| Campos, obligatoriedades y restricciones del modelo propuesto | Aprobación por bloque y contrato del issue de implementación; registrar cambios en #5/#6 o en el issue responsable. Antes de #55/#56 deben quedar definidos los campos del expediente; no basta con cerrar esta documentación. |
| Tipo/áreas y cardinalidades | #7, antes de #40/#41/#50. |
| Estados y cierres; fechas de actividad y operaciones tras cierre | #8, antes de #53/#59/#64/#66. |
| Correcciones/reversos | #9, antes de #72/#73. |
| Corte, procedencia y duplicados de saldos previos | #10, antes de #68/#95. |
| Contactos y asociaciones de coordinadores | #11, antes de #44/#46/#48. |
| Historia académica/clasificación y cumplimiento a una fecha | #12, antes de #55/#65 y reportes afectados. |
| Modalidad y habilidades blandas | #13; sin implementación por deducción. |
| Equipo objetivo y metas de calidad/recuperación | #14; no inferir capacidad por el equipo donde se edita este documento. |
| Contrato de identidad, normalización y sesiones | #20, antes de #21–#29. |
| Semántica final de filtros y consultas | #75, junto con #7/#12. |
| Carga inicial SQL, corte y conciliación | #95 junto con #10; no ejecutarla ni diseñar un importador web en este issue. |
| PDF, adjuntos e importador de uso diario | Fuera del alcance solicitado de esta entrega; una implementación adicional requerirá decisión e issue específico. |

## Comprobación del issue #6

El criterio de esta entrega se verifica leyendo el alcance, el avance acreditado, la separación de horas y la tabla de decisiones. No declara implementadas las funciones futuras, no cierra #2 y no convierte un modelo propuesto en SQL aprobado.
