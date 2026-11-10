# Contrato de modalidad de proyectos

Revisión 0.1 aceptada · Issue [#13](https://github.com/Luisnativii/ServicioHorasSociales/issues/13) · 10 de noviembre de 2026

**Estado: contrato aceptado.** El propietario aceptó D-MO-01 a D-MO-04 al solicitar corregir e integrar los PR #128–#134. Modalidad es un catálogo obligatorio por proyecto; “habilidades blandas” no crea otro campo.

## Decisiones vigentes

- “Modalidad” es un solo concepto.
- “Habilidades blandas” puede ser un valor de modalidad; no es un segundo campo ni otro catálogo.
- El ADMIN puede elegir una modalidad existente o crear una nueva.
- Modalidad no se confunde con tipo, área, naturaleza ni ubicación territorial.
- No se precarga “Habilidades blandas” ni ningún otro valor sin una lista aprobada.

## Decisión aceptada

### D-MO-01 · Catálogo administrable asociado al proyecto

- `modalidad` es un catálogo administrable con identificador, nombre y estado activo.
- Cada proyecto referencia exactamente una modalidad.
- La modalidad es obligatoria al crear un proyecto.
- No pertenece a estudiante, participación, movimiento de horas ni institución.
- No se admiten varias modalidades por proyecto ni texto libre fuera del catálogo.

La capacidad de “crear una nueva” añade una opción al catálogo y luego la selecciona; no guarda una etiqueta aislada solo para ese proyecto.

### D-MO-02 · Nombres y duplicados

- El nombre se recorta y no puede quedar vacío.
- Se rechazan duplicados que solo cambian mayúsculas/minúsculas o espacios exteriores.
- Se conserva el nombre visible ingresado; la comparación canónica no obliga a mostrarlo en minúsculas.
- No se inventan códigos, colores, descripciones ni jerarquías.

### D-MO-03 · Desactivación sin borrado

- Una modalidad usada no se elimina físicamente.
- Desactivarla impide seleccionarla en proyectos nuevos o al cambiar una selección.
- Los proyectos que ya la referencian continúan mostrándola como inactiva.
- Reactivarla permite volver a seleccionarla.

### D-MO-04 · Alcance temporal y de reportes

- La entrega inicial trata modalidad como dato vigente del proyecto, no como dimensión histórica del libro de horas.
- Editar la modalidad de un proyecto con actividad acreditada queda fuera del alcance hasta que #12 defina si requiere historia; la API debe rechazar ese cambio en vez de reatribuir actividad pasada.
- Los reportes no agregan un filtro de modalidad por deducción. Si se requiere, #75 debe definir su semántica y el historial que utiliza.
- “Habilidades blandas” no activa cálculos, resultados de aprendizaje ni evaluaciones especiales.

## Interfaz y permisos

- Solo ADMIN autenticado administra modalidades y proyectos.
- El formulario de proyecto permite seleccionar una modalidad activa.
- Crear una modalidad desde ese flujo usa la misma validación y protección CSRF que el mantenimiento del catálogo.
- No hay acceso público ni roles adicionales.

## Efecto sobre alcance y presupuesto

La funcionalidad requiere una tarea de implementación propia después de aprobar este contrato: migración y consultas del catálogo, API protegida y control web reutilizable en proyectos. No debe añadirse silenciosamente a #40/#42/#43, que están acotados a tipos y áreas.

El trabajo consume el presupuesto orientativo existente de 300 horas; no abre un presupuesto adicional ni acredita horas reales. La estimación y las horas reales se registran únicamente al crear/ejecutar la tarea de implementación, conforme a `REGISTRO_HORAS.md`.

## Efecto sobre el backlog

| Issue | Ajuste tras la aprobación |
| --- | --- |
| Nueva tarea acotada | Implementar catálogo, API y control web de modalidad antes de proyectos. |
| #50 | Añadir la referencia obligatoria de modalidad al proyecto, solo después de integrar la tarea del catálogo. |
| #51 / #52 | Validar y editar modalidad activa en API y formularios de proyecto. |
| #12 | Decidir historia únicamente si se permitirá cambiar modalidad con actividad acreditada. |
| #75 | Incluir modalidad en filtros/reportes solo mediante contrato explícito. |

## Casos de aceptación

1. Crear “Habilidades blandas” produce una modalidad, no un campo adicional.
2. Intentar crear el mismo nombre con otras mayúsculas se rechaza.
3. Una modalidad inactiva sigue visible en un proyecto existente y no aparece para un alta nueva.
4. Un proyecto nuevo sin modalidad se rechaza.
5. Cambiar modalidad después de acreditar actividad se rechaza mientras no exista contrato histórico.

## Registro de aprobación

El 10 de noviembre de 2026 el propietario solicitó corregir e integrar los siete PR documentales abiertos, incluido #130. Esa instrucción acepta D-MO-01 a D-MO-04. La implementación requiere un issue propio y no precargará valores sin fuente.
