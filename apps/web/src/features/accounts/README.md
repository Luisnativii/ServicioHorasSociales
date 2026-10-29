# Cuentas administradoras

Ruta privada `/administradores`, registrada en el menú Archivo y protegida por SessionGuard. Consume la API real de #29: GET `/api/v1/admin/cuentas` y POST `/{id}/activar` o `/{id}/desactivar`. No hay altas, eliminación, edición de correo, cambio de contraseña ni selector de roles.

La paginación y el orden de creación son del servidor; página/tamaño/orden se conservan en la URL. Las consultas usan el signal de TanStack y el cliente central; no se descarga el catálogo completo para paginar localmente. La respuesta se valida y se proyecta a `{id, correo, activo}` sin conservar hashes ni otros campos en la caché de cuentas.

El diálogo muestra el efecto real de activar/desactivar, avisa al operar sobre la propia cuenta y no impone un mínimo de cuentas activas. Desactivar no cierra las sesiones ya emitidas: solo bloquea nuevos logins. El texto aclara que la creación de otra cuenta cuando todas estén inactivas corresponde a la herramienta de administración, sin construir un formulario web para ello.

No hay actualización optimista: cambia la fila después de confirmar la respuesta del backend y vuelve a consultar el listado. No se reintentan escrituras. Ante un resultado incierto (red, fallo de servidor o respuesta no verificable), la confirmación se sustituye por actualizar el listado; esa acción solo hace GET y permite revisar el estado antes de decidir otro cambio. Errores controlados conservan el diálogo y la selección. Los diálogos nativos restauran el foco al botón invocador, o al encabezado si la fila dejó de existir.

Pruebas del cliente de cuentas incluidas en `npm run test:session`, dentro de Docker. La comprobación visual de los cambios reales utiliza exclusivamente cuentas sintéticas en hs-test y las retira al terminar; no crear cuentas ni desactivar usuarios reales como prueba.
