# Inventario de equipo y contrato de calidad

Revisión 0.1 para aprobación · Issue [#14](https://github.com/Luisnativii/ServicioHorasSociales/issues/14) · 10 de noviembre de 2026

**Estado: propuesta acotada.** Se distingue el equipo de desarrollo comprobado del servidor objetivo todavía no identificado. Las cifras de rendimiento y recuperación son candidatos de aceptación, no resultados medidos ni garantías.

## Inventario comprobado

### Equipo de desarrollo actual

Datos obtenidos mediante consultas de solo lectura durante este issue:

| Elemento | Valor comprobado |
| --- | --- |
| Uso | Desarrollo local; no se declara servidor de producción. |
| Sistema anfitrión | macOS 26.6.2 / Darwin 25.6.0. |
| Arquitectura | `arm64`. |
| Procesador | Apple M1. |
| Memoria física | 8 GiB. |
| Almacenamiento | Volumen raíz APFS de 228 GiB; 52 GiB disponibles al consultar. |
| Docker | Cliente y Engine 29.8.0; contenedores Linux `arm64`. |
| Docker Compose | v5.5.1. |

Este inventario no autoriza formatear, reinstalar ni convertir la Mac en servidor. Go, Node y PostgreSQL continúan ejecutándose en contenedores conforme a las reglas del repositorio.

### Servidor objetivo

No hay evidencia suficiente para identificar el equipo que alojará el portal. Hasta que el propietario proporcione o confirme estos datos, el servidor permanece **PENDIENTE**:

| Dato requerido | Estado |
| --- | --- |
| Fabricante/modelo y responsable físico | Pendiente. |
| CPU y arquitectura (`amd64`/`arm64`) | Pendiente. |
| RAM utilizable | Pendiente. |
| Disco, espacio libre y tipo de almacenamiento | Pendiente. |
| Sistema operativo y versión | Pendiente. |
| Red, dirección fija o reserva, acceso a Internet | Pendiente. |
| Energía, UPS y comportamiento tras reinicio | Pendiente. |
| Horario de disponibilidad y ventanas de mantenimiento | Pendiente. |
| Destino externo y retención de respaldos | Pendiente. |

No se infieren estos valores desde el equipo donde se edita el repositorio. #90 no puede preparar una instalación específica y #82 no puede atribuir capacidad al servidor hasta completar el inventario.

## Metas candidatas de calidad

Se conservan las referencias del plan como propuesta para medir en el servidor confirmado:

| Aspecto | Meta candidata | Condición de medición |
| --- | --- | --- |
| Datos sintéticos | 10 000 estudiantes, 1 000 proyectos y 200 000 movimientos | Generador repetible en `hs-test`; ningún dato real. |
| Concurrencia | Cinco clientes simultáneos | Ensayo controlado con operaciones representativas y sin afectar `hs-dev`. |
| Listados | p95 menor de 500 ms para páginas de 25 filas | Medición dentro de la red local del servidor, con filtros definidos. |
| Dashboard | p95 menor de 1 500 ms | Mismos filtros y conjunto de control; registrar SQL e índices. |
| Excel | Hasta 50 000 filas en menos de 30 s | Registrar tiempo, memoria máxima y tamaño del archivo. |
| Recuperación | RPO máximo de 24 h y RTO máximo de 4 h | Solo se acepta después de restaurar un respaldo en volumen aislado. |
| Salida | Cero defectos críticos o altos abiertos | Recorrido principal y respaldo restaurado con evidencia. |

Estas metas no se ejecutaron en #14. El volumen de referencia se puede reducir para una comprobación rápida de desarrollo, pero una muestra menor no demuestra la meta de aceptación.

## Protocolo de medición

Cada informe de rendimiento debe registrar:

- commit, versión y digests de las imágenes evaluadas;
- inventario del anfitrión, arquitectura y límites asignados a Docker;
- semilla/versión del generador y conteos reales por entidad;
- escenario, filtros, concurrencia, calentamiento y número de muestras;
- p50, p95, máximo y tasa de errores, sin seleccionar solo los mejores resultados;
- consumo máximo de CPU, memoria y disco observado;
- fecha, duración y responsable del ensayo.

Para p95 se requieren al menos 100 observaciones válidas por escenario después del calentamiento. Los errores cuentan y se informan por separado; no se eliminan para mejorar el percentil. Listados, dashboard y exportación se miden de extremo a extremo desde un cliente de la red objetivo, además del análisis interno que corresponda.

Si el equipo confirmado no alcanza una meta, primero se revisan consultas, índices y límites de recursos. Ajustar el objetivo requiere resultado reproducible y aprobación explícita; no se rebaja silenciosamente para declarar éxito.

## Recuperación y disponibilidad

- RPO 24 h implica que la política de respaldos debe limitar a 24 horas la pérdida máxima aceptada; una tarea programada sin evidencia de ejecución no demuestra RPO.
- RTO 4 h se mide desde la declaración del incidente hasta que la versión y los datos restaurados superan las comprobaciones acordadas.
- La restauración se ensaya en un volumen aislado y compara registros de control; nunca se borra `hs-dev_pgdata` para probarla.
- Disponibilidad, mantenimiento, UPS, alertas y destino externo deben acordarse con el equipo objetivo. No se promete continuidad con hardware o red desconocidos.

## Fases de carga

1. **Comprobación rápida:** conjunto pequeño en `hs-test` para validar que el escenario funciona.
2. **Ensayo de referencia:** volumen candidato completo en el servidor confirmado o en hardware equivalente documentado.
3. **Aceptación:** repetición controlada, informe reproducible y aprobación humana de metas o ajustes.

Ninguna fase usa información personal real. No se ejecuta una carga pesada sin comprobar antes espacio libre, límites de Docker, duración estimada y posibilidad de detener el ensayo sin afectar otros entornos.

## Tiempo de desarrollo y fechas

El [registro de horas](REGISTRO_HORAS.md) sigue sin duraciones históricas confirmadas. Este inventario no calcula horas a partir de commits, CI o sesiones de chat, no descuenta automáticamente el presupuesto orientativo de 300 horas y no promete una fecha de entrega.

## Criterio de aprobación

Para cerrar #14 el propietario debe:

1. identificar el servidor objetivo o aprobar expresamente que todavía no existe;
2. confirmar o ajustar las metas candidatas;
3. acordar RPO/RTO y responsable de respaldos;
4. autorizar que #82 use este protocolo cuando exista el sistema funcional.

Hasta entonces se conserva `needs:decision`; no se instala software, no se carga volumen sintético y no se presenta la Mac de desarrollo como servidor.
