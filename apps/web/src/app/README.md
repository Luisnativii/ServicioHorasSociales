# Sesión y transporte

`main.tsx` instala BrowserRouter, QueryClientProvider y SessionProvider. La pantalla de salud sigue disponible; #31 implementa acceso y #32 el layout Archivo. No hay pantallas de negocio simuladas.

`session.tsx` expone `useSession()` con estado `loading`, `anonymous`, `authenticated` o `error`, identidad `{id, correo}` y acciones `login`, `logout`, `validate`. Login/logout esperan al backend real y no reintentan; rechazan una segunda operación de acceso simultánea. Los componentes de #31 deben capturar los errores de las acciones y conservar entradas conforme a sus contratos.

El transporte `lib/api.ts` solo permite rutas relativas `/api/`, envía cookies same-origin, impide caché y redirecciones, y prepara/lee la cookie CSRF en mutaciones. No lee el JWT, no lo guarda y no añade Authorization. La cookie CSRF se vuelve a leer en cada escritura para respetar su rotación tras login. Login y logout preparan CSRF explícitamente; las preparaciones concurrentes se comparten y se serializan al cambiar la identidad.

Las consultas TanStack deben pasar su `signal` al transporte. Queries y mutaciones tienen retry desactivado. Un 401 autenticado borra identidad y cachés; los demás 401 simultáneos o respuestas tardías de la sesión anterior se descartan mediante generaciones. Los 401 de credenciales incorrectas no se confunden con errores de red. Las mutaciones inciertas nunca se reenvían automáticamente: la interfaz informa el fallo para que el usuario pueda comprobar el resultado antes de decidir.

El servidor conserva la autoridad sobre el vencimiento fijo de dos horas. Validar `/auth/me` no prolonga el JWT. El proveedor comprueba identidad al montar, al recuperar foco/visibilidad y cada minuto mientras la página está visible; una llamada privada que devuelva 401 también elimina el estado inmediatamente. No se inventa una fecha de expiración tras recargar, porque la API no devuelve `exp`; no hay refresh ni contador que autorice acceso por sí mismo. Una comprobación con fallo de conexión muestra `error`, no una sesión anónima ficticia.

Logout limpia memoria/cachés aunque la respuesta sea incierta; si falla la red, la acción rechaza y no debe anunciar éxito. El servidor solo borra cookies y no revoca copias del JWT. La próxima validación vuelve a comprobar la identidad real.

`npm run test:session` compila los módulos bajo `/tmp` y ejecuta pruebas con el runner incluido en Node, sin otra dependencia. Ejecutar dentro de Docker; CI las incorpora mediante `scripts/check.sh`. Cubren concurrencia de 401/CSRF, rotación, respuestas tardías, fallos de red y escrituras sin reintento. No persisten credenciales ni fixtures en la base.
