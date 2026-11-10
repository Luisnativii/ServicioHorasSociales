# Contrato de coordinadores y contactos

Revisión 0.1 para aprobación · Issue [#11](https://github.com/Luisnativii/ServicioHorasSociales/issues/11) · 9 de noviembre de 2026

**Estado: propuesta acotada.** El propietario ya decidió guardar coordinadores como texto repetible, sin tabla, catálogo ni enum. La ubicación exacta, los campos y su obligatoriedad requieren aceptación expresa antes de generar SQL.

## Decisiones vigentes

- Un coordinador o contacto no obtiene una cuenta de acceso por estar registrado.
- El coordinador se guarda como dato textual y puede repetirse en distintas carreras.
- No se deduplican personas ni se crea una entidad global de persona.
- El contacto institucional y el responsable de un proyecto son datos independientes.
- Teléfonos y extensiones se conservan como texto.
- Editar el contacto de una institución no modifica responsables ya guardados en sus proyectos.

Estas decisiones sustituyen la propuesta histórica de tablas `coordinador` y `carrera_coordinador`.

## Propuesta para aprobación

### D-PE-01 · Coordinador académico en la carrera

- Cada carrera contiene un único campo opcional `coordinador_nombre`.
- Es texto libre no vacío cuando se informa; el mismo nombre puede aparecer en varias carreras.
- No hay identificador de coordinador, relación muchos-a-muchos, catálogo, desactivación ni pantalla independiente.
- Cambiar el texto actualiza la carrera; no se interpreta como cambio de cuenta o permisos.

Esta propuesta traduce la simplificación del propietario sin mantener la cardinalidad antigua de “varios coordinadores por carrera”.

### D-PE-02 · Contacto principal de la institución

La institución conserva un solo bloque de contacto principal en la misma fila:

| Campo candidato | Obligatorio propuesto | Regla |
| --- | --- | --- |
| `contacto_nombre` | Sí | Texto no vacío. |
| `contacto_correo` | No | Si se informa, debe tener formato de correo válido. |
| `contacto_telefono` | No | Texto; conserva prefijos, espacios y ceros. |
| `contacto_extension` | No | Texto; solo tiene sentido junto con teléfono. |

No se guarda una lista de contactos institucionales ni una persona reutilizable. La edición modifica únicamente el bloque actual de la institución.

### D-PE-03 · Responsable propio del proyecto

Cada proyecto conserva otro bloque, independiente del institucional:

| Campo candidato | Obligatorio propuesto | Regla |
| --- | --- | --- |
| `responsable_nombre` | Sí | Texto no vacío. |
| `responsable_correo` | No | Si se informa, debe tener formato de correo válido. |
| `responsable_telefono` | No | Texto; conserva prefijos, espacios y ceros. |
| `responsable_extension` | No | Texto; solo tiene sentido junto con teléfono. |

El formulario puede facilitar copiar datos como entrada inicial, pero la persistencia no crea un vínculo vivo: después de guardar, ambos bloques se editan por separado.

### D-PE-04 · Sin identidad ni auditoría de personas por deducción

- Ningún campo referencia `administrador` ni crea roles de coordinador o contacto.
- El ADMIN autenticado sigue siendo quien realiza y, cuando corresponda, audita la operación.
- No se infieren documento de identidad, cargo, género, dirección, múltiples teléfonos ni otros datos personales.
- No se conserva historial de cambios de estos textos hasta que un requisito posterior lo apruebe expresamente.

## Efecto sobre el backlog

| Issue | Ajuste requerido tras la aprobación |
| --- | --- |
| #34 / #35 / #36 | La carrera expone `coordinador_nombre`; no crea módulo ni catálogo de coordinadores. |
| #44 | Sustituir “coordinadores y vínculos académicos” por persistencia/validación del coordinador textual si no quedó incluida en #34. |
| #45 | Editar el coordinador dentro del formulario de carrera; retirar administración independiente y asociaciones. |
| #46 / #47 / #49 | Persistir y editar el bloque de contacto principal dentro de institución. |
| #48 | Limitarse a validación y mantenimiento del contacto institucional en línea; no crear tabla de personas. |
| #50 / #51 / #52 | Persistir y editar el responsable propio del proyecto, separado del contacto institucional. |

Los títulos y criterios históricos de #44/#45/#48 deben ajustarse antes de iniciarlos para no restaurar tablas descartadas.

## Casos de aceptación

1. Dos carreras pueden guardar el mismo `coordinador_nombre` sin compartir una fila ni una cuenta.
2. Una institución y uno de sus proyectos pueden tener nombres y teléfonos distintos.
3. Editar el contacto institucional no altera el responsable del proyecto existente.
4. Teléfono `+503 2222-0000` y extensión `004` conservan su representación textual.
5. Un correo informado con formato inválido se rechaza; omitirlo está permitido.

## Criterio de aprobación

El propietario debe aceptar expresamente D-PE-01 a D-PE-04 o indicar los campos/obligatoriedades que desea cambiar. Hasta entonces se mantiene `needs:decision` y no se generan migraciones.
