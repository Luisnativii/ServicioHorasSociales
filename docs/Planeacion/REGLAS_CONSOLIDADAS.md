# Reglas de negocio consolidadas

Revisión 1 · 26 de octubre de 2026 · Issue [#5](https://github.com/Luisnativii/ServicioHorasSociales/issues/5).

Estado: contexto funcional confirmado por el propietario; consolidación documental para su revisión. Los identificadores F corresponden al [inventario de fuentes](FUENTES_Y_TRAZABILIDAD.md). El [diccionario propuesto](MODELO_DATOS_PROPUESTO.md) tiene aprobación independiente.

## Identidad y acceso

| ID | Estado | Regla | Fuente |
| --- | --- | --- | --- |
| R01 | CONFIRMADO | Un único rol ADMIN, varias cuentas con acceso completo. Estudiantes, coordinadores y contactos no obtienen una cuenta por estar registrados. | F02; F01 «Acceso y sesiones» con la sustitución vigente; F03. |
| R02 | CONFIRMADO | Inicio de sesión mediante correo electrónico y contraseña. | F02. |
| R03 | CONFIRMADO | Creación de cuentas solo mediante herramienta de administración ejecutada en Docker. Sin alta de cuentas mediante formulario ni endpoint web; sin registro público. | F02. |
| R04 | CONFIRMADO | Se puede desactivar la última cuenta ADMIN activa. El acceso se recupera mediante la herramienta de administración. | F02. |
| R05 | CONFIRMADO | Operaciones privadas protegidas por JWT y sesiones revocables. No hay catálogo público ni acceso Google/Gmail en este alcance. | F02; F01 «Acceso y sesiones»; F03. |
| R06 | PENDIENTE | Acordar normalización/unicidad del correo, política de contraseñas, duración, rotación, revocación, recuperación y operaciones de cuentas del panel. Las duraciones 15 minutos / 8 horas son una propuesta del plan, no una aprobación. | F02; #20, antes de #21–#33. |

## Catálogos, instituciones y proyectos

| ID | Estado | Regla | Fuente |
| --- | --- | --- | --- |
| R07 | CONFIRMADO | Carrera vinculada con facultad; catálogos académicos y territoriales para selección y filtros. Una opción usada se conserva en el historial y su desactivación impide ofrecerla en nuevas altas. | F01 etapas 07 y 10; F03. |
| R08 | CONFIRMADO | Jerarquía territorial departamento → municipio → distrito sobre datos verificados. No inventar nombres, códigos ni asociaciones. | F01 etapa 07; F03; carga en #37. |
| R09 | CONFIRMADO | Una institución puede tener varios proyectos; cada proyecto corresponde a una institución y una ubicación física de ejecución. La sede institucional se distingue de la ubicación del proyecto. | F01 reglas funcionales y etapas 08–09; F03. |
| R10 | CONFIRMADO | Contacto de institución y responsable del proyecto son independientes. Teléfonos y extensiones se conservan como texto. | F01 etapa 08; F03. |
| R11 | PENDIENTE | Definir campos obligatorios y cardinalidad de contactos/coordinadores, vínculos con carreras y proyectos y si se reutiliza una misma persona. | F01 decisiones abiertas; F02 D07–D09; #11. |
| R12 | CONFIRMADO | Un proyecto inicia ACTIVO y EXTERNA por defecto. La fecha final no se exige en el alta y se captura al finalizar. No hay acción de publicar. Desde ACTIVO puede pasar a ABANDONADO, CANCELADO o FINALIZADO, sin reapertura. | F01 etapa 09; F03. |
| R13 | CONFIRMADO | Un cambio de naturaleza afecta las horas nuevas; conserva la categoría de los movimientos anteriores. | F01 reglas funcionales y etapa 12; F03. |
| R14 | PENDIENTE | Elegir un tipo o varios, una o varias áreas, su independencia y su historia. No seleccionar una alternativa SQL antes de aprobar #7 y, para historia, #12. | F01 decisiones abiertas; F02 D04–D05, D20. |
| R15 | PENDIENTE | Definir o excluir modalidad y habilidades blandas antes de crear campos o catálogos. | F01 decisiones abiertas; F02 D06; #13. |

## Estudiantes y participaciones

| ID | Estado | Regla | Fuente |
| --- | --- | --- | --- |
| R16 | CONFIRMADO | Expediente único por carnet, conservado como texto con sus ceros iniciales; nombres, apellidos y relación con carrera. La facultad actual se obtiene de la carrera. No guardar un total de horas editable en el estudiante. | F01 etapa 10; F03. |
| R17 | PENDIENTE | Confirmar obligatoriedad y validación exacta de campos del estudiante, incluyendo género y datos de contacto; decidir historia cuando cambie carrera. No inferir longitud de carnet a partir de un ejemplo. | F02 D21 y D19; #5, #6 y #12. |
| R18 | CONFIRMADO | Un estudiante puede participar en varios proyectos y un proyecto tener varios estudiantes. Una sola participación por par estudiante/proyecto; sin reincorporación después de abandono. | F01 etapa 11; F03. |
| R19 | CONFIRMADO | Cupo solicitado informativo: excederlo advierte y permite asignar. Asignación solo con proyecto ACTIVO. | F01 etapa 11; F03. |
| R20 | CONFIRMADO | Abandono individual conserva participación, horas e historial. El proyecto pasa a ABANDONADO cuando tiene al menos una participación y todas están ABANDONADAS. Tener cero participantes activos no demuestra esa condición. | F01 etapa 11; F03. |
| R21 | PENDIENTE | Precisar estados y efectos de cancelar/finalizar sobre cada participante, combinación de finalizados y abandonados, fechas y campos editables después del cierre. | F01 decisiones abiertas; F02 D10–D15; #8. |

## Horas, saldos y reportes

| ID | Estado | Regla | Fuente |
| --- | --- | --- | --- |
| R22 | CONFIRMADO | Horas enteras agregadas mediante movimientos; acreditación ordinaria positiva. `fecha_horas` es informada por el administrador y `registrado_en` conserva el instante de captura. No sustituir el libro por un saldo editable. | F01 reglas funcionales y etapa 12; F03. |
| R23 | CONFIRMADO | Acreditar solo con proyecto y participación ACTIVOS; no admitir horas tardías después del cierre. Estado, historial y acreditación deben conservar consistencia ante operaciones concurrentes. | F01 etapas 09, 11 y 12; F03. |
| R24 | CONFIRMADO | Reintentar la misma solicitud no duplica movimientos. Reutilizar su clave con un cuerpo distinto produce conflicto. | F01 etapa 12; F03. |
| R25 | CONFIRMADO | Meta de 300 horas internas Y 300 externas por estudiante, sin compensar categorías. El exceso de internas no cubre la falta de externas. Se admiten saldos previos separados por categoría, con procedencia por acordar. | F01 reglas funcionales y etapa 12; F03. |
| R26 | PENDIENTE | Acordar corte, idempotencia y procedencia de saldos previos; no fabricar proyectos, fechas o territorios históricos desconocidos. | F01 decisiones abiertas; F02 D17–D18; #10. |
| R27 | PENDIENTE | Aprobar mecanismo de corrección, reversos y casos después del cierre. Hasta resolverlo no exponer edición, borrado ni reversos de horas. | F01 decisiones abiertas y etapa 12; F02 D16; #9. |
| R28 | PENDIENTE | Acordar límites de `fecha_horas`, relación con fechas del proyecto y zona horaria de cortes. Distinguir fecha de actividad de instante de registro no resuelve esas políticas. | F02 D22; #5 y #8. |
| R29 | CONFIRMADO | Dashboard y XLSX comparten filtros y cifras. Separar estudiantes únicos, participaciones, proyectos y movimientos; evitar multiplicar horas al combinar áreas. La actividad del período se mide por `fecha_horas`; el acumulado y los saldos sin historia se distinguen de esa actividad. | F01 etapas 13–14; F03. |
| R30 | PENDIENTE | Precisar clasificación e historia académica para reportes, significado de cumplimiento a una fecha y semántica final de filtros. | F01 decisiones abiertas y etapa 13; F02 D19, D20, D24 y D25; #7, #12 y #75. |
| R31 | CONFIRMADO | En XLSX, carnet como texto y horas como números. El contenido ingresado por usuarios se exporta como valores, sin convertirlo en fórmulas. | F01 etapa 14; F03. |

## Alcance y procedencia

| ID | Estado | Regla | Fuente |
| --- | --- | --- | --- |
| R32 | CONFIRMADO | Modelo nuevo para revisión; no recuperar por suposición el v0.2 ni exigir 19 tablas. | F02 y F03. |
| R33 | CONFIRMADO | La base Docker, esquema vacío y CI existentes se reutilizan. Las 300 horas estimadas de desarrollo no son horas ya trabajadas ni las metas estudiantiles. | F01 etapa 00, F04 y V0.0.2. |
| R34 | PENDIENTE | Ratificar exclusiones y cualquier ampliación (PDF, adjuntos, importación o nuevas funciones) en #6; no incluirlas por deducción en el modelo. | F02 D03, D23 y D26–D28. |

## Ejemplos de revisión derivados del contexto confirmado

Estos casos usan datos sintéticos; verifican la interpretación de reglas y no afirman que la aplicación ya las implemente.

| Caso | Entrada/acción | Resultado esperado y trazabilidad |
| --- | --- | --- |
| E01 | Dos cuentas ADMIN creadas por herramienta | Ambas pueden autenticarse por correo/contraseña con el mismo ámbito de acceso. R01–R03. |
| E02 | Desactivar la única cuenta activa | La condición de última cuenta no lo impide; recuperación por herramienta. El efecto exacto sobre sesiones depende de R06. R04. |
| E03 | Carnet sintético `00200022` | Se conserva como texto; otro expediente con el mismo carnet se rechaza. La longitud del ejemplo no se vuelve una restricción. R16–R17. |
| E04 | Proyecto solicita cinco alumnos; se asigna el sexto | Advertencia y asignación permitida si cumple las demás reglas. R19. |
| E05 | Proyecto sin participantes / proyecto con todos abandonados | El primero no se abandona automáticamente; el segundo sí. No equiparar cero activos con todos abandonados. R20. |
| E06 | 350 internas y 250 externas | No cumple; faltan 50 externas. Pendientes por categoría: `max(300 - acumulado, 0)`. R25. |
| E07 | 40 externas; cambiar naturaleza; agregar 10 internas | Conserva 40 externas y agrega 10 internas. R13 y R22. |
| E08 | Reintento idéntico / misma clave con otra cantidad | Un movimiento / conflicto, respectivamente. R24. |
| E09 | Acreditación ordinaria después de cierre | Rechazada. Una eventual corrección tiene decisión separada y no habilita horas tardías. R23 y R27. |
| E10 | Saldo previo sin fecha conocida | Puede afectar acumulado según contrato de #10; no se atribuye a un período o territorio inventado. R26 y R29. |

## Límite de aprobación

La ratificación del contexto no elige las alternativas marcadas PENDIENTE. Para cada decisión nueva se registrarán fecha, regla afectada y confirmación del propietario en su issue. El modelo documental puede revisarse por bloques; ninguna tabla queda aprobada solo por aparecer en el diagrama.
