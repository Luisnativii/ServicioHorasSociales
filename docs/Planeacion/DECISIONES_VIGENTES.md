# Aclaraciones vigentes de la planeación

Fecha: 25 de octubre de 2026, America/El_Salvador.

## Identidad y acceso — confirmado por el propietario en este chat

> «va a ser un solo rol de administrador, el cual va a tener un acceso completo, siempre validado por JWT ya que pueden existir multiples cuentas de administrador».

Consecuencias para el backlog:

- Un único rol ADMIN y varias cuentas administradoras.
- Cada cuenta tiene identidad propia y acceso completo al ámbito de la aplicación.
- JWT y sesiones revocables protegen las operaciones; registrar el actor cuando la regla exija trazabilidad.
- No crear una restricción de cuenta única ni rechazar el segundo ADMIN por ese motivo.
- No añadir roles de estudiantes/coordinadores ni acceso público por deducción.
- Alta inicial, gestión posterior, identificador de login y recuperación requieren el contrato técnico y funcional previsto en su issue, sin inventar políticas del último administrador.

Esta aclaración sustituye las frases de una sola cuenta ADMIN del plan de programación original. El documento original se conserva como fuente histórica sin reescribirlo.

## Fuentes faltantes

El propietario indicó que el modelo v0.2 y el contexto funcional completo están en otro chat y que compartirá una exportación Markdown. Todavía no se ha incorporado una especificación completa aprobada en este repositorio.

Las reglas que el plan declara confirmadas se usan para estructurar el backlog, pero no autorizan a inventar el diccionario ausente. Las cardinalidades, columnas y decisiones pendientes deben contrastarse con la fuente completa antes de sus migraciones. Las tareas afectadas tienen dependencias y quedan en backlog.

Se encontró además el chat local «Organiza ideas de Horas Sociales», que contiene la lluvia de ideas inicial y preguntas; no equivale a la consolidación final de decisiones ni al modelo v0.2.

## Flujo de trabajo y versiones

La configuración existente en `docs/GITFLOW.md` prevalece sobre los ejemplos anteriores del plan: `main` conserva solo README; trabajo e integración en `develop`; rama por issue; PR y checks; tags automáticos `VMAJOR.MINOR.PATCH`.

Las versiones de ejemplo por etapa del plan no se crean por adelantado ni se generan manualmente para cada issue. La promoción de releases, candidatas y GHCR se define en una tarea posterior explícita; los tags actuales de integración no autorizan despliegues.

## Avance que se reutiliza

V0.0.1 y el issue #1 ya incorporan inicialización, Docker, PostgreSQL 18, esquema vacío, salud real, frontend mínimo, migración inicial, persistencia y CI. El issue #2 conserva la validación manual pendiente. No se vuelve a construir la base del proyecto desde cero ni se acredita otra vez el mismo trabajo.
