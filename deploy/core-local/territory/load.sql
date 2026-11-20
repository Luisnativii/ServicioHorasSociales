-- Fuente: BCR/ONEC, Censo 2024, TAB_POB_AREA_1, columnas territoriales.
-- Procedimiento del operador: docs/CATALOGO_TERRITORIAL_CORE.md.
\set ON_ERROR_STOP on
\encoding UTF8
BEGIN;
SET LOCAL search_path = pg_catalog, pg_temp;

CREATE TEMP TABLE territorio_fuente (
    departamento_codigo text NOT NULL CHECK (departamento_codigo ~ '^(0[1-9]|1[0-4])$'),
    departamento_nombre text NOT NULL CHECK (btrim(departamento_nombre) <> ''),
    municipio_codigo text NOT NULL CHECK (municipio_codigo ~ '^[0-9]{2}$' AND municipio_codigo <> '00'),
    municipio_nombre text NOT NULL CHECK (btrim(municipio_nombre) <> ''),
    distrito_codigo text NOT NULL CHECK (distrito_codigo ~ '^[0-9]{2}$' AND distrito_codigo <> '00'),
    distrito_nombre text NOT NULL CHECK (btrim(distrito_nombre) <> ''),
    PRIMARY KEY (departamento_codigo, municipio_codigo, distrito_codigo),
    UNIQUE (departamento_codigo, distrito_codigo)
) ON COMMIT DROP;

