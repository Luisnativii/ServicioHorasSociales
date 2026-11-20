# Catálogo territorial del core

El core guarda la jerarquía departamento → municipio → distrito en `public.geo_departments`, `public.geo_municipalities` y `public.geo_districts`. Sus claves y referencias son UUID. El catálogo esperado de El Salvador comprende **14 departamentos, 44 municipios y 262 distritos**. La jerarquía se verifica contra el [Decreto Legislativo 762](https://www.jurisprudencia.gob.sv/DocumentosBoveda/D/2/2020-2029/2023/06/F9625.PDF) y su [reforma 978](https://jurisprudenciaylegislacion.fgr.gob.sv/decreto-no-978-reforma-a-la-ley-especial-para-la-reestructuracion-municipal/).

La carga usa las tres columnas territoriales de la [tabla de población por área del Censo 2024 del BCR/ONEC](https://censo2024.bcr.gob.sv/wp-content/uploads/tablas-geoportal/2025/1104/TAB_POB_AREA_1.pdf). SHA-256 del PDF consultado `c5f6ab3fb97922c8d9519369aec68a2a5a2e9933bbe40772e01fccf7f2b06d48`. Se conservan los ceros iniciales y los nombres publicados por el censo. Se excluyen las filas de totales y los datos de población.

| Nivel | Cobertura | Código almacenado |
| --- | ---: | --- |
| Departamento | 14 | `DEPTO` de dos caracteres, del `01` al `14`. |
| Municipio | 44 | `MUNIC` de dos caracteres, en el ámbito del departamento. |
| Distrito | 262 | `DISTO` de dos caracteres, en la jerarquía de su municipio y departamento. |

Por ejemplo, `01 / 03 / 08` identifica San Francisco Menéndez, municipio Ahuachapán Sur, departamento Ahuachapán. Los códigos sirven para conciliar el catálogo con la fuente. Las referencias del portal y del core usan los UUID existentes de PostgreSQL. [Diccionario del Censo 2024](https://censo2024.bcr.gob.sv/wp-content/uploads/tablas-geoportal/2025/1104/Uso_de_la_Base_de_Datos_del_Censo_2024.pdf)

## Comparación con el catálogo de texto

El [normalizador de texto del core en `3f8d6a4`](https://github.com/jaalgori0/horas-sociales-core-backend/blob/3f8d6a4169b598674b63b4618ced5e4589e96c27/src/modules/shared/geo-el-salvador.ts) enumera los 262 municipios anteriores a la reforma. Se compararon por departamento y nombre con los distritos del censo. Coinciden 257 al ignorar mayúsculas y tildes. Los cinco restantes se corresponden por departamento y denominación histórica:

| Departamento | Texto del core | Nombre censal |
| --- | --- | --- |
| Cabañas | Dolores | Villa Dolores |
| Chalatenango | San José Cancasque | Cancasque |
| Chalatenango | San José Las Flores | Las Flores |
| San Miguel | San Antonio del Mosco | San Antonio |
| Usulután | San Buenaventura | San Buena Ventura |

## Inventario y carga

El inventario de `hs-core-local` en PostgreSQL 16.15 arrojó **0 departamentos, 0 municipios y 0 distritos**. El inventario de la instancia desplegada queda a cargo del operador del core con la conexión autorizada. El [SQL de inventario](../deploy/core-local/territory/inventory.sql) es de solo lectura y muestra conteos, códigos, padres y cobertura por departamento. Ejecútalo antes de acordar la carga y nuevamente después.

```bash
psql -X -f deploy/core-local/territory/inventory.sql
```

El operador selecciona la base autorizada mediante su configuración habitual de PostgreSQL. Para completar un catálogo vacío o parcial compatible, ensaya primero en una copia aislada y revisa los conteos y referencias. El [SQL de carga](../deploy/core-local/territory/load.sql) toma un bloqueo de escritura sobre las tres tablas mientras concilia códigos, nombres y padres. Inserta las filas faltantes en una transacción y conserva los UUID existentes. Si encuentra una fila incompatible, aborta y deja la base como estaba. En ese caso compara el inventario con el censo y acuerda la correspondencia con el core antes de reintentar.

```bash
psql -X -f deploy/core-local/territory/load.sql
psql -X -f deploy/core-local/territory/inventory.sql
```

El ensayo aislado en PostgreSQL 16.15 parte de las migraciones 001–003 y una institución sintética vinculada a un distrito ya cargado. Se verifican **14/44/262**, los **262** distritos con municipio y departamento, los UUID originales, la referencia de la institución y el registro de migraciones. La segunda ejecución debe insertar cero filas. Si aparece una discrepancia, toda la carga debe revertirse. El procedimiento pertenece al operador del core y se ejecuta explícitamente, fuera del arranque del portal y de Goose.
