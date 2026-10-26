# Planeación del Portal de Horas Sociales

Este directorio conserva el plan original y organiza su ejecución en tareas pequeñas para IA. Solo se implementa un issue por sesión y después de comprobar sus dependencias.

## Documentos

- [Plan original de 20 etapas](Plan_Programacion_IA_Horas_Sociales.md): fuente recibida; se conserva sin reescritura.
- [Aclaraciones vigentes](DECISIONES_VIGENTES.md): un rol ADMIN, múltiples cuentas y prioridad de las decisiones recientes.
- [Contexto breve para la IA](CONTEXTO_IA.md): reglas comunes que evitan repetir el plan completo en cada prompt.
- [Prompt para exportar el otro chat](PROMPT_EXPORTAR_CONTEXTO.md): permite traer entrevistas, reglas y modelo sin convertir propuestas en requisitos.
- [Índice de issues](INDICE_ISSUES.md): orden, hitos, dependencias y enlaces reales del backlog.

## Orden de construcción

PostgreSQL y el esquema inicial ya funcionan en V0.0.1. No se reinicializa el proyecto ni se reconstruye ese avance.

1. Publicar y aceptar la planeación; consolidar fuentes y resolver las decisiones que realmente afecten a cada módulo.
2. Completar pruebas aisladas, contrato HTTP y privilegios de base que aún faltan.
3. Acordar contrato de cuentas ADMIN y crear únicamente la persistencia de identidad/sesiones necesaria. Sobre esa base implementar Argon2id, JWT, CSRF, login, renovación y logout.
4. Proteger el acceso real y el layout; después desarrollar catálogos, instituciones y estudiantes. Estos últimos no necesitan esperar todo el ciclo de proyectos cuando sus dependencias reales ya están resueltas.
5. Proyectos → participaciones y cierres → libro de horas/saldos → reportes y Excel.
6. Pruebas integrales, recuperación y releases aprobados; servidor y producción se ejecutan únicamente en su etapa autorizada.

La autenticación no necesita que se creen de golpe todas las tablas de negocio. El modelo completo pendiente bloquea solamente los cambios que realmente lo requieren.

## Uso del backlog

Los issues incluyen un prompt acotado, criterios de aceptación, etapa de origen, límites y dependencias nativas de GitHub. Los 10 hitos agrupan áreas de entrega; el número del issue identifica la tarea y genera su rama al aplicar `status:ready`.

Las tareas permanecen en backlog hasta decidir ejecutarlas. Las decisiones y la exportación de contexto no se presentan como aprobadas. La comprobación automática adicional de dependencias se contempla en su propio issue; mientras se incorpora, deben revisarse manualmente antes de marcar una tarea lista.

El trabajo previo se acredita en #1 y la validación manual inicial sigue en #2. No se crean otra vez esas tareas ni se agregan otras 300 horas al presupuesto. Los hitos no fijan fechas ni afirman tiempo ya trabajado.

La versión se rige por [GitFlow](../GITFLOW.md), no por las etiquetas de ejemplo del plan antiguo. `main` sigue reservada con README; esta documentación se prepara mediante PR hacia `develop`.
