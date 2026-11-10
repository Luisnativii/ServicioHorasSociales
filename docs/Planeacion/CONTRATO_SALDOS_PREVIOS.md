# Contrato de saldos previos

Revisión 0.1 aceptada · Issue [#10](https://github.com/Luisnativii/ServicioHorasSociales/issues/10) · 10 de noviembre de 2026

**Estado: contrato aceptado.** El propietario aceptó corte inclusivo, procedencia, representación única e idempotencia al solicitar corregir e integrar los PR #128–#134. La fecha exacta y la carga SQL siguen correspondiendo a #95.

## Decisión aceptada

- La carga inicial usa una sola `fecha_corte` global, obligatoria e inclusiva: un saldo resume horas reconocidas hasta esa fecha, incluida.
- La fecha calendario exacta no se inventa en este issue. El propietario debe declararla y aprobarla en #95 antes del ensayo de carga; el manifiesto no puede cambiarla después de aplicarse.
- Los saldos se separan estrictamente en `INTERNA` y `EXTERNA`. No existe un saldo combinado ni compensación entre categorías.
- Cada saldo representa un acumulado conocido de un estudiante y categoría que no está representado por movimientos detallados.
- Las horas son enteras y mayores o iguales que cero. Un cero solo es conocido si la fuente lo declara expresamente; la ausencia de un registro no demuestra cero.
- El saldo no se vincula con un proyecto, participación, territorio o fecha de actividad ficticios.
- La operación diaria posterior al corte usa movimientos ordinarios; el saldo inicial permanece inmutable y separado del libro detallado.

## Una sola representación histórica

Para cada par estudiante/categoría se elige una sola representación hasta la fecha de corte:

| Historia disponible hasta el corte | Representación |
| --- | --- |
| Existe detalle confiable con actividad, proyecto y fecha | Movimientos detallados; no crear saldo para ese par. |
| Solo existe un acumulado confiable | Un saldo previo; no crear movimientos históricos para ese par. |
| El dato es desconocido o no está respaldado | No fabricar saldo cero ni movimientos; registrar el faltante en la conciliación. |

Si existe un saldo previo para un estudiante/categoría, la carga rechaza movimientos de esa misma categoría con `fecha_horas` menor o igual al corte. Los movimientos posteriores al corte sí conviven con el saldo. Recuperar detalle antiguo en el futuro exige una conciliación aprobada que sustituya la representación sin sumarla dos veces.

## Procedencia mínima

Cada saldo debe pertenecer a un lote y conservar:

| Dato | Regla aceptada |
| --- | --- |
| `lote_id` | UUID único de la carga conciliada. |
| `estudiante_id` | Estudiante identificado sin modificar ni normalizar su carnet por intuición. |
| `naturaleza` | Solo `INTERNA` o `EXTERNA`. |
| `horas` | Entero mayor o igual que cero; no nulo para un saldo conocido. |
| `fecha_corte` | La fecha global inclusiva declarada en el lote. |
| `fuente_ref` | Referencia no secreta y verificable al origen; no contiene credenciales ni datos personales innecesarios. |
| `registrado_por` / `registrado_en` | ADMIN responsable e instante técnico de aplicación. |

El manifiesto del lote conserva además nombre y versión de la fuente, responsable de la entrega, cantidad de filas, suma interna, suma externa y huella SHA-256 del contenido canónico. Si una fila combina fuentes distintas, debe identificar la fuente específica de ese saldo; no basta con escribir “migración” o “dato anterior”.

## Idempotencia y duplicados

- `lote_id` es único. Repetirlo con la misma huella devuelve el resultado ya aplicado y no inserta otra vez.
- Repetir `lote_id` con una huella diferente produce conflicto.
- Una huella ya aplicada con otro `lote_id` se rechaza como lote duplicado.
- Dentro del lote solo existe una fila por estudiante/categoría.
- Tras la carga solo existe un saldo inicial por estudiante/categoría. No se sobrescribe para “ajustar” totales.
- Lote, saldos y registro de auditoría se confirman en una sola transacción; un fallo no deja una carga parcial.

Estas reglas reconocen una carga ya aplicada sin depender del nombre del archivo ni del momento en que se vuelva a ejecutar el SQL.

## Conciliación antes y después de cargar

Antes de ejecutar SQL en #95 se genera un informe de ensayo que:

1. declara la fecha de corte y la huella del conjunto;
2. resuelve cada carnet contra un único estudiante y separa los no encontrados o ambiguos;
3. rechaza duplicados, categorías desconocidas, decimales, negativos y campos de procedencia incompletos;
4. detecta conflictos con movimientos detallados hasta el corte;
5. enumera estudiantes/categorías sin dato y no los convierte en cero;
6. presenta cantidad de filas y sumas por categoría para aprobación humana.

Después de la transacción se vuelven a comparar cantidad, sumas y huella con el manifiesto. Cualquier diferencia invalida el resultado; no se corrige manualmente una parte del lote ni se oculta el faltante.

## Efecto en cálculos e informes

- El acumulado vigente por categoría es `saldo previo conocido + movimientos ordinarios posteriores al corte`, más correcciones solo si se aprueba su contrato.
- El saldo contribuye al progreso de 300 internas y 300 externas, sin crear actividad en un proyecto o período.
- Los dashboards de actividad por fechas, proyectos o territorios excluyen los saldos sin historia.
- Un saldo puede demostrar acumulado al corte, pero no la fecha exacta de cada actividad ni cuándo se alcanzó una meta antes del corte.
- Cuando falte saldo para una categoría esperada, el informe debe distinguir “sin dato previo” de cero confirmado. #75 define la presentación final de filtros y estados incompletos.

## Casos de aceptación

1. Un estudiante tiene saldo conocido de 120 internas al corte y luego recibe 20 internas: su acumulado vigente es 140, pero la actividad posterior al corte es 20.
2. Tiene 90 externas detalladas antes del corte: no se crea además un saldo de 90 externas.
3. La fuente no informa horas internas: no se inserta cero ni se declara que el estudiante carecía de horas.
4. Reejecutar el mismo lote y huella no duplica saldos; cambiar una hora con el mismo `lote_id` produce conflicto.
5. Una fila sin fecha, proyecto o territorio histórico conserva únicamente el saldo y la procedencia; no inventa dimensiones para aparecer en reportes de actividad.
6. Un lote con 50 filas falla en la fila 40: no queda ninguna de sus filas aplicada.

## Límites y registro de aprobación

- La carga inicial real, sus respaldos y el procedimiento SQL corresponden a #95.
- Este contrato no crea importador web, endpoints, migraciones ni datos de ejemplo que imiten información real.
- R28 fija las reglas temporales generales de movimientos; este contrato solo define la frontera del corte inicial.
- Corregir o sustituir un saldo ya aplicado requiere un mecanismo auditado que se acuerde antes de implementarlo; no se edita ni elimina directamente.
- El 10 de noviembre de 2026 el propietario solicitó corregir e integrar los siete PR documentales abiertos, incluido #133. Esa instrucción acepta esta revisión; la fecha exacta de corte sigue pendiente para #95 y debe constar en el manifiesto antes de cargar.
