# Identidad visual — Archivo

Revisión 1 · 28 de octubre de 2026 · Issue [#109](https://github.com/Luisnativii/ServicioHorasSociales/issues/109).

**Dirección elegida por el propietario:** Archivo. Acepta su apariencia y solicita que «Nuevo estudiante» abra una página propia, en lugar de desplegar el formulario debajo del listado. Este documento registra esa elección y concreta su uso; no crea pantallas de negocio ni aprueba campos del estudiante todavía pendientes.

## Paleta y superficies

| Token | Claro | Oscuro | Uso |
| --- | --- | --- | --- |
| Fondo | `#F3F1EB` | `#192029` | Área de trabajo, tono piedra |
| Superficie | `#FFFEFA` | `#222B35` | Tablas, campos y formularios |
| Texto principal | `#26323A` | `#ECEDE9` | Contenido y títulos |
| Texto secundario | `#606B73` | `#B5BDC3` | Etiquetas y contexto |
| Acento | `#253D53` | `#B7CFDF` | Acción principal, selección y enlaces |
| Texto sobre acento | `#FFFEFA` | `#152B3E` | Botón principal |
| Detalle cobre | `#975339` | `#DBA78C` | Acentos puntuales, sin comunicar estado por sí solo |
| Bordes | `#DCDEDB` | `#3B4651` | Separadores de tabla y campos |
| Navegación | `#E8E7E1` | `#1E2731` | Barra lateral |

Contrastes calculados sobre el fondo claro: texto principal 11.62:1, secundario superior a 4.5:1 y botón principal 11.12:1. En oscuro, el texto secundario supera 8.6:1 sobre el fondo. La implementación comprueba además la combinación efectiva de cada componente; el foco y los estados deben conservar contraste en ambas apariencias.

Propuesta de estados semánticos: error `#A52A2A` / `#F1A8A8`, éxito `#286344` / `#A8D3B7` y advertencia `#755318` / `#E5C985` (claro/oscuro). Siempre acompañar color con texto o icono con nombre accesible. No usar cobre como sustituto indistinto de error/éxito.

## Tipografía y distribución

- Títulos de página: Georgia o serif equivalente instalada, 30–34 px en escritorio y 28 px en móvil. El nombre del sistema usa la misma familia con moderación.
- Cuerpo, navegación, formularios y tablas: Segoe UI, Helvetica Neue, Arial, sans-serif; 14–16 px. Texto secundario mínimo 12 px.
- Números de horas alineados por columna y con cifras tabulares. No usar serif en tablas ni texto de controles.
- Barra lateral sobria en escritorio; navegación adaptable en móvil. Ancho del contenido fluido con márgenes de 24–32 px; reducir a 16 px en pantallas pequeñas.
- Ritmo de separación de 4/8 px. Campos y botones con esquinas de 4 px, tablas con separadores horizontales. Una acción principal por tarea.
- El tema puede seguir el sistema o la preferencia elegida en la interfaz cuando se implemente. Ambos conservan la misma estructura e identidad.

## Navegación de estudiantes — decisión expresa

1. El listado vive en una página propia, propuesta `/estudiantes`.
2. «Nuevo estudiante» navega dentro de la misma aplicación a otra página, propuesta `/estudiantes/nuevo`; conserva la navegación general y el contexto «Estudiantes / Nuevo».
3. El formulario reemplaza el área principal del listado. No se despliega debajo de la tabla ni en un modal.
4. Incluir «Volver a estudiantes» y «Cancelar». Volver conserva búsqueda/página/filtros del listado cuando sea posible.
5. Si hay cambios sin guardar, confirmar antes de abandonar; cancelar esa confirmación conserva lo escrito.
6. Al guardar correctamente, navegar al expediente o al listado según el contrato del módulo, mostrando la confirmación. Esa elección funcional se concreta antes de implementar #57.

La muestra usa datos y campos ilustrativos. Obligatoriedades, validaciones, catálogos y campos finales corresponden a #55/#56/#57 y sus decisiones, no a esta maqueta.

## Componentes y estados

- Formularios con etiqueta visible sobre cada control, ayuda próxima al campo y errores concretos junto al dato incorrecto. Mantener lo escrito ante errores.
- Resumen de errores al enviar si hay varios campos; foco en el primer error o en el resumen, sin depender solo del color.
- Botones con verbo y objeto: «Guardar estudiante», «Cancelar», «Volver a estudiantes». Mostrar progreso y evitar doble envío mientras se procesa.
- Tablas con encabezados legibles, nombre del estudiante como enlace y horas internas/externas en columnas separadas; no inventar estadísticas ni completar valores desconocidos con cero.
- Distinguir carga, lista vacía, búsqueda sin coincidencias y error de conexión. Cada estado tiene mensaje breve y acción útil cuando corresponda.
- Mantener foco visible y navegación con teclado. Controles táctiles con área suficiente; nombres accesibles en iconos sin texto.
- En móvil, formulario en una columna y navegación reordenada. Una tabla que requiera varias columnas puede desplazarse en su propio contenedor, sin provocar desbordamiento de toda la página.

## Uso en los issues

- #31: acceso administrativo con estos colores, tipografía y estados.
- #32: navegación general, layout y componentes comunes.
- #36 y siguientes: reutilizar componentes, sin rediseñar la identidad en cada módulo.
- #57: formulario de estudiante en ruta independiente, según la corrección del propietario.

El backend de identidad no depende de esta elección. La aplicación actual sigue mostrando la base técnica hasta implementar sus pantallas; no confundir la comparación visual con un CRUD existente. El PR de esta guía deja la decisión disponible para los issues de interfaz.
