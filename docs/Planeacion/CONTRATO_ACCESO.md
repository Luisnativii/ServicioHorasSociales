# Contrato de acceso administrativo

Revisión 0.1 · 28 de octubre de 2026 · Issue [#20](https://github.com/Luisnativii/ServicioHorasSociales/issues/20).

**Estado: revisión 0.1 y M10 revisión 0.2 aceptados por el propietario el 28 de octubre de 2026.** Respondió «sí lo apruebo» a la propuesta concreta y autorizó integrar #108; la confirmación consta en [#20](https://github.com/Luisnativii/ServicioHorasSociales/issues/20#issuecomment-6483287093). Los títulos de propuesta conservan la procedencia del texto revisado; no son un bloqueo pendiente. Este contrato no aprueba otros bloques del modelo ni afirma que el login ya esté implementado.

## Decisiones humanas que se conservan

- Varias cuentas, un único rol ADMIN y acceso completo. Estudiantes/coordinadores no son usuarios del portal.
- Login por correo electrónico y contraseña; cuentas creadas únicamente mediante herramienta de administración dentro de Docker, sin alta web ni registro público.
- Contraseña almacenada hasheada, nunca en texto claro. El propietario solicita una contraseña sencilla: «6 dígitos con un número».
- Sesión de al menos dos horas.
- Sin recuperación de contraseña y sin revocación por el momento, según las instrucciones posteriores del 28 de octubre. Sustituyen el plan de sesiones revocables/renovación anterior.
- Se permite desactivar la última cuenta activa. No se exige conservar una cuenta activa ni se habilita recuperación pública.

## Propuesta de comportamiento para aceptar

| Tema | Propuesta de esta revisión |
| --- | --- |
| Contraseña | Mínimo seis caracteres y al menos un dígito ASCII de 0 a 9; sin exigir mayúsculas ni símbolos. Máximo 128 bytes para acotar entradas. No recortar ni normalizar la contraseña. Esta interpretación de «6 dígitos» requiere aceptación. |
| Correo | Recortar espacios exteriores, pasar a minúsculas y guardar un único correo canónico, único. Validar una dirección sin nombre visible; rechazar correo vacío. Aplicar la misma normalización en herramienta/login. |
| Duración | Dos horas exactas desde cada login exitoso, sin renovación automática ni extensión por actividad. Al vencer, volver a iniciar sesión. |
| Dispositivos | Se admiten logins simultáneos; no hay listado de sesiones ni cierre remoto. |
| Logout | Eliminar las cookies del navegador que ejecuta el cierre. Sin invalidación en servidor: una copia del JWT permanece válida hasta su vencimiento. |
| Desactivación | Impide nuevos logins. Los JWT ya emitidos conservan su validez hasta vencer; no añadir una revocación inmediata por desactivación. |
| Recuperación | Sin rutas/pantallas de olvido, reset por correo ni comando de reemplazo de contraseña en esta entrega. El operador conserva la herramienta de creación: si pierde acceso, puede crear otra cuenta, sin sobrescribir la anterior. |
| Gestión de cuentas | Listar, activar y desactivar cuentas; no borrar identidades ni editar correo/contraseña desde la web en esta primera entrega. Permitir cero cuentas activas. Registrar actor y resultado de cambios administrativos sin contraseñas/tokens en logs. |

## Persistencia mínima de identidad propuesta — M10 revisión 0.2

Una sola tabla `horas_sociales.administrador`. No se crea `sesion_admin`, tabla de refresh, catálogo de roles ni tabla de recuperación.

| Campo | Tipo y restricción propuestos |
| --- | --- |
| `id` | UUID, PK, generado con `gen_random_uuid()` al insertar. |
| `correo` | `text NOT NULL UNIQUE`; correo canónico no vacío, con comprobación de `correo = lower(btrim(correo))`. Validación completa en herramienta y API. |
| `contrasena_hash` | `text NOT NULL`, formato PHC de Argon2id; nunca guardar ni devolver la contraseña. |
| `activo` | `boolean NOT NULL DEFAULT true`; sin restricción de cantidad mínima de cuentas activas. |
| `creado_en` | `timestamptz NOT NULL DEFAULT now()`, instante técnico. |
| `actualizado_en` | `timestamptz NOT NULL DEFAULT now()`, actualizado por las operaciones de cambio; no por consultas/login. |

No existe borrado de cuentas en este contrato. El actor de las operaciones que lo requieren se obtiene de `sub` del JWT y puede seguir referenciando una cuenta desactivada. El diseño de auditoría de cada módulo conserva su aprobación propia; este documento no crea una tabla genérica de auditoría ni tablas de negocio.

Hash propuesto: Argon2id v19, memoria 65 536 KiB, tres iteraciones, paralelismo cuatro, salt aleatorio de 16 bytes y resultado de 32 bytes. Guardar algoritmo, parámetros, salt y resultado en la cadena PHC. Verificar en tiempo constante, sin calcular un hash nuevo con salt distinto para comparar. Son parámetros iniciales que deben medirse dentro de Docker al implementar, con concurrencia acotada. Referencia: [RFC 9106](https://www.rfc-editor.org/rfc/rfc9106.html#section-4).

## JWT y cookies propuestos

JWT firmado con HS256 y una clave aleatoria independiente de al menos 32 bytes, generada una sola vez en un archivo protegido bajo `.secrets/`, fuera de Git e imágenes. Conservar la clave entre arranques; no derivarla de una contraseña, no imprimirla y no generar otra silenciosamente. Algoritmo permitido fijo: rechazar `none`, otros algoritmos y firmas inválidas.

Claims obligatorios: `sub` (UUID de cuenta), `iss` (`horas-sociales-api`), `aud` (`horas-sociales-web`), `iat` y `exp`. Validar firma, emisor, audiencia, forma del sujeto y tiempo; exigir que `exp - iat` corresponda al plazo de dos horas y rechazar emisión futura, admitiendo una tolerancia de reloj documentada de 30 s. No incluir contraseña, hash ni datos personales adicionales. No hay `jti` utilizado para revocación ni persistencia de tokens. Las verificaciones siguen las [buenas prácticas JWT de RFC 8725](https://www.rfc-editor.org/rfc/rfc8725.html#section-3).

Cookie `hs_access`: `HttpOnly`, `SameSite=Strict`, `Path=/api`, `Max-Age=7200`, sin `Domain`. Usar `Secure` con HTTPS; la excepción de desarrollo HTTP se limita a `localhost`/`127.0.0.1`. No guardar el JWT en localStorage ni exponerlo en JSON. Cerrar elimina la misma cookie con `Max-Age=0` y los mismos atributos de ruta.

Para mutaciones, incluido login/logout, verificar origen permitido y un token CSRF mediante doble envío firmado (cookie legible `hs_csrf` y cabecera `X-CSRF-Token`). La firma evita aceptar un valor arbitrario elegido por el cliente; para operaciones autenticadas debe ligarse al JWT vigente. Su validación y comparación se concretan y prueban en #25, sin tabla de sesiones. No habilitar CORS genérico con credenciales ni aceptar un origen exterior por defecto.

## Contrato HTTP propuesto

El formato de error común lo entrega #17; aún no existen estas rutas. No exponer tokens ni hashes en respuestas. Las respuestas de autenticación llevan `Cache-Control: no-store`.

| Operación | Contrato previsto |
| --- | --- |
| `GET /api/v1/auth/csrf` | Sin sesión requerida. Preparar cookie CSRF firmada para login; 204 sin datos de cuenta. |
| `POST /api/v1/auth/login` | Objeto JSON `{correo, contrasena}`, origen/token CSRF válidos. 200 con `{id, correo}` y cookies al autenticar; 400 entrada inválida; 401 genérico tanto para cuenta inexistente como contraseña incorrecta o cuenta inactiva; 403 fallo CSRF/origen; 429 si excede el límite de intentos. |
| `GET /api/v1/auth/me` | JWT válido: 200 con `{id, correo}`. Sin token, alterado o vencido: 401. Consultar identidad no extiende vencimiento; una cuenta desactivada con JWT vigente conserva el acceso hasta expirar, según la decisión sin revocación. |
| `POST /api/v1/auth/logout` | Origen/token CSRF válidos: 204; borrar cookies incluso si ya venció el JWT. No consultar ni modificar un registro de sesiones inexistente. |
| Rutas de refresh/reset/revocación | No se implementan en esta entrega; una ampliación requiere nueva decisión. |

Propuesta de control de intentos para #26: cinco fallos por IP en una ventana de cinco minutos, memoria acotada y límite de concurrencia para hashes; devolver 429 con `Retry-After`. Usar la IP de la conexión/proxy local conocido; no confiar en `X-Forwarded-For` arbitrario. No bloquear ni modificar permanentemente una cuenta por intentos fallidos. Precisar el comportamiento con varias instancias solo si se aprueba ese despliegue.

Los handlers y middleware usan el contexto y los plazos existentes, errores genéricos y SQL parametrizado. El rol ADMIN es fijo, sin aceptar roles o IDs de actor enviados en el cuerpo. No hay endpoints públicos de negocio.

## Herramienta de cuentas propuesta

Comando de creación ejecutado dentro de Docker, con correo validado y contraseña leída de entrada interactiva protegida; no pasar la contraseña como argumento de línea ni guardarla en logs/historial. La primera y las siguientes cuentas siguen la misma operación. Un correo existente produce conflicto y no cambia su hash ni su estado. No hay cuenta o contraseña de ejemplo creada automáticamente.

Crear una cuenta cuando no hay ninguna activa sigue permitido. No se incorpora un comando de reset o reactivación a la herramienta por deducción; la gestión web de estado pertenece a #29 y requiere JWT/CSRF vigentes. Los errores no muestran la contraseña/hash ni la conexión de base.

## Ajuste de issues propuesto, tras aprobación

Las dependencias nativas actuales no se modifican ni se cierran como si ya estuvieran entregadas. Antes de marcar #21 o sus dependientes listos, aceptar esta revisión e integrar sus prerrequisitos.

| Issue | Ajuste necesario por las nuevas decisiones |
| --- | --- |
| #21 | Migrar únicamente `administrador` conforme al bloque M10 aceptado; consultas sqlc. Sigue dependiendo de #18 integrado y de #20 aceptado/integrado. |
| #22 | Sesiones revocables fuera del alcance actual: cerrar como no planificada tras aceptar la exclusión; no crear su tabla. |
| #23 | Herramienta de creación; retirar recuperación/reset y dependencia de #22. |
| #24 | Middleware JWT con validación estricta; retirar consulta a sesión/revocación y dependencia de #22. |
| #25 | Cookies/origen/CSRF sin sesiones persistidas; depende del contrato y helpers HTTP. |
| #26 | Login y me con Argon2id y control de intentos. |
| #27 | Refresh/rotación fuera del alcance actual: cerrar como no planificada tras aceptar la exclusión. |
| #28 | Logout de cookies, sin revocación; sustituir dependencia de #27 por #26/#25. |
| #29/#33 | Listado y activación/desactivación; sin alta/reset web. |
| #30/#31/#32 | Cliente de sesión/login/layout sin renovación concurrente; manejar vencimiento de dos horas y cierre de cookies. Ajustar dependencia de #27. |
| #81 | Conservar pruebas de concurrencia de horas/cierres; retirar escenarios de refresh/reutilización excluidos y su dependencia de #27. |

Las exclusiones deben registrarse también en las dependencias de cualquier otro issue que las cite; no desbloquear implementaciones marcando una exclusión como funcionalidad construida. El modelo M10 revisión 0.1 y los prompts del plan quedan como antecedentes; esta revisión los sustituirá para identidad una vez aceptada.

## Casos de aceptación para revisar antes de migrar

1. Crear dos cuentas por herramienta: ambas acceden con el mismo rol y distinto identificador.
2. Contraseña de cinco caracteres o sin número: rechazada; seis caracteres con un número: admitida sin exigir símbolos/mayúsculas.
3. Correo con espacios exteriores/mayúsculas: se normaliza; repetir el correo produce conflicto y conserva la cuenta original.
4. Contraseña incorrecta, cuenta ausente o inactiva: mismo 401 genérico; no se envían ni registran secretos.
5. JWT auténtico emitido hace una hora: válido; vencido hace más de la tolerancia: 401. Firma/algoritmo/emisor/audiencia alterados: 401.
6. Cerrar sesión: navegador pierde la cookie; una copia previa del JWT sigue válida hasta expirar. No prometer cierre remoto.
7. Desactivar la última cuenta: permitido; sus tokens previos no se revocan. No hay nuevo login con esa cuenta; el operador puede crear otra mediante herramienta.
8. Mutación sin CSRF o desde origen no permitido: 403 sin cambios de datos.
9. Reiniciar contenedores con la misma clave: no invalida JWT vigentes ni cambia hashes. No se crean cuentas/secretos de ejemplo.
10. Hash persistido es PHC Argon2id y se comprueba dentro de Docker; no hay columna de contraseña en claro ni tabla de sesiones.

**Aprobación registrada:** se aceptaron la interpretación de seis caracteres con número, el plazo fijo de dos horas, los efectos de logout/desactivación sin revocación y M10 revisión 0.2. La implementación de identidad corresponde a #21; no se crean sesiones/refresh/reset excluidos. Los demás módulos conservan sus aprobaciones propias.
