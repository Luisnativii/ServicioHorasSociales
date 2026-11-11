# Contrato de historia académica y clasificación temporal

Revisión 0.1 propuesta · Issue [#12](https://github.com/Luisnativii/ServicioHorasSociales/issues/12) · 10 de noviembre de 2026

**Estado: propuesta para aprobación.** Este documento define la semántica temporal necesaria para estudiantes, proyectos, horas y reportes. No crea tablas, no ejecuta migraciones y no aprueba campos del expediente que pertenecen a otros bloques.

## Decisiones propuestas

| ID | Tema | Propuesta |
| --- | --- | --- |
| D-HT-01 | Academia vigente e historia | El expediente muestra la carrera y facultad vigentes. Cada cambio académico crea una vigencia nueva y no modifica vigencias anteriores. |
| D-HT-02 | Participación y movimientos | La participación conserva la vigencia académica con la que se asignó al estudiante. Cada movimiento se atribuye a la vigencia académica y clasificación del proyecto aplicables en `fecha_horas`. |
| D-HT-03 | Clasificación de proyecto | Tipo, área y modalidad tienen historia de vigencias. Los cambios son prospectivos y no trasladan actividad anterior a la clasificación nueva. |
| D-HT-04 | Actividad por período | La actividad pertenece al período de `fecha_horas`; `registrado_en` sirve únicamente para auditoría. Los saldos sin historia no se atribuyen a un período. |
| D-HT-05 | Cumplimiento por período | “Completó en el período” significa que alcanzó por primera vez 300 horas internas y 300 externas en una fecha conocida dentro del período. Si el saldo previo ya demuestra cumplimiento sin fecha histórica, la fecha de cumplimiento queda desconocida y no se imputa a un período. |

## Historia académica

El estudiante mantiene una sola vigencia académica activa. Una vigencia conserva, como mínimo:

- estudiante;
- carrera y facultad aplicables;
- `vigente_desde` inclusiva y `vigente_hasta` exclusiva, sin solapamientos;
- instante de registro y cuenta ADMIN responsable;
- motivo del cambio.

La carrera actual del expediente corresponde a la vigencia abierta. No se guarda una facultad actual independiente que pueda contradecir la carrera; la facultad histórica sí se conserva en la vigencia para evitar que un cambio futuro en el catálogo reescriba reportes anteriores.

Un cambio de carrera:

1. cierra la vigencia anterior;
2. abre una nueva vigencia desde la fecha efectiva;
3. no cambia participaciones, movimientos ni reportes anteriores;
4. se rechaza si crea un vacío, un solapamiento o intenta alterar una fecha que ya tiene actividad acreditada bajo otra vigencia.

La participación referencia la vigencia académica existente al asignar al estudiante. Esto permite mostrar la academia de ingreso al proyecto aunque el expediente cambie después. Los movimientos usan la vigencia que cubre su `fecha_horas`, de modo que un estudiante puede continuar en la misma participación y atribuir actividad posterior a su nueva carrera sin reescribir la anterior.

## Historia de clasificación del proyecto

Cada proyecto conserva vigencias de clasificación con:

- tipo, área y modalidad aplicables;
- fecha efectiva de inicio y fin;
- instante de registro, ADMIN responsable y motivo.

La clasificación vigente se muestra en el proyecto y se usa para actividad nueva. Un cambio de tipo, área o modalidad cierra la vigencia anterior y abre otra; no edita la vigencia que ya respalda movimientos. La fecha efectiva no puede retroceder sobre actividad acreditada con otra clasificación.

Cada movimiento referencia la vigencia de clasificación que cubre `fecha_horas`. La naturaleza `INTERNA` o `EXTERNA` continúa copiada directamente en el movimiento conforme a R13; no se deduce de esta historia.

## Semántica de reportes

| Consulta | Regla temporal |
| --- | --- |
| Estudiantes o proyectos actuales | Usa carrera, facultad, tipo, área y modalidad vigentes. |
| Actividad de un período | Filtra por `fecha_horas` y agrupa usando las vigencias académica y de clasificación referenciadas por cada movimiento. |
| Capturas administrativas | Puede filtrar `registrado_en`, pero debe nombrarse como fecha de registro y no como fecha de actividad. |
| Avance acumulado | Suma movimientos vigentes y saldos previos aprobados, sin atribuir saldos sin historia a una carrera, clasificación o período inventados. |
| Participaron en el período | Cuenta estudiantes únicos con al menos un movimiento vigente cuya `fecha_horas` cae en el período. |
| Completaron en el período | Cuenta estudiantes cuya primera fecha conocida de cumplimiento de ambas metas cae en el período. |

Los filtros académicos y de clasificación aplicados a actividad usan el valor histórico del movimiento. Un estudiante o proyecto que cambió durante el período puede aparecer en más de un grupo, pero el total de horas no se duplica y el indicador de estudiantes únicos se deduplica según el alcance explícito de la consulta.

## Cumplimiento y saldos previos

El cumplimiento no es un booleano editable. Se deriva del libro vigente y requiere simultáneamente al menos 300 horas internas y 300 externas.

- La fecha de cumplimiento es la primera `fecha_horas` conocida en la que ambas sumas alcanzan la meta.
- Una corrección puede cambiar esa fecha; el reporte se recalcula y la auditoría explica el cambio.
- Un saldo previo aporta al acumulado desde su corte, pero no inventa las fechas de las actividades que lo originaron.
- Si ambos umbrales ya están cubiertos únicamente por saldos sin historia, solo puede afirmarse “cumplimiento conocido al corte”, con fecha histórica desconocida; no “completó en el año”.
- Si el saldo no completa ambas metas y un movimiento posterior conocido completa la categoría faltante, esa `fecha_horas` sí determina el primer cumplimiento conocido.

## Casos de aceptación

1. Un estudiante cambia de carrera el 1 de julio: sus horas de junio permanecen en la carrera anterior y las de julio se atribuyen a la nueva.
2. El estudiante continúa en el mismo proyecto después del cambio: la participación conserva la academia de asignación y cada movimiento conserva la academia de su fecha.
3. Un proyecto cambia de área el 1 de agosto: la actividad de julio conserva el área anterior y la de agosto usa la nueva.
4. Un movimiento con `fecha_horas` de marzo registrado en abril pertenece a marzo; abril solo aparece en auditoría de captura.
5. Un saldo sin historia ya cubre 300 internas y 300 externas al corte: el estudiante cumple en el acumulado, pero no se cuenta como completado en un mes o año concreto.
6. Un saldo aporta 300 internas y 250 externas; un movimiento de 50 externas el 15 de mayo fija el primer cumplimiento conocido en esa fecha.
7. Una corrección reduce horas por debajo de la meta: el cumplimiento y su fecha se recalculan sin borrar el resultado anterior de auditoría.

## Efecto sobre el backlog

La aprobación de D-HT-01 a D-HT-05 permite diseñar la historia física en #55 y #65, aplicar cierres sin perder contexto en #53, y definir agregaciones en #74/#75. La migración debe impedir vigencias solapadas y referencias temporales contradictorias; este issue no selecciona todavía la implementación SQL exacta.

Los nombres oficiales de tipos y áreas continúan pendientes de #41. La obligatoriedad de género, contactos y demás campos del estudiante sigue fuera de este contrato.

## Registro de aprobación pendiente

Para aceptar esta revisión, el propietario debe confirmar D-HT-01 a D-HT-05. Hasta entonces #12 conserva `needs:decision` y los módulos dependientes no deben tratar esta propuesta como esquema aprobado.
