# Aclaraciones vigentes de la planeación

Actualización de identidad: 28 de octubre de 2026, America/El_Salvador. Conserva las decisiones anteriores y señala sus sustituciones.

## Identidad y acceso — confirmado por el propietario en este chat

> «va a ser un solo rol de administrador, el cual va a tener un acceso completo, siempre validado por JWT ya que pueden existir multiples cuentas de administrador».

Consecuencias para el backlog:

- Un único rol ADMIN y varias cuentas administradoras.
- Cada cuenta tiene identidad propia y acceso completo al ámbito de la aplicación.
- JWT protege las operaciones; registrar el actor cuando la regla exija trazabilidad. El 28 de octubre el propietario excluyó revocación y recuperación de contraseña por el momento; no implementar sesiones revocables por el texto antiguo del plan.
- No crear una restricción de cuenta única ni rechazar el segundo ADMIN por ese motivo.
- No añadir roles de estudiantes/coordinadores ni acceso público por deducción.
- Inicio de sesión mediante correo electrónico y contraseña.
- Creación de cuentas exclusivamente mediante herramienta de administración ejecutada en Docker; sin formulario ni endpoint web de alta.
- Se permite desactivar la última cuenta ADMIN activa. No imponer un mínimo de una cuenta activa; las altas siguen disponibles por herramienta, sin una función de recuperación/reset de contraseña.
- El propietario aceptó el [contrato de acceso](CONTRATO_ACCESO.md), revisión 0.1, y M10 revisión 0.2 el 28 de octubre: mínimo seis caracteres con número, hash Argon2id, dos horas fijas y una sola tabla de administradores. La aceptación está registrada en #20; no aprueba los demás bloques de negocio.

Las decisiones y aceptación constan en [#20](https://github.com/Luisnativii/ServicioHorasSociales/issues/20). El logout elimina cookies sin invalidación remota; desactivar impide nuevos logins sin revocar JWT emitidos. #22/#27 se cerraron como no planificados y sus dependientes se ajustaron al contrato aceptado, sin afirmar que esas funciones se implementaron. #21 puede ejecutar únicamente el bloque M10, después de integrar sus dependencias.

Esta aclaración sustituye las frases de una sola cuenta ADMIN del plan de programación original. El documento original se conserva como fuente histórica sin reescribirlo.

## Fuentes y confirmación de la lógica

El modelo v0.2 y las entrevistas originales citados en el plan no se recuperaron. El propietario autorizó diseñar y revisar juntos un modelo nuevo; no se intenta reconstruir las supuestas 19 tablas del documento ausente.

El 26 de octubre, al trabajar #5, el propietario confirmó que la lógica de negocio ya se revisó y coincide con el contexto disponible; autorizó omitir la recuperación de `Logica_Negocio.md`. Se consolida en [REGLAS_CONSOLIDADAS.md](REGLAS_CONSOLIDADAS.md), con [procedencia y sustituciones](FUENTES_Y_TRAZABILIDAD.md). No se afirma haber incorporado ni vuelto a leer aquella exportación.

El [modelo nuevo](MODELO_DATOS_PROPUESTO.md) es una propuesta para revisión: campos, tipos, claves, cardinalidades y restricciones todavía no están aprobados. Las decisiones que el contexto mantiene abiertas conservan sus issues. No generar SQL ni desbloquear dependientes por la sola existencia de estos documentos.

Se encontró además el chat local «Organiza ideas de Horas Sociales», que contiene la lluvia de ideas inicial y preguntas; no equivale a la consolidación final de decisiones ni al modelo v0.2.

## Alcance de carga y operación — issue #6

El 26 de octubre el propietario aclaró que introducirá los datos iniciales mediante consultas SQL. El administrador podrá introducir y editar datos por la interfaz conforme a los contratos de cada módulo; la carga inicial no sustituye esas pantallas. Excel exportará los datos de los dashboards con los mismos filtros y cifras. PDF, adjuntos e importador web no forman parte del alcance solicitado.

La carga inicial se concreta en #95, con la procedencia, corte y prevención de duplicados de #10. Las fechas históricas reales y el instante de incorporación se distinguen. Este acuerdo no autoriza ejecutar SQL ni inventar datos históricos en #6. El detalle está en [ALCANCE_VIGENTE.md](ALCANCE_VIGENTE.md); las horas de desarrollo se registran aparte en [REGISTRO_HORAS.md](REGISTRO_HORAS.md).

## Flujo de trabajo y versiones

La configuración existente en `docs/GITFLOW.md` prevalece sobre los ejemplos anteriores del plan: `main` conserva solo README; trabajo e integración en `develop`; rama por issue; PR y checks; tags automáticos `VMAJOR.MINOR.PATCH`.

Las versiones de ejemplo por etapa del plan no se crean por adelantado ni se generan manualmente para cada issue. La promoción de releases, candidatas y GHCR se define en una tarea posterior explícita; los tags actuales de integración no autorizan despliegues.

## Avance que se reutiliza

V0.0.1 y el issue #1 ya incorporan inicialización, Docker, PostgreSQL 18, esquema vacío, salud real, frontend mínimo, migración inicial, persistencia y CI. El issue #2 conserva la validación manual pendiente. No se vuelve a construir la base del proyecto desde cero ni se acredita otra vez el mismo trabajo.
