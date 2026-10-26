# Registro de horas de desarrollo

Revisión 1 · Inicio del registro: 26 de octubre de 2026 · America/Guatemala · Issue [#6](https://github.com/Luisnativii/ServicioHorasSociales/issues/6).

Este registro contabiliza trabajo de desarrollo del proyecto. Las metas estudiantiles de 300 horas internas y 300 externas se calculan en el sistema de negocio y no se registran aquí.

## Presupuesto e información disponible

- Presupuesto orientativo del proyecto completo: **300 horas**, ya incluidas las etapas iniciales y su validación. No comienza otro presupuesto al abrir el backlog.
- Histórico anterior: existen entregas y evidencia, pero no se dispone aquí de duraciones reales confirmadas. Se solicitó al propietario si conserva un registro.
- Total real confirmado y saldo global: **no calculables todavía**. Un dato desconocido no equivale a cero ni se reemplaza por la estimación de su etapa.
- La duración de CI, fechas de commits, tiempo del chat o tiempo de respuesta de la IA no acreditan tiempo trabajado por el propietario.

## Avance histórico por conciliar

Estas filas son referencias de entregas, **no imputaciones de horas**. No se suman al registro real de sesiones.

| Referencia | Trabajo acreditado | Duración real | Estado del dato |
| --- | --- | --- | --- |
| #1 / V0.0.1 | Inicialización, Docker, base mínima y CI; evidencia en [VALIDACION.md](../VALIDACION.md) | Sin dato confirmado | Pendiente de registro del propietario. |
| #2 | Validación manual de la base | Sin dato confirmado | Aceptación manual pendiente; no dar por realizada. |
| #3 / PR #100 | Desglose del plan, fuentes e índice de issues | Sin dato confirmado | Entrega integrada; duración por conciliar. |
| #5 / PR #101 | Consolidación de reglas y propuesta nueva de modelo | Sin dato confirmado | Entrega integrada; duración por conciliar. |
| #6 | Alcance y preparación de este registro | Sin dato confirmado | Sesión en curso; no inferir duración. |

Si una duración histórica se recupera, registrar su sesión con referencia al issue y marcar aquí «conciliada con H-...». No agregar otra vez el total de la entrega además de sus sesiones.

## Sesiones reales

**Todavía no hay sesiones con duración confirmada registradas.** La tabla está intencionalmente vacía; los encabezados sirven para la primera entrada real.

| ID | Fecha local | Issue | Actividad concreta | Horas reales | Estado y fuente de confirmación |
| --- | --- | --- | --- | --- | --- |

Para añadir una entrada, usar un ID único `H-0001`, `H-0002`, etc.; fecha en formato `AAAA-MM-DD`; número de issue; actividad específica; horas decimales no negativas; y confirmación del propietario. Si se conoce solo una aproximación, marcarla **estimada**, fuera del total real confirmado, hasta su validación.

## Cálculo y correcciones del registro

1. Sumar únicamente sesiones **confirmadas**, identificadas sin duplicados. Un tramo de trabajo pertenece a una entrada; si abarca varios issues, repartir su duración real o señalar una referencia principal sin duplicar el tiempo.
2. Mientras haya trabajo histórico sin dato, mostrar el subtotal conocido como parcial y no afirmar el saldo total del proyecto.
3. Con histórico completo, calcular saldo orientativo como `300 - total_real_confirmado`. Si es negativo, mostrar el exceso y revisar la estimación; no ocultarlo ni reiniciar el presupuesto.
4. Corregir una duración con motivo y referencia al dato anterior, manteniendo trazabilidad en Git. No completar tiempos por deducción.
5. Revisar la estimación restante con el propietario al terminar un bloque. Cambiar una estimación de alcance requiere registrar la decisión, no alterar retroactivamente horas reales.

Actualizar este documento al confirmar sesiones no modifica el libro de horas sociales de estudiantes ni autoriza implementar funciones de negocio.
