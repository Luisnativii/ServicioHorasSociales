# Fuentes y trazabilidad de la consolidación

Revisión 1 · 26 de octubre de 2026 · America/Guatemala · Issue [#5](https://github.com/Luisnativii/ServicioHorasSociales/issues/5).

## Estado de esta entrega

**CONFIRMADO:** el propietario indicó en esta sesión que la lógica de negocio ya fue revisada y coincide con el contexto disponible. Autorizó omitir la recuperación de la exportación `Logica_Negocio.md`.

**PROPUESTA:** el [modelo nuevo](MODELO_DATOS_PROPUESTO.md) se presenta para revisión. La confirmación de la lógica no convierte sus nombres físicos, columnas, tipos, restricciones o alternativas en un esquema aprobado.

**PENDIENTE:** las decisiones que el contexto todavía deja abiertas se conservan en sus issues. La aceptación de este issue no sustituye las aprobaciones de #7–#13 o #20 ni autoriza migraciones de negocio.

## Inventario de fuentes

| ID | Fuente y procedencia | Disponibilidad y uso |
| --- | --- | --- |
| F01 | [Plan original](Plan_Programacion_IA_Horas_Sociales.md), conservado por #3 / PR #100 e integrado en `develop` en V0.0.2 | Disponible. Reglas funcionales, etapas y casos; sus decisiones abiertas siguen pendientes y sus instrucciones antiguas se interpretan con las aclaraciones posteriores. |
| F02 | [Issue #5](https://github.com/Luisnativii/ServicioHorasSociales/issues/5), apartado «Decisiones humanas vigentes — 25 de octubre de 2026», consultado el 26 de octubre | Disponible. Registra múltiples ADMIN, correo/contraseña, alta por herramienta, desactivación de la última cuenta y autorización para diseñar un modelo nuevo. Es un registro de decisiones; no es el diccionario v0.2. |
| F03 | Confirmación directa del propietario en este chat, 26 de octubre de 2026 | «omitelo eso ya se reviso y se confirmo la logica del negocio de horas sociales y concuerdan con el contexto que ya tenes». Permite consolidar el contexto disponible sin exigir otra exportación. |
| F04 | [GitFlow](../GITFLOW.md), `AGENTS.md` y [decisiones de inicialización](../DECISIONES.md) | Disponibles. Regulan trabajo, versiones y avance técnico existente. No definen nuevas reglas funcionales. |
| F05 | Modelo v0.2 e entrevistas originales citados por el plan | No recuperados. No se atribuyen columnas ni cardinalidades a esos documentos. El modelo propuesto se identifica como nuevo. |
| F06 | Exportación `Logica_Negocio.md`, mencionada en la revisión histórica del issue #5 | No incorporada a esta copia. Su recuperación se omite por F03. No se recrea un supuesto original ni se afirma haber verificado su contenido en esta sesión. |
| F07 | Aclaraciones directas del propietario al trabajar [#6](https://github.com/Luisnativii/ServicioHorasSociales/issues/6), 26 de octubre de 2026 | Carga inicial por consultas SQL; operación posterior por ADMIN y Excel de dashboards. Dispone de actividades estudiantiles con fechas y proyectos. Aclaró que son horas sociales de estudiantes y que el desarrollo no tiene registro confirmado; no se mezclan ambos registros. |

El issue #5 registra para F06 un SHA-256 histórico `b3336e532201781859f835da6ad888a75a865be1b867cb6e4cd1b8030d334f3b`. Se conserva como referencia atribuida al issue, sin declararlo verificado aquí. No es el hash de ninguno de los documentos nuevos.

### Integridad del plan original

El archivo F01 permanece sin editar. Su blob Git en V0.0.2 es `9e605bfb3ae1fe6b2fa4461fbb6b0b721ffcd965`. El SHA-256 del texto UTF-8 con saltos LF es `76c5776d400522bc5c80ff5e3a717041f8468c634a2bfd21a4b9ed4b1c6c8653`, coincidente con el registrado en el índice. Un checkout con CRLF puede tener un SHA-256 de archivo distinto sin modificar el blob versionado.

## Cómo interpretar los estados

- **CONFIRMADO:** regla funcional respaldada por F02 o por la ratificación F03 del contexto vigente de F01. La aprobación funcional no selecciona por sí sola una representación SQL.
- **PROPUESTA:** diseño técnico nuevo o alternativa que se presenta al propietario. Incluye todo el diccionario de datos de esta entrega.
- **PENDIENTE:** falta una decisión o aprobación identificada. Se indica su issue responsable y la implementación que bloquea.

Las [reglas consolidadas](REGLAS_CONSOLIDADAS.md) asignan identificadores para vincular cada propuesta con su origen. Una decisión posterior debe citar las reglas o propuestas afectadas y actualizar su estado, sin reescribir F01.

## Contradicciones y sustituciones

| Referencia anterior | Aclaración que prevalece | Efecto |
| --- | --- | --- |
| F01: una sola cuenta ADMIN | F02: múltiples cuentas con un único rol ADMIN | No crear restricción de cuenta única ni roles adicionales. |
| Identificador de acceso y alta todavía abiertos | F02: correo y contraseña; creación exclusivamente por herramienta en Docker | No crear formulario ni endpoint web de alta de cuentas. Los detalles de validación corresponden a #20. |
| Posible protección de la última cuenta activa | F02: se permite desactivarla y recuperar por herramienta | No imponer un mínimo de una cuenta activa. El efecto sobre sesiones se acuerda en #20. |
| Recuperar el modelo v0.2 como condición de diseño | F02: diseñar y revisar un modelo nuevo | No llamar «v0.2 recuperado» a la propuesta ni fijar un número de 19 tablas. |
| Exportación de lógica todavía solicitada | F03: omitirla; lógica revisada y confirmada | La ausencia de F06 no bloquea esta consolidación. Las decisiones que ya estaban abiertas siguen identificadas. |
| Versiones Git sugeridas por etapa en F01 | F04: tags `VMAJOR.MINOR.PATCH` publicados por CI tras integrar un PR | No crear tags por adelantado ni modificar versiones publicadas. |
| F02 decía que #3 / PR #100 aún no estaban integrados | #3 cerrado y PR #100 integrado en `develop`, V0.0.2 | La dependencia de inicio de #5 está satisfecha. |

## Pendientes que conserva el issue #5

El mapa siguiente procede del resumen de pendientes publicado en F02. Los identificadores D pertenecen a ese registro histórico; aquí no se afirma reproducir las 34 preguntas completas de una exportación ausente.

| Referencia histórica | Decisión que falta concretar | Responsable y bloqueo |
| --- | --- | --- |
| D01–D02 | Revisar el modelo nuevo y registrar aceptación por bloques | #5; migraciones que usen sus campos o relaciones. |
| D03, D23, D26–D28, D31 | Alcance, terminología, metas, PDF, documentos, importación y fecha aislada | #6; cualquier ampliación de funcionalidad. |
| D04–D05, D20 | Cardinalidad de tipos/áreas y clasificación histórica | #7 y #12; #40, #41, #50 y agregaciones afectadas. |
| D06 | Modalidad y habilidades blandas | #13; sus campos, catálogos o pantallas si se aprueban. |
| D07–D09 | Contactos y asociaciones de coordinadores | #11; #44, #46 y #48. |
| D10–D15 | Cierre de participaciones, fechas y edición después del cierre | #8; #53, #59, #62 y #64. |
| D16 | Correcciones y operaciones posteriores al cierre | #9; #72 y #73. |
| D17–D18 | Corte, procedencia y prevención de duplicados en saldos previos | #10; #68 y #95. |
| D19, D24 | Historia académica y cumplimiento a una fecha | #12; #55, #65 y reportes históricos. |
| D21 | Campos obligatorios del expediente y reglas exactas del carnet | #5 y #6; #55 y #56. |
| D22 | Fechas y límites de acreditación | #5 y #8; #66 y #67. |
| D25 | Límites y semántica de filtros | #12 y #75; dashboard y Excel. |
| D29 | Contrato de sesiones y operaciones de cuentas restantes | #20; migraciones y servicios de identidad. Las decisiones de F02 ya prevalecen. |
| D30 | Catálogo territorial oficial y sus códigos | #37; carga territorial. |
| D32–D34 | Equipo, respaldo, dominio y registro de horas reales | #14, #92, #94 y #6; pruebas operativas y despliegue. |

## Revisión y aprobación

- [x] Dependencia #3 integrada en `develop` y etiquetada V0.0.2.
- [x] Confirmación F03 registrada sin exigir recuperar F06.
- [x] Fuentes originales conservadas; modelo ausente identificado.
- [x] Reglas y propuestas diferenciadas y vinculadas a fuentes.
- [x] Propietario acepta la consolidación documental y el modelo como propuesta inicial; criterio final de #5 marcado e integración por PR #101 en V0.0.3. Las columnas y decisiones técnicas conservan su aprobación independiente.
- [ ] Aprobaciones por bloque registradas con referencia a las propuestas concretas.

La consolidación de #5 ya está integrada. Las implementaciones posteriores deben verificar las aprobaciones de su bloque y sus propios issues de decisiones; cerrar #5 no autoriza por arrastre todo el SQL del modelo propuesto.
