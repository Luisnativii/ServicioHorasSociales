# Decisiones de inicialización

- Desarrollo: `hs-dev`, único puerto publicado `127.0.0.1:8080`; Vite escucha en `0.0.0.0:8080` dentro del contenedor y conserva `/api` hacia `http://api:8080`.
- `app` comunica web/API; `data`, interna, comunica API/PostgreSQL/migrator. La web no pertenece a `data`. API y PostgreSQL no publican puertos en el equipo.
- PostgreSQL 18 almacena el clúster en el volumen `hs-dev_pgdata`, montado en `/var/lib/postgresql`. No eliminar el volumen para reparar fallos.
- Goose crea `horas_sociales`; su registro de versiones reside en `public.goose_db_version`. No hay tablas de negocio. La migración fundacional rechaza rollback para evitar marcar como revertido un esquema que se conservó.
- La cuenta local es propietaria durante esta inicialización. Antes de implementar negocio se definirán roles limitados para aplicación y migraciones. La conexión sin TLS permanece dentro de la red privada local de Docker.
- API: `net/http`, chi, pgxpool, máximo configurable 1–50 conexiones (5 por defecto), conexión con timeout de 3 s, comprobación inicial de 5 s, readiness de 2 s y apagado de 10 s. Los logs no imprimen errores de conexión originales ni secretos.
- El healthcheck de la API usa liveness para conservar el proceso durante una caída de PostgreSQL. Readiness consulta la base en cada petición. `depends_on: service_healthy` controla el arranque, no la recuperación en ejecución.
- Desarrollo: fuentes montadas en solo lectura; Go compila en `/tmp` con el UID/GID del usuario; Vite usa el usuario `node`. Los lockfiles se generan con el UID/GID del anfitrión.
- El observador Go compila antes de reemplazar el proceso; si la edición no compila, conserva el binario anterior y muestra el error en logs. Si cambia `go.mod`, regenerar `go.sum` dentro de Go en Docker y reconstruir la imagen.
- Runtime: API y migrator usan Distroless fijado por digest y UID 65532; web usa Nginx sin root, UID 101, puerto 8080. Nginx resuelve el servicio API por DNS de Docker para tolerar cambios de dirección del contenedor.
- Todos los paquetes directos tienen versión exacta; `go.sum` y `package-lock.json` se conservan. No se usan imágenes `latest` ni CDN para Tailwind.
- La pantalla consulta al abrir y al pulsar «Volver a comprobar». No realiza sondeo automático ni inventa estados. Tanto los errores HTTP como los de red se muestran como servicio no disponible.
