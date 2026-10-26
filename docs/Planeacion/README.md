# Planeación del Portal de Horas Sociales

Este directorio conserva el plan original y organiza su ejecución en tareas pequeñas para IA. Solo se implementa un issue por sesión y después de comprobar sus dependencias.

## Documentos

- [Plan original de 20 etapas](Plan_Programacion_IA_Horas_Sociales.md): fuente recibida; se conserva sin reescritura.
- [Aclaraciones vigentes](DECISIONES_VIGENTES.md): un rol ADMIN, múltiples cuentas y prioridad de las decisiones recientes.
- [Alcance vigente](ALCANCE_VIGENTE.md): usuarios, funciones previstas, avance existente, exclusiones y responsables de decisiones abiertas.
- [Registro de horas de desarrollo](REGISTRO_HORAS.md): presupuesto de 300 horas separado de las metas estudiantiles; sesiones reales solo cuando se confirmen.
- [Fuentes y trazabilidad](FUENTES_Y_TRAZABILIDAD.md): procedencia, confirmación de la lógica, sustituciones y bloqueos por issue.
- [Reglas consolidadas](REGLAS_CONSOLIDADAS.md): reglas confirmadas y decisiones pendientes, con referencias y ejemplos sintéticos.
- [Modelo de datos nuevo propuesto](MODELO_DATOS_PROPUESTO.md): diccionario, claves y relaciones para revisión por bloques; no es una migración aprobada.
- [Contexto breve para la IA](CONTEXTO_IA.md): reglas comunes que evitan repetir el plan completo en cada prompt.
- [Prompt para exportar el otro chat](PROMPT_EXPORTAR_CONTEXTO.md): referencia histórica opcional; el propietario autorizó omitir otra exportación en #5.
- [Índice de issues](INDICE_ISSUES.md): orden, hitos, dependencias y enlaces reales del backlog.

## Orden de construcción

PostgreSQL y el esquema inicial ya funcionan en V0.0.1. No se reinicializa el proyecto ni se reconstruye ese avance.

1. Publicar y aceptar la planeación; consolidar fuentes y resolver las decisiones que realmente afecten a cada módulo.
2. Completar pruebas aisladas, contrato HTTP y privilegios de base que aún faltan.
3. Acordar contrato de cuentas ADMIN y crear únicamente la persistencia de identidad/sesiones necesaria. Sobre esa base implementar Argon2id, JWT, CSRF, login, renovación y logout.
4. Proteger el acceso real y el layout; después desarrollar catálogos, instituciones y estudiantes. Estos últimos no necesitan esperar todo el ciclo de proyectos cuando sus dependencias reales ya están resueltas.
5. Proyectos → participaciones y cierres → libro de horas/saldos → reportes y Excel.
6. Pruebas integrales, recuperación y releases aprobados; servidor y producción se ejecutan únicamente en su etapa autorizada.

La autenticación no necesita que se creen de golpe todas las tablas de negocio. El v0.2 no se recuperó y se propone un modelo nuevo. La lógica disponible está confirmada por el propietario; la aprobación de campos y relaciones y las decisiones abiertas bloquean los cambios que realmente las requieren.

## Uso del backlog

Los issues incluyen un prompt acotado, criterios de aceptación, etapa de origen, límites y dependencias nativas de GitHub. Los 10 hitos agrupan áreas de entrega; el número del issue identifica la tarea y genera su rama al aplicar `status:ready`.

Las tareas permanecen en backlog hasta decidir ejecutarlas. Las decisiones y la exportación de contexto no se presentan como aprobadas. La comprobación automática adicional de dependencias se contempla en su propio issue; mientras se incorpora, deben revisarse manualmente antes de marcar una tarea lista.

El trabajo previo se acredita en #1 y la validación manual inicial sigue en #2. No se crean otra vez esas tareas ni se agregan otras 300 horas al presupuesto. Los hitos no fijan fechas ni afirman tiempo ya trabajado.

La versión se rige por [GitFlow](../GITFLOW.md), no por las etiquetas de ejemplo del plan antiguo. `main` sigue reservada con README; esta documentación se prepara mediante PR hacia `develop`.
