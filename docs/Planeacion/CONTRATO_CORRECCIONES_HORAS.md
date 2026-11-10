# Contrato de correcciones del libro de horas

Revisión 0.1 para aprobación · Issue [#9](https://github.com/Luisnativii/ServicioHorasSociales/issues/9) · 9 de noviembre de 2026

**Estado: propuesta acotada.** Este contrato no habilita edición ni borrado de movimientos, no ejecuta SQL sobre datos reales y no sustituye la decisión del propietario de realizar una carga inicial manual. Debe aprobarse antes de implementar #72/#73.

## Decisión propuesta

- Un movimiento de horas acreditado es inmutable: nunca se actualiza ni se elimina.
- Una corrección crea un reverso auditable vinculado al movimiento original y, cuando corresponde, un movimiento de reemplazo vinculado a la misma corrección.
- Solo una corrección efectiva puede consumir cada movimiento original. Un segundo intento se rechaza.
- La corrección conserva el movimiento original, su categoría, fecha, participación, actor e instante de registro.
- Cada operación registra motivo obligatorio, ADMIN responsable, instante del servidor e identificador idempotente de solicitud.
- El reverso compensa exactamente las horas del original; no acepta una cantidad libre ni una categoría distinta.
- El reemplazo pertenece a la misma participación y puede corregir horas, fecha o categoría. Sus horas deben ser positivas y no pueden superar las horas del original.
- Omitir el reemplazo anula por completo la acreditación errónea. Una corrección nunca aumenta el total global de horas del estudiante.

No se expone una operación genérica de valores negativos. El reverso solo existe como parte del caso de uso de corrección y siempre referencia un movimiento ordinario válido.

## Corrección frente a acreditación tardía

La participación y el proyecto activos siguen siendo requisito para una acreditación ordinaria. Tras el cierre se permite corregir un movimiento que ya existía antes del cierre, pero no agregar horas nuevas ni aumentar su cantidad.

| Operación después del cierre | Resultado propuesto |
| --- | --- |
| Anular un movimiento erróneo | Permitido mediante reverso completo sin reemplazo. |
| Reducir su cantidad | Permitido mediante reverso y reemplazo menor. |
| Cambiar `INTERNA` por `EXTERNA` conservando horas | Permitido mediante reverso y reemplazo; puede cambiar el cumplimiento por categoría. |
| Corregir la fecha de actividad | Permitido si la fecha corregida satisface el contrato temporal que se apruebe en R28 y no es posterior al cierre. |
| Aumentar horas o registrar una actividad ausente | Rechazado; es acreditación tardía, no corrección. |
| Cambiar de estudiante, proyecto o participación | Rechazado; requiere anular el error y resolver el dato correcto por un flujo autorizado, nunca reubicar historia. |

La corrección no reabre el proyecto ni la participación, no cambia su estado de cierre y no altera sus fechas. Los reportes recalculan los totales vigentes y conservan una vista de auditoría con original, reverso y reemplazo.

## Modelo candidato

La implementación puede representar el reverso y el reemplazo como movimientos relacionados, pero el SQL definitivo pertenece a #72. Como mínimo debe preservar:

| Dato | Regla propuesta |
| --- | --- |
| `correccion.id` | Identificador único de la operación administrativa. |
| `movimiento_original_id` | FK obligatoria; único entre correcciones efectivas. |
| `movimiento_reverso_id` | FK obligatoria y única al movimiento que compensa exactamente el original. |
| `movimiento_reemplazo_id` | FK opcional; ausente en anulación total. |
| `motivo` | Texto obligatorio, no vacío. |
| `corregido_por` / `corregido_en` | ADMIN responsable e instante técnico inmutables. |
| `solicitud_id` | Clave idempotente única; el mismo cuerpo devuelve el mismo resultado y otro cuerpo produce conflicto. |

Original, reverso, reemplazo y registro de corrección se guardan en una sola transacción. La operación bloquea el original para que dos solicitudes concurrentes no lo corrijan dos veces. La cuenta ADMIN se conserva como referencia aunque después se desactive.

## Efecto sobre categoría e historia

- El original mantiene la naturaleza capturada cuando fue acreditado.
- El reverso usa esa misma naturaleza y neutraliza el original en cálculos vigentes.
- El reemplazo guarda la naturaleza corregida; no se recalcula desde la naturaleza actual del proyecto.
- Los totales se derivan del libro completo. No se escribe un total editable en el estudiante.
- La actividad histórica puede mostrarse como registrada y como vigente. Un reporte por fecha debe declarar cuál vista usa; #75 fija la semántica final de filtros.
- Una corrección posterior al cierre puede cambiar retroactivamente el cumplimiento por categoría. El evento de auditoría permite explicar el cambio sin fingir que el reemplazo se capturó en la fecha original.

## Casos de aceptación

1. Se registraron 12 horas externas en lugar de 8: el reverso neutraliza 12 y el reemplazo acredita 8 externas; el efecto neto es 8.
2. Un movimiento duplicado de 6 horas se anula sin reemplazo; el original sigue visible en auditoría y deja de sumar.
3. Se clasificaron 10 horas como internas y eran externas: reverso de 10 internas y reemplazo de 10 externas; el total global no aumenta.
4. Después del cierre se intenta cambiar 8 por 11 horas: se rechaza por ser un aumento tardío.
5. Dos ADMIN intentan corregir el mismo original: solo una transacción puede confirmarse; la otra recibe conflicto.
6. Reintentar la misma `solicitud_id` con el mismo cuerpo no duplica reversos; reutilizarla con otro cuerpo produce conflicto.

## Límites y aprobación

- La carga inicial manual indicada por el propietario no se ejecuta ni se diseña en este issue; #10/#95 definen saldos y conciliación.
- No se corrigen saldos previos con este contrato hasta que #10 defina su procedencia y granularidad.
- R28 debe fijar los rangos exactos de fecha antes de implementar correcciones de fecha.
- La matriz de cierre de #8 es una propuesta separada. Este contrato solo exige distinguir acreditación ordinaria de corrección administrativa.
- Hasta la aceptación expresa del propietario, R27 permanece como propuesta, el issue conserva `needs:decision` y no se generan migraciones ni endpoints.
