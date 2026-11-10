# Contrato de clasificación de proyectos

Revisión 0.1 aceptada · Issue [#7](https://github.com/Luisnativii/ServicioHorasSociales/issues/7) · 10 de noviembre de 2026

**Estado: contrato aceptado.** El propietario aceptó D-CL-01 y D-CL-02 al solicitar corregir e integrar los PR #128–#134. Este documento no crea tablas ni autoriza valores de catálogo sin una fuente.

## Evidencia y decisiones vigentes

| Tema | Estado | Resultado |
| --- | --- | --- |
| Cardinalidad por proyecto | CONFIRMADO | Cada proyecto elige exactamente un tipo y exactamente un área. No se admiten varios tipos ni varias áreas para un mismo proyecto. |
| Tipo frente a área | ACEPTADO | Son catálogos independientes; no existen combinaciones, jerarquía ni dependencia entre ambos. |
| Reclasificación e historia | ACEPTADO | Un cambio es prospectivo y conserva la atribución histórica de la actividad anterior. |
| Valores iniciales | PENDIENTE DE FUENTE | El plan menciona 26 tipos y 12 áreas, pero las fuentes disponibles no contienen sus nombres oficiales. No se inventan ni se precargan etiquetas. |

La confirmación de cardinalidad sustituye las referencias históricas a “un tipo y varias áreas”. No autoriza una tabla puente para tipo o área.

## Decisión aceptada

### D-CL-01 · Catálogos independientes

- Tipo y área son catálogos separados e independientes.
- Elegir un tipo no limita las áreas disponibles y elegir un área no limita los tipos.
- No existe una relación área → tipo ni una lista de pares permitidos mientras una fuente aprobada no la defina.
- Desactivar una opción impide ofrecerla en nuevas altas o reclasificaciones, pero no elimina ni oculta referencias históricas.

Esta alternativa es la mínima compatible con lo confirmado y evita inventar una jerarquía que las fuentes no describen.

### D-CL-02 · Cambios prospectivos con historia

- Una reclasificación cambia el tipo o área vigente del proyecto desde el momento administrativo del cambio.
- No reescribe la clasificación atribuida a actividad registrada con anterioridad.
- Los listados de proyectos actuales muestran la clasificación vigente.
- Los reportes de actividad por período atribuyen cada actividad a la clasificación vigente cuando fue registrada; no trasladan retroactivamente horas anteriores.
- El mecanismo físico de historia o instantánea se define con #12 y con la migración que introduzca la actividad. Este issue aprueba semántica, no una tabla concreta.
- El contrato de filtros y conteos de #75 resolverá cómo contar proyectos únicos cuando hayan cambiado de clasificación, sin alterar la atribución histórica de la actividad.

Esta alternativa mantiene coherencia con la regla ya confirmada para naturaleza: los cambios afectan actividad nueva y conservan la categoría de lo anterior.

## Fuente pendiente para los catálogos

Antes de #41 se necesita una lista revisable que incluya, como mínimo:

- los 26 nombres oficiales de tipo;
- los 12 nombres oficiales de área;
- procedencia y fecha de vigencia;
- confirmación de que no hay duplicados ni relaciones obligatorias entre ambas listas.

Hasta recibirla, #40 puede diseñarse únicamente después de aprobar D-CL-01 y D-CL-02, sin semillas reales. #41 permanece bloqueado para la carga de valores.

## Trabajo que desbloquea la aceptación

| Issue | Alcance desbloqueado |
| --- | --- |
| #12 | Precisar historia académica y temporal de reportes junto con esta semántica de clasificación. |
| #40 | Persistir dos catálogos y una referencia vigente de cada uno por proyecto, sin relaciones muchos-a-muchos. |
| #42 / #43 | Administrar y mostrar tipos y áreas independientes, incluida desactivación sin borrado histórico. |
| #50 | Exigir un tipo y un área al persistir proyectos conforme al contrato aprobado. |
| #65 / #74 / #75 | Diseñar la conservación histórica y los reportes sin reasignar actividad pasada. |

#41 no queda desbloqueado para cargar datos hasta recibir y aprobar los 38 nombres oficiales.

## Registro de aprobación

El 10 de noviembre de 2026 el propietario solicitó corregir e integrar los siete PR documentales abiertos, incluido #128. Esa instrucción acepta expresamente D-CL-01 y D-CL-02. La fuente de los 26 tipos y 12 áreas sigue pendiente y mantiene bloqueada únicamente la carga de #41.
