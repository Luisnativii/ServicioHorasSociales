# Prompt para traer la lógica de negocio del otro chat

Copia el siguiente texto en el chat que contiene las entrevistas, decisiones y modelo de datos. Guarda el resultado como `docs/Planeacion/Logica_Negocio.md` o comparte aquí el archivo Markdown que genere.

```text
Prepara un archivo Markdown llamado Logica_Negocio.md con la especificación funcional completa y vigente del Portal de Horas Sociales, usando las entrevistas, aclaraciones, modelos y decisiones de esta conversación. Servirá como fuente para otra IA que implementará tareas pequeñas; debe poder entender las reglas sin leer el historial del chat.

No inventes requisitos ni conviertas propuestas de la IA en aprobaciones. Distingue en cada regla: CONFIRMADO POR EL USUARIO, PROPUESTA o PENDIENTE. Cuando una corrección reemplace una idea anterior, conserva la decisión vigente y registra la contradicción en un apartado de cambios. Si no puedes recuperar un archivo o dato mencionado, anótalo como faltante.

Aclaración más reciente que debes incorporar: hay UN SOLO ROL ADMIN, pero pueden existir MÚLTIPLES CUENTAS ADMINISTRADORAS; todas tienen acceso completo y deben autenticarse con JWT. Una referencia anterior a una sola cuenta administradora queda sustituida. Los estudiantes, coordinadores y contactos no obtienen acceso por estar registrados.

Organiza el Markdown con estas secciones:
1. Alcance, usuarios, exclusiones y glosario. Aclara si organizaciones e instituciones son el mismo módulo.
2. Reglas confirmadas por módulo: administradores/sesiones; catálogos académicos y territoriales; instituciones/contactos; proyectos; estudiantes; participaciones; horas/saldos/correcciones; dashboard y Excel.
3. Modelo de datos v0.2 y cambios posteriores: todas las entidades existentes, campos y tipos, obligatoriedad, PK/FK, unicidad, restricciones, índices propuestos y cardinalidades. Conserva los nombres acordados. No completes tablas faltantes por intuición. Distingue diseño aprobado de sugerencias técnicas. Añade el diagrama Mermaid solo si corresponde al modelo realmente definido.
4. Estados y transiciones de proyecto y participación: acción, condición previa, efecto, fecha, actor e historia que debe conservarse. Incluye cierre, abandono individual/total, cancelación, finalización y reincorporación.
5. Libro de horas: internas/externas, requisitos de cumplimiento, enteros, fechas, categoría histórica, idempotencia, saldos previos, corte y doble conteo; correcciones/reversos y operaciones posteriores al cierre.
6. Historia académica y de clasificación: efecto de cambiar carrera, facultad, tipo/áreas o naturaleza del proyecto sobre movimientos y reportes anteriores.
7. Reportes y exportación: definición exacta de estudiante único, participación, actividad del período, acumulado y cumplimiento; filtros y reglas para antecedentes desconocidos.
8. Casos de ejemplo con entradas, acciones y resultado esperado, incluidos errores y límites. Usa únicamente datos sintéticos.
9. Decisiones pendientes y documentos/fotos/listados faltantes: pregunta concreta, módulo afectado y qué implementación bloquea.
10. Registro de fuentes, versiones y aclaraciones relevantes; separa lo dicho por el usuario de lo sugerido por la IA.

Dedica especial atención a cardinalidad de tipos/áreas, contactos/coordinadores, cierre de participaciones, saldo inicial sin historia y correcciones después del cierre. Si ya se resolvieron, conserva la respuesta; si no, no la resuelvas por cuenta propia.

Entrega el archivo .md para descargar. Si no puedes generar archivos, devuelve su contenido completo en un único bloque Markdown listo para guardar, sin resumir ni omitir las tablas del diccionario. No incluyas contraseñas, tokens, datos reales de estudiantes ni transcripciones innecesarias del historial. No programes la aplicación ni generes SQL de implementación en este encargo.
```

Al recibirlo, se revisará y aprobará en el issue de consolidación de fuentes. La importación de contexto no cambia automáticamente el esquema ni convierte cada propuesta en un requisito.