COPY territorio_fuente FROM STDIN;
01	Ahuachapán	01	Ahuachapán Centro	01	Ahuachapán
01	Ahuachapán	01	Ahuachapán Centro	02	Apaneca
01	Ahuachapán	01	Ahuachapán Centro	04	Concepción de Ataco
01	Ahuachapán	01	Ahuachapán Centro	11	Tacuba
01	Ahuachapán	02	Ahuachapán Norte	03	Atiquizaya
01	Ahuachapán	02	Ahuachapán Norte	05	El Refugio
01	Ahuachapán	02	Ahuachapán Norte	09	San Lorenzo
01	Ahuachapán	02	Ahuachapán Norte	12	Turín
01	Ahuachapán	03	Ahuachapán Sur	06	Guaymango
01	Ahuachapán	03	Ahuachapán Sur	07	Jujutla
01	Ahuachapán	03	Ahuachapán Sur	08	San Francisco Menéndez
01	Ahuachapán	03	Ahuachapán Sur	10	San Pedro Puxtla
02	Santa Ana	01	Santa Ana Centro	10	Santa Ana
02	Santa Ana	02	Santa Ana Este	02	Coatepeque
02	Santa Ana	02	Santa Ana Este	04	El Congo
02	Santa Ana	03	Santa Ana Norte	06	Masahuat
02	Santa Ana	03	Santa Ana Norte	07	Metapán
02	Santa Ana	03	Santa Ana Norte	11	Santa Rosa Guachipilín
02	Santa Ana	03	Santa Ana Norte	13	Texistepeque
02	Santa Ana	04	Santa Ana Oeste	01	Candelaria de La Frontera
02	Santa Ana	04	Santa Ana Oeste	03	Chalchuapa
02	Santa Ana	04	Santa Ana Oeste	05	El Porvenir
02	Santa Ana	04	Santa Ana Oeste	08	San Antonio Pajonal
02	Santa Ana	04	Santa Ana Oeste	09	San Sebastián Salitrillo
02	Santa Ana	04	Santa Ana Oeste	12	Santiago de La Frontera
03	Sonsonate	01	Sonsonate Centro	09	Nahulingo
03	Sonsonate	01	Sonsonate Centro	11	San Antonio del Monte
03	Sonsonate	01	Sonsonate Centro	14	Santo Domingo de Guzmán
03	Sonsonate	01	Sonsonate Centro	15	Sonsonate
03	Sonsonate	01	Sonsonate Centro	16	Sonzacate
03	Sonsonate	02	Sonsonate Este	02	Armenia
03	Sonsonate	02	Sonsonate Este	03	Caluco
03	Sonsonate	02	Sonsonate Este	04	Cuisnahuat
03	Sonsonate	02	Sonsonate Este	05	Santa Isabel Ishuatán
03	Sonsonate	02	Sonsonate Este	06	Izalco
03	Sonsonate	02	Sonsonate Este	12	San Julián
03	Sonsonate	03	Sonsonate Norte	07	Juayúa
03	Sonsonate	03	Sonsonate Norte	08	Nahuizalco
03	Sonsonate	03	Sonsonate Norte	10	Salcoatitán
03	Sonsonate	03	Sonsonate Norte	13	Santa Catarina Masahuat
03	Sonsonate	04	Sonsonate Oeste	01	Acajutla
04	Chalatenango	01	Chalatenango Centro	01	Agua Caliente
04	Chalatenango	01	Chalatenango Centro	08	Dulce Nombre de María
04	Chalatenango	01	Chalatenango Centro	10	El Paraíso
04	Chalatenango	01	Chalatenango Centro	13	La Reina
04	Chalatenango	01	Chalatenango Centro	16	Nueva Concepción
04	Chalatenango	01	Chalatenango Centro	22	San Fernando
04	Chalatenango	01	Chalatenango Centro	24	San Francisco Morazán
04	Chalatenango	01	Chalatenango Centro	31	San Rafael
04	Chalatenango	01	Chalatenango Centro	32	Santa Rita
04	Chalatenango	01	Chalatenango Centro	33	Tejutla
04	Chalatenango	02	Chalatenango Norte	04	Citalá
04	Chalatenango	02	Chalatenango Norte	12	La Palma
04	Chalatenango	02	Chalatenango Norte	25	San Ignacio
04	Chalatenango	03	Chalatenango Sur	02	Arcatao
04	Chalatenango	03	Chalatenango Sur	03	Azacualpa
04	Chalatenango	03	Chalatenango Sur	05	Comalapa
04	Chalatenango	03	Chalatenango Sur	06	Concepción Quezaltepeque
04	Chalatenango	03	Chalatenango Sur	07	Chalatenango
04	Chalatenango	03	Chalatenango Sur	09	El Carrizal
04	Chalatenango	03	Chalatenango Sur	11	La Laguna
04	Chalatenango	03	Chalatenango Sur	14	Las Vueltas
04	Chalatenango	03	Chalatenango Sur	15	Nombre de Jesús
04	Chalatenango	03	Chalatenango Sur	17	Nueva Trinidad
04	Chalatenango	03	Chalatenango Sur	18	Ojos de Agua
04	Chalatenango	03	Chalatenango Sur	19	Potonico
04	Chalatenango	03	Chalatenango Sur	20	San Antonio de La Cruz
04	Chalatenango	03	Chalatenango Sur	21	San Antonio Los Ranchos
04	Chalatenango	03	Chalatenango Sur	23	San Francisco Lempa
04	Chalatenango	03	Chalatenango Sur	26	San Isidro Labrador
04	Chalatenango	03	Chalatenango Sur	27	Cancasque
04	Chalatenango	03	Chalatenango Sur	28	Las Flores
04	Chalatenango	03	Chalatenango Sur	29	San Luis del Carmen
04	Chalatenango	03	Chalatenango Sur	30	San Miguel de Mercedes
05	La Libertad	01	La Libertad Centro	02	Ciudad Arce
05	La Libertad	01	La Libertad Centro	15	San Juan Opico
05	La Libertad	02	La Libertad Costa	05	Chiltiupán
05	La Libertad	02	La Libertad Costa	08	Jicalapa
05	La Libertad	02	La Libertad Costa	09	La Libertad
05	La Libertad	02	La Libertad Costa	18	Tamanique
05	La Libertad	02	La Libertad Costa	20	Teotepeque
05	La Libertad	03	La Libertad Este	01	Antiguo Cuscatlán
05	La Libertad	03	La Libertad Este	06	Huizúcar
05	La Libertad	03	La Libertad Este	10	Nuevo Cuscatlán
05	La Libertad	03	La Libertad Este	14	San José Villanueva
05	La Libertad	03	La Libertad Este	22	Zaragoza
05	La Libertad	04	La Libertad Norte	12	Quezaltepeque
05	La Libertad	04	La Libertad Norte	16	San Matías
05	La Libertad	04	La Libertad Norte	17	San Pablo Tacachico
05	La Libertad	05	La Libertad Oeste	03	Colón
05	La Libertad	05	La Libertad Oeste	07	Jayaque
05	La Libertad	05	La Libertad Oeste	13	Sacacoyo
05	La Libertad	05	La Libertad Oeste	19	Talnique
05	La Libertad	05	La Libertad Oeste	21	Tepecoyo
05	La Libertad	06	La Libertad Sur	04	Comasagua
05	La Libertad	06	La Libertad Sur	11	Santa Tecla
06	San Salvador	01	San Salvador Centro	03	Ayutuxtepeque
06	San Salvador	01	San Salvador Centro	04	Cuscatancingo
06	San Salvador	01	San Salvador Centro	08	Mejicanos
06	San Salvador	01	San Salvador Centro	14	San Salvador
06	San Salvador	01	San Salvador Centro	19	Ciudad Delgado
06	San Salvador	02	San Salvador Este	07	Ilopango
06	San Salvador	02	San Salvador Este	13	San Martín
06	San Salvador	02	San Salvador Este	17	Soyapango
06	San Salvador	02	San Salvador Este	18	Tonacatepeque
06	San Salvador	03	San Salvador Norte	01	Aguilares
06	San Salvador	03	San Salvador Norte	05	El Paisnal
06	San Salvador	03	San Salvador Norte	06	Guazapa
06	San Salvador	04	San Salvador Oeste	02	Apopa
06	San Salvador	04	San Salvador Oeste	09	Nejapa
06	San Salvador	05	San Salvador Sur	10	Panchimalco
06	San Salvador	05	San Salvador Sur	11	Rosario de Mora
06	San Salvador	05	San Salvador Sur	12	San Marcos
06	San Salvador	05	San Salvador Sur	15	Santiago Texacuangos
06	San Salvador	05	San Salvador Sur	16	Santo Tomás
07	Cuscatlán	01	Cuscatlán Norte	06	Oratorio de Concepción
07	Cuscatlán	01	Cuscatlán Norte	07	San Bartolomé Perulapía
07	Cuscatlán	01	Cuscatlán Norte	09	San José Guayabal
07	Cuscatlán	01	Cuscatlán Norte	10	San Pedro Perulapán
07	Cuscatlán	01	Cuscatlán Norte	15	Suchitoto
07	Cuscatlán	02	Cuscatlán Sur	01	Candelaria
07	Cuscatlán	02	Cuscatlán Sur	02	Cojutepeque
07	Cuscatlán	02	Cuscatlán Sur	03	El Carmen
07	Cuscatlán	02	Cuscatlán Sur	04	El Rosario
07	Cuscatlán	02	Cuscatlán Sur	05	Monte San Juan
07	Cuscatlán	02	Cuscatlán Sur	08	San Cristobal
07	Cuscatlán	02	Cuscatlán Sur	11	San Rafael Cedros
07	Cuscatlán	02	Cuscatlán Sur	12	San Ramón
07	Cuscatlán	02	Cuscatlán Sur	13	Santa Cruz Analquito
07	Cuscatlán	02	Cuscatlán Sur	14	Santa Cruz Michapa
07	Cuscatlán	02	Cuscatlán Sur	16	Tenancingo
08	La Paz	01	La Paz Centro	02	El Rosario
08	La Paz	01	La Paz Centro	03	Jerusalén
08	La Paz	01	La Paz Centro	04	Mercedes La Ceiba
08	La Paz	01	La Paz Centro	06	Paraíso de Osorio
08	La Paz	01	La Paz Centro	07	San Antonio Masahuat
08	La Paz	01	La Paz Centro	08	San Emigdio
08	La Paz	01	La Paz Centro	12	San Juan Tepezontes
08	La Paz	01	La Paz Centro	14	San Miguel Tepezontes
08	La Paz	01	La Paz Centro	16	San Pedro Nonualco
08	La Paz	01	La Paz Centro	18	Santa María Ostuma
08	La Paz	01	La Paz Centro	19	Santiago Nonualco
08	La Paz	01	La Paz Centro	22	San Luis La Herradura
08	La Paz	02	La Paz Este	10	San Juan Nonualco
08	La Paz	02	La Paz Este	17	San Rafael Obrajuelo
08	La Paz	02	La Paz Este	21	Zacatecoluca
08	La Paz	03	La Paz Oeste	01	Cuyultitán
08	La Paz	03	La Paz Oeste	05	Olocuilta
08	La Paz	03	La Paz Oeste	09	San Francisco Chinameca
08	La Paz	03	La Paz Oeste	11	San Juan Talpa
08	La Paz	03	La Paz Oeste	13	San Luis Talpa
08	La Paz	03	La Paz Oeste	15	San Pedro Masahuat
08	La Paz	03	La Paz Oeste	20	Tapalhuaca
09	Cabañas	01	Cabañas Este	02	Guacotecti
09	Cabañas	01	Cabañas Este	05	San Isidro
09	Cabañas	01	Cabañas Este	06	Sensuntepeque
09	Cabañas	01	Cabañas Este	08	Victoria
09	Cabañas	01	Cabañas Este	09	Villa Dolores
09	Cabañas	02	Cabañas Oeste	01	Cinquera
09	Cabañas	02	Cabañas Oeste	03	Ilobasco
09	Cabañas	02	Cabañas Oeste	04	Jutiapa
09	Cabañas	02	Cabañas Oeste	07	Tejutepeque
10	San Vicente	01	San Vicente Norte	01	Apastepeque
10	San Vicente	01	San Vicente Norte	04	Santa Clara
10	San Vicente	01	San Vicente Norte	05	Santo Domingo
10	San Vicente	01	San Vicente Norte	06	San Esteban Catarina
10	San Vicente	01	San Vicente Norte	07	San Ildefonso
10	San Vicente	01	San Vicente Norte	08	San Lorenzo
10	San Vicente	01	San Vicente Norte	09	San Sebastián
10	San Vicente	02	San Vicente Sur	02	Guadalupe
10	San Vicente	02	San Vicente Sur	03	San Cayetano Istepeque
10	San Vicente	02	San Vicente Sur	10	San Vicente
10	San Vicente	02	San Vicente Sur	11	Tecoluca
10	San Vicente	02	San Vicente Sur	12	Tepetitán
10	San Vicente	02	San Vicente Sur	13	Verapaz
11	Usulután	01	Usulután Este	03	California
11	Usulután	01	Usulután Este	04	Concepción Batres
11	Usulután	01	Usulután Este	06	Ereguayquín
11	Usulután	01	Usulután Este	10	Jucuarán
11	Usulután	01	Usulután Este	13	Ozatlán
11	Usulután	01	Usulután Este	17	San Dionisio
11	Usulután	01	Usulután Este	18	Santa Elena
11	Usulután	01	Usulután Este	20	Santa María
11	Usulután	01	Usulután Este	22	Tecapán
11	Usulután	01	Usulután Este	23	Usulután
11	Usulután	02	Usulután Norte	01	Alegría
11	Usulután	02	Usulután Norte	02	Berlín
11	Usulután	02	Usulután Norte	05	El Triunfo
11	Usulután	02	Usulután Norte	07	Estanzuelas
11	Usulután	02	Usulután Norte	09	Jucuapa
11	Usulután	02	Usulután Norte	11	Mercedes Umaña
11	Usulután	02	Usulután Norte	12	Nueva Granada
11	Usulután	02	Usulután Norte	16	San Buena Ventura
11	Usulután	02	Usulután Norte	21	Santiago de María
11	Usulután	03	Usulután Oeste	08	Jiquilisco
11	Usulután	03	Usulután Oeste	14	Puerto El Triunfo
11	Usulután	03	Usulután Oeste	15	San Agustín
11	Usulután	03	Usulután Oeste	19	San Francisco Javier
12	San Miguel	01	San Miguel Centro	03	Comacarán
12	San Miguel	01	San Miguel Centro	06	Chirilagua
12	San Miguel	01	San Miguel Centro	09	Moncagua
12	San Miguel	01	San Miguel Centro	12	Quelepa
12	San Miguel	01	San Miguel Centro	17	San Miguel
12	San Miguel	01	San Miguel Centro	20	Uluazapa
12	San Miguel	02	San Miguel Norte	01	Carolina
12	San Miguel	02	San Miguel Norte	02	Ciudad Barrios
12	San Miguel	02	San Miguel Norte	04	Chapeltique
12	San Miguel	02	San Miguel Norte	11	Nuevo Edén de San Juan
12	San Miguel	02	San Miguel Norte	13	San Antonio
12	San Miguel	02	San Miguel Norte	14	San Gerardo
12	San Miguel	02	San Miguel Norte	16	San Luis de La Reina
12	San Miguel	02	San Miguel Norte	19	Sesori
12	San Miguel	03	San Miguel Oeste	05	Chinameca
12	San Miguel	03	San Miguel Oeste	07	El Tránsito
12	San Miguel	03	San Miguel Oeste	08	Lolotique
12	San Miguel	03	San Miguel Oeste	10	Nueva Guadalupe
12	San Miguel	03	San Miguel Oeste	15	San Jorge
12	San Miguel	03	San Miguel Oeste	18	San Rafael Oriente
13	Morazán	01	Morazán Norte	01	Arambala
13	Morazán	01	Morazán Norte	02	Cacaopera
13	Morazán	01	Morazán Norte	03	Corinto
13	Morazán	01	Morazán Norte	07	El Rosario
13	Morazán	01	Morazán Norte	10	Joateca
13	Morazán	01	Morazán Norte	11	Jocoaitique
13	Morazán	01	Morazán Norte	14	Meanguera
13	Morazán	01	Morazán Norte	16	Perquín
13	Morazán	01	Morazán Norte	18	San Fernando
13	Morazán	01	Morazán Norte	20	San Isidro
13	Morazán	01	Morazán Norte	24	Torola
13	Morazán	02	Morazán Sur	04	Chilanga
13	Morazán	02	Morazán Sur	05	Delicias de Concepción
13	Morazán	02	Morazán Sur	06	El Divisadero
13	Morazán	02	Morazán Sur	08	Gualococti
13	Morazán	02	Morazán Sur	09	Guatajiagua
13	Morazán	02	Morazán Sur	12	Jocoro
13	Morazán	02	Morazán Sur	13	Lolotiquillo
13	Morazán	02	Morazán Sur	15	Osicala
13	Morazán	02	Morazán Sur	17	San Carlos
13	Morazán	02	Morazán Sur	19	San Francisco Gotera
13	Morazán	02	Morazán Sur	21	San Simón
13	Morazán	02	Morazán Sur	22	Sensembra
13	Morazán	02	Morazán Sur	23	Sociedad
13	Morazán	02	Morazán Sur	25	Yamabal
13	Morazán	02	Morazán Sur	26	Yoloaiquín
14	La Unión	01	La Unión Norte	01	Anamorós
14	La Unión	01	La Unión Norte	02	Bolívar
14	La Unión	01	La Unión Norte	03	Concepción de Oriente
14	La Unión	01	La Unión Norte	06	El Sauce
14	La Unión	01	La Unión Norte	09	Lislique
14	La Unión	01	La Unión Norte	11	Nueva Esparta
14	La Unión	01	La Unión Norte	12	Pasaquina
14	La Unión	01	La Unión Norte	13	Polorós
14	La Unión	01	La Unión Norte	15	San José
14	La Unión	01	La Unión Norte	16	Santa Rosa de Lima
14	La Unión	02	La Unión Sur	04	Conchagua
14	La Unión	02	La Unión Sur	05	El Carmen
14	La Unión	02	La Unión Sur	07	Intipucá
14	La Unión	02	La Unión Sur	08	La Unión
14	La Unión	02	La Unión Sur	10	Meanguera del Golfo
14	La Unión	02	La Unión Sur	14	San Alejo
14	La Unión	02	La Unión Sur	17	Yayantique
14	La Unión	02	La Unión Sur	18	Yucuaiquín
\.

