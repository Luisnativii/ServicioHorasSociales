# Contrato de integración con PostgreSQL del core

Propuesta técnica para [#139](https://github.com/Luisnativii/ServicioHorasSociales/issues/139), basada en [el SQL del core en 3f8d6a4](https://github.com/jaalgori0/horas-sociales-core-backend/tree/3f8d6a4169b598674b63b4618ced5e4589e96c27/src/database/migrations) y en el [arranque local del PR #141](https://github.com/Luisnativii/ServicioHorasSociales/pull/141). Las decisiones pendientes de ambos equipos están indicadas en la matriz.

## Esquemas, responsables y acceso

| Objeto | Responsable | Acceso previsto del portal |
| --- | --- | --- |
| `public` y `public.schema_migrations` | Equipo del core. Aplica las migraciones del core y administra el histórico. | Cuenta Go de mínimo privilegio, con permisos por objeto. El registro de migraciones se verifica con el operador. |
| `horas_sociales` | Equipo del portal. Aplica migraciones de sus objetos con Goose y mantiene ADMIN y sus catálogos. | Rol de aplicación para operaciones autorizadas en el esquema. Credenciales de migración separadas de la API. |
| `public.goose_db_version` | Registro de Goose del portal en la prueba local. | Confirmar ubicación y propiedad con el operador antes del primer arranque sobre la base compartida real. Restringirlo al migrador. |
| Instancia y volumen PostgreSQL | Operador del core. | El Compose de la laptop conecta por `DATABASE_URL` a la instancia existente. Instancia y volumen se administran desde el stack del core. |

Confirmar con el operador el host o alias de la red Docker, puerto interno, base, modo TLS y roles. Introducir la conexión como secreto del servicio API. Usar nombres de esquema explícitos en SQL y conservar el acceso web privado de #94. El operador verifica las migraciones y la conexión desde el contenedor de la API antes de habilitar los módulos.

## Matriz de datos

La columna de escritura registra decisiones pendientes. En la prueba actual, el permiso de lectura del core se limita a facultades y carreras.

| Módulo | Contrato observado | Decisión de escritura o compatibilidad |
| --- | --- | --- |
| Facultades y carreras | `public.faculties.id` y `public.careers.id` son UUID. `careers.faculty_id` referencia a la facultad. Ambos tienen estado activo. | Catálogos de solo lectura para el portal, según la decisión del propietario del 16 de noviembre de 2026. El equipo del core administra los valores existentes. La API consulta directamente la base compartida y permite recuperar referencias inactivas para el historial. #34, #35 |
| Territorio | `geo_departments` → `geo_municipalities` → `geo_districts`, con UUID. | Verificar cobertura y fuente. El operador carga faltantes tras aprobar #37. |
| Instituciones y contactos | `institutions.district_id` representa sede. La 002 agrega `codigo_css` único. `project_team_members` pertenece al proyecto. | Acordar alta y desactivación. Resolver la representación de contactos institucionales y coordinadores sin asignarles cuentas de acceso. #44, #46–#48 |
| Proyectos y clasificación | `projects` referencia institución y admite distrito de ejecución. `project_careers` vincula carreras. La 002 identifica históricos con `(institution_id, correlativo_css)`. El ETL usa `atributos.area_css`. | Confirmar cuándo mostrar sede heredada, qué campos edita cada servicio y cómo enlazar tipo/área del portal con la clasificación previa. #41, #50, #51 |
| Estudiantes | `students.id` UUID, carnet textual de ocho dígitos y `career_id` del core. La 002 permite nombres y apellidos nulos juntos y conserva el nombre completo histórico. | Acordar campos editables e historia académica. Mantener los campos de DUI fuera de consultas ordinarias y de los DTO. #55, #56 |
| Participaciones y cierres | `project_enrollments` impone unicidad por estudiante/proyecto y coherencia entre estado terminal y `cerrada_en`. Express bloquea el proyecto al asignar y rechaza cupo lleno. | Acordar un protocolo transaccional compartido y resolver cupo bloqueante frente a aviso, transiciones, retiro y cierre. #53, #59, #60, #62, #64, #81 |
| Acreditación y saldos | `hour_movements` referencia participación, tiene horas positivas, naturaleza interna/externa, fecha efectiva y `solicitud_id` único cuando está presente. `v_student_hours` agrega los movimientos. | Acordar idempotencia y registro de actor antes de escribir. Resolver saldos previos, correcciones y su efecto en agregados sin duplicar la carga histórica. #65–#69, #72 |
| Acceso y auditoría | El portal autentica con su propio ADMIN/JWT. `hour_movements.registrado_por` referencia a `public.users.id`. Los triggers del core consultan `app.actor`. | Acordar correspondencia de identidades y atribución transaccional con el core. Validar también los permisos que necesitan los triggers para cada mutación habilitada. |
| Dashboard y Excel | La vista de horas agrega por estudiante. Las participaciones y los movimientos tienen granularidades distintas. | Acordar filtros y fechas compartidos y cómo reportar historia incompleta. Conciliar contra datos de control del core. #74, #75, #79, #95 |
| Citas, notificaciones y mapa | `citas`, `cita_eventos`, `correos`, `dias_bloqueados` y `map_markers` son objetos del core. | Fuera del alcance de integración inicial del portal. |



1. Ejecutar `bash scripts/check-core-readiness.sh`. La comprobación, incluida en `scripts/ci.sh`, aplica SQL 001–003 y una facultad sintética en `hs-test-core`, con red y volumen separados de desarrollo y de las otras pruebas. Verifica readiness con el catálogo accesible, el esquema ausente y PostgreSQL detenido. También comprueba que la API conserva el catálogo y el registro de migraciones.
2. El Compose de base compartida configura `CHECK_CORE_CATALOG=true`. Readiness consulta `public.careers` y `public.faculties` mediante la conexión de la API. Un catálogo vacío sigue siendo válido. Si la consulta falla, readiness responde 503 y la respuesta pública solo informa el estado.
3. Comprobar en una instancia aislada que iniciar el portal conserva las filas del core y su registro de migraciones. Verificar nombres de esquemas, permisos y aislamiento con la cuenta real de ejecución de la API.
4. Contrastar nombres, versión PostgreSQL y migraciones con la instancia de la laptop en una ventana coordinada con su operador. Registrar el resultado y las decisiones de la matriz antes de otorgar permisos de escritura.
