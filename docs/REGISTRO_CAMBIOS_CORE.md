# Registro de cambios del core pendientes de coordinar

Este registro conserva los cambios preparados en el core local para compartirlos con su equipo antes de aplicarlos en la base compartida. Cada entrada identifica el commit, la prueba y la coordinación pendiente.

## 23 de noviembre de 2026 · Extensión del contacto institucional

- **Solicitud del portal** [#48](https://github.com/Luisnativii/ServicioHorasSociales/issues/48). El contrato de contactos requiere conservar la extensión por separado y como texto. La migración 008 del core ya aporta `contacto_nombre`, `contacto_email` y `telefono` en `public.institutions`.
- **Entrega local** Rama `codex/48-extension-contacto-institucional`, commit `8f6ccd8`, sobre la revisión `82179f8` del core. El archivo `src/database/migrations/015_extension_contacto_institucional.sql` agrega `contacto_extension text` nullable y exige un teléfono cuando hay una extensión con contenido. Las filas anteriores quedan con extensión nula y conservan los demás datos.
- **Validación** PostgreSQL 16 en un contenedor Docker desechable. Se aplicaron las migraciones 001–014, se insertó una institución sintética y se aplicó la 015 mediante `psql` con transacción y `ON_ERROR_STOP=1`. Se comprobó la conservación de los datos anteriores, la lectura exacta de `004`, el rechazo de una extensión en blanco o sin teléfono y la limpieza conjunta de teléfono y extensión.
- **Pendiente con el equipo del core** Revisar el commit y confirmar la numeración disponible antes de integrar la migración. Acordar cómo las mutaciones del core que limpian `telefono` limpiarán también `contacto_extension` en la misma operación. Coordinar la aplicación mediante el migrador del core y los permisos por columna del rol Go para leer y escribir los campos del contacto, según los endpoints que se habiliten. Verificar estos puntos antes de activar la escritura del portal en la base compartida.
- **Estado** Preparado y probado localmente. La migración sigue pendiente en la base compartida y la API Go de #48 sigue pendiente de implementación.
