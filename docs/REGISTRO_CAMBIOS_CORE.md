# Registro de cambios del core pendientes de coordinar

Este registro conserva los cambios preparados en el core local para compartirlos con su equipo antes de aplicarlos en la base compartida. Cada entrada identifica el commit, la prueba y la coordinación pendiente.

## 23 de noviembre de 2026 · Extensión del contacto institucional

- **Solicitud del portal** [#48](https://github.com/Luisnativii/ServicioHorasSociales/issues/48). El contrato de contactos requiere conservar la extensión por separado y como texto. La migración 008 del core ya aporta `contacto_nombre`, `contacto_email` y `telefono` en `public.institutions`.
- **Entrega local** Rama `codex/48-extension-contacto-institucional`, commit `8f6ccd8`, sobre la revisión `82179f8` del core. El archivo `src/database/migrations/015_extension_contacto_institucional.sql` agrega `contacto_extension text` nullable y exige un teléfono cuando hay una extensión con contenido. Las filas anteriores quedan con extensión nula y conservan los demás datos.
- **Validación** PostgreSQL 16 en un contenedor Docker desechable. Se aplicaron las migraciones 001–014, se insertó una institución sintética y se aplicó la 015 mediante `psql` con transacción y `ON_ERROR_STOP=1`. Se comprobó la conservación de los datos anteriores, la lectura exacta de `004`, el rechazo de una extensión en blanco o sin teléfono y la limpieza conjunta de teléfono y extensión.
- **Pendiente con el equipo del core** Revisar el commit y confirmar la numeración disponible antes de integrar la migración. Acordar cómo las mutaciones del core que limpian `telefono` limpiarán también `contacto_extension` en la misma operación. Coordinar la aplicación mediante el migrador del core y los permisos por columna del rol Go para leer y escribir los campos del contacto, según los endpoints que se habiliten. Verificar estos puntos antes de activar la escritura del portal en la base compartida.
- **Estado** Preparado y probado localmente. La API Go usa las columnas de 008 y 015. El portal tiene un fixture que reproduce únicamente el bloque institucional para pruebas y desarrollo local. La migración sigue pendiente en la base compartida.

## 23 de noviembre de 2026 · Expedientes y vigencias académicas

- **Solicitud del portal** [#55](https://github.com/Luisnativii/ServicioHorasSociales/issues/55). El expediente usa `public.students` y deriva la facultad del catálogo. El portal escribe carnet, nombres, apellidos, correo institucional y carrera. El rol Go no requiere leer ni escribir DUI ni modificar el historial directamente.
- **Contrato para coordinar** [SQL de ensayo](../deploy/core-local/estudiantes-vigencias.sql). Un trigger compartido en `public.students` registra el alta o cambio de carrera de cualquier escritor en la misma transacción, conserva la facultad de la vigencia y cierra la anterior en el instante de inicio de la nueva. En expedientes preexistentes sin historial, el inicio de la vigencia anterior es desconocido. No se reconstruyen cambios históricos sin fuente.
- **Activación** El operador del core debe portar el SQL a su migrador, acordar los privilegios y validar la atribución de actor y motivo desde sus propios escritores. El portal exige el trigger habilitado y permiso de lectura del historial para dar altas o cambiar de carrera. La lectura y la edición de otros campos conservan su propio contrato de permisos.

## 23 de noviembre de 2026 · Proyectos y clasificación

- **Solicitud** [#50](https://github.com/Luisnativii/ServicioHorasSociales/issues/50). El portal lee UUID, correlativo textual por institución, área CSS, ubicación heredada y estado del core. El distrito de ejecución se selecciona aparte de la sede institucional.
- **Persistencia** Los catálogos de tipo, área y modalidad y las vigencias de clasificación se guardan en el esquema del portal. El ETL conserva `atributos.area_css` como fuente histórica. Las migraciones 001–014 no contienen un catálogo de equivalencias entre esa etiqueta literal y el catálogo del portal. El entorno local inicial tampoco contenía sus valores históricos. Por ello se muestran las etiquetas literales cuando existan y se conservan los históricos sin equivalencia. La migración no carga valores de negocio.
- **Activación local** `PROJECTS_WRITES_ENABLED=true` en el Compose local habilita altas y edición; el fixture otorga permisos acotados a la cuenta del portal. Habilitar escrituras en la base compartida requiere validar con el operador los permisos y la convivencia con las reglas del core para naturaleza y estados. Las transiciones terminales pertenecen a #53.

## 23 de noviembre de 2026 · Asignación a proyectos

- **Solicitud** [#59](https://github.com/Luisnativii/ServicioHorasSociales/issues/59). Las dos vistas leen `public.project_enrollments`. Las altas conservan el UUID y los valores iniciales del core. Postgres impide duplicar el par estudiante/proyecto.
- **Concurrencia** Se bloquea primero la fila de `projects`, se comprueba que siga activa y se cuentan ocupaciones en una consulta posterior al bloqueo. El conteo incluye `ASIGNADA` y `EN_CURSO`. #53 debe usar el mismo bloqueo antes de una transición.
- **Activación local** `ENROLLMENTS_WRITES_ENABLED=true` habilita el flujo con aviso y confirmación de exceso. El core Node rechaza el exceso. Antes de activar escrituras compartidas, su operador debe unificar la política y verificar permisos y bloqueos de los escritores. La cuenta local recibe lectura e inserción de las columnas necesarias, sin UPDATE ni DELETE de participaciones.