CREATE TEMP TABLE territorio_departamentos ON COMMIT DROP AS
SELECT DISTINCT departamento_codigo AS codigo, departamento_nombre AS nombre
FROM territorio_fuente;
ALTER TABLE territorio_departamentos ADD PRIMARY KEY (codigo), ADD UNIQUE (nombre);

CREATE TEMP TABLE territorio_municipios ON COMMIT DROP AS
SELECT DISTINCT departamento_codigo, municipio_codigo AS codigo, municipio_nombre AS nombre
FROM territorio_fuente;
ALTER TABLE territorio_municipios
    ADD PRIMARY KEY (departamento_codigo, codigo), ADD UNIQUE (departamento_codigo, nombre);

LOCK TABLE public.geo_departments, public.geo_municipalities, public.geo_districts
    IN SHARE ROW EXCLUSIVE MODE;

DO $verificar$
BEGIN
    IF (SELECT count(*) FROM territorio_departamentos) <> 14
       OR (SELECT count(*) FROM territorio_municipios) <> 44
       OR (SELECT count(*) FROM territorio_fuente) <> 262 THEN
        RAISE EXCEPTION 'Cobertura oficial inesperada: se esperan 14 departamentos, 44 municipios y 262 distritos';
    END IF;

    -- Una discrepancia requiere conciliación por el operador antes de insertar.
    IF EXISTS (
        SELECT codigo, nombre FROM public.geo_departments
        EXCEPT
        SELECT codigo, nombre FROM territorio_departamentos
    ) OR EXISTS (
        SELECT d.codigo, m.codigo, m.nombre
        FROM public.geo_municipalities AS m
        JOIN public.geo_departments AS d ON d.id = m.department_id
        EXCEPT
        SELECT departamento_codigo, codigo, nombre FROM territorio_municipios
    ) OR EXISTS (
        SELECT d.codigo, m.codigo, x.codigo, x.nombre
        FROM public.geo_districts AS x
        JOIN public.geo_municipalities AS m ON m.id = x.municipality_id
        JOIN public.geo_departments AS d ON d.id = m.department_id
        EXCEPT
        SELECT departamento_codigo, municipio_codigo, distrito_codigo, distrito_nombre
        FROM territorio_fuente
    ) THEN
        RAISE EXCEPTION 'El catálogo del core difiere de la fuente censal. Ejecuta inventory.sql y concilia los códigos, nombres y padres con el operador';
    END IF;
END
$verificar$;

INSERT INTO public.geo_departments (codigo, nombre)
SELECT codigo, nombre FROM territorio_departamentos
ON CONFLICT (codigo) DO NOTHING;

INSERT INTO public.geo_municipalities (department_id, codigo, nombre)
SELECT d.id, m.codigo, m.nombre
FROM territorio_municipios AS m
JOIN public.geo_departments AS d ON d.codigo = m.departamento_codigo
ON CONFLICT (department_id, codigo) DO NOTHING;

INSERT INTO public.geo_districts (municipality_id, codigo, nombre)
SELECT m.id, x.distrito_codigo, x.distrito_nombre
FROM territorio_fuente AS x
JOIN public.geo_departments AS d ON d.codigo = x.departamento_codigo
JOIN public.geo_municipalities AS m ON m.department_id = d.id AND m.codigo = x.municipio_codigo
ON CONFLICT (municipality_id, codigo) DO NOTHING;

COMMIT;
