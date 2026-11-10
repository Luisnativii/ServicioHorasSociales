# Contrato de estados y cierres

Revisión 0.1 para aprobación · Issue [#8](https://github.com/Luisnativii/ServicioHorasSociales/issues/8) · 9 de noviembre de 2026

**Estado: propuesta acotada.** El propietario confirmó que una participación abierta puede cerrarse como finalizada o abandonada. Los efectos del cierre global, las fechas y el estado de cancelación requieren aceptación expresa antes de generar SQL.

## Decisiones vigentes

- Un proyecto inicia `ACTIVO` y solo puede salir hacia `ABANDONADO`, `CANCELADO` o `FINALIZADO`.
- No se reabren proyectos terminales.
- Una participación inicia `ACTIVA` y puede cerrarse individualmente como `FINALIZADA` o `ABANDONADA`.
- Abandonar conserva la participación, sus horas y su historia; no permite reinscripción al mismo proyecto.
- Un proyecto se abandona automáticamente cuando tiene al menos una participación y todas están `ABANDONADAS`.
- Cero participaciones no equivale a “todas abandonadas”.
- Un proyecto o participación cerrado no admite nuevas asignaciones ni acreditaciones ordinarias, aunque se informe una fecha anterior.

## Estados propuestos

| Entidad | Estados | Terminales |
| --- | --- | --- |
| Proyecto | `ACTIVO`, `ABANDONADO`, `CANCELADO`, `FINALIZADO` | Todos excepto `ACTIVO` |
| Participación | `ACTIVA`, `ABANDONADA`, `CANCELADA`, `FINALIZADA` | Todos excepto `ACTIVA` |

`CANCELADA` se reserva para una participación activa afectada por la cancelación administrativa del proyecto. No significa abandono del estudiante ni finalización del trabajo.

## Matriz de transiciones

### Participación individual

| Origen | Acción | Destino | Condiciones y efectos |
| --- | --- | --- | --- |
| `ACTIVA` | Finalizar | `FINALIZADA` | Fecha de cierre válida; conserva horas; bloquea nuevas horas y cambios de estado. |
| `ACTIVA` | Abandonar | `ABANDONADA` | Fecha de cierre válida; conserva horas; bloquea nuevas horas y reinscripción. Puede activar el abandono automático del proyecto. |
| Terminal | Cualquier cierre o reapertura | Rechazo | No se sobrescribe el cierre ni se crea una segunda transición. |

La participación no se cancela individualmente. `CANCELADA` solo puede producirse dentro de la transacción que cancela el proyecto.

### Proyecto

| Origen | Acción | Destino | Efecto sobre participaciones |
| --- | --- | --- | --- |
| `ACTIVO` | Finalizar | `FINALIZADO` | Requiere al menos una participación y ninguna `ACTIVA`; conserva `FINALIZADA`/`ABANDONADA` sin reescribirlas. |
| `ACTIVO` | Cancelar | `CANCELADO` | En la misma transacción cambia cada participación `ACTIVA` a `CANCELADA`; conserva las ya terminales. |
| `ACTIVO` | Último abandono | `ABANDONADO` | Automático solo si existe al menos una participación y todas quedan `ABANDONADAS`. |
| Terminal | Reabrir o cerrar otra vez | Rechazo | Conserva el evento original y todos los estados. |

No existe abandono manual del proyecto en esta revisión. El estado `ABANDONADO` se deriva exclusivamente del abandono total.

## Casos sin participantes activos

| Situación | Resultado |
| --- | --- |
| Proyecto sin participaciones | Sigue `ACTIVO`; puede cancelarse, pero no finalizarse ni abandonarse. |
| Una o más participaciones y todas `ABANDONADAS` | Proyecto `ABANDONADO` automáticamente. |
| Todas `FINALIZADAS` | Proyecto sigue `ACTIVO` hasta que el ADMIN lo finaliza. |
| Mezcla de `FINALIZADAS` y `ABANDONADAS`, sin activas | No es abandono total; el ADMIN puede finalizar el proyecto. |
| Existen participaciones `ACTIVAS` | No se puede finalizar; se deben cerrar individualmente o cancelar el proyecto. |

## Fechas e historia

Cada cierre registra:

- `fecha_cierre`: fecha administrativa informada por el ADMIN;
- `cerrada_en`: instante técnico del servidor;
- `cerrada_por`: cuenta ADMIN responsable;
- `motivo`: obligatorio para abandono individual y cancelación de proyecto; opcional al finalizar.

Reglas propuestas:

- `fecha_cierre` no puede ser futura ni anterior a la fecha de inicio/asignación correspondiente.
- No puede ser anterior a la última `fecha_horas` ya acreditada en la participación o proyecto afectado.
- El abandono automático del proyecto usa la misma `fecha_cierre`, actor e instante del último abandono individual.
- Cancelar un proyecto aplica su fecha, actor y motivo a las participaciones activas que pasan a `CANCELADA`.
- Estado e historial se guardan en una sola transacción; las filas se bloquean para coordinar cierres, asignaciones y horas concurrentes.

## Operaciones después del cierre

- Se rechazan nuevas participaciones y horas ordinarias.
- Se rechazan reaperturas, cambios de estado y cambios de fechas de cierre.
- Los datos y movimientos previos permanecen consultables.
- La edición o corrección de horas se rige exclusivamente por #9; no se habilita mediante este contrato.
- Los campos descriptivos de un proyecto terminal quedan en solo lectura. Una corrección administrativa futura necesita un contrato e historial propios.

## Efecto sobre el backlog

La aceptación desbloquea la semántica de #53, #59, #62, #64, #66 y sus interfaces #54/#63/#70. Las migraciones deben conservar eventos y actores; no basta con actualizar una columna de estado sin historia.

## Casos de aceptación

1. Abandonar una de dos participaciones mantiene el proyecto `ACTIVO`.
2. Abandonar la última participación activa cuando todas las existentes están abandonadas cierra el proyecto como `ABANDONADO` en la misma transacción.
3. Un proyecto sin participaciones no se abandona ni finaliza por evaluar un conjunto vacío.
4. Una mezcla de finalizadas y abandonadas puede cerrar el proyecto como `FINALIZADO`, sin convertir abandonos en finalizaciones.
5. Cancelar un proyecto convierte únicamente sus participaciones activas en `CANCELADAS`.
6. Cualquier intento de asignar o acreditar después del cierre se rechaza, incluso con fecha anterior.
7. Una carrera entre acreditar y cerrar deja exactamente una operación válida según el bloqueo transaccional.

## Criterio de aprobación

El propietario debe aceptar expresamente la matriz, el estado `CANCELADA`, las reglas de fecha y el bloqueo de edición posterior. Hasta entonces se mantiene `needs:decision` y no se generan migraciones.
