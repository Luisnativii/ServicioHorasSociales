\set ON_ERROR_STOP on
BEGIN ISOLATION LEVEL REPEATABLE READ READ ONLY;

SELECT current_database() AS base, current_setting('server_version') AS postgres;
SELECT
    (SELECT count(*) FROM public.geo_departments) AS departamentos,
    (SELECT count(*) FROM public.geo_municipalities) AS municipios,
    (SELECT count(*) FROM public.geo_districts) AS distritos,
    (SELECT count(*) FROM public.geo_municipalities AS m
     LEFT JOIN public.geo_departments AS d ON d.id = m.department_id
     WHERE d.id IS NULL) AS municipios_sin_departamento,
    (SELECT count(*) FROM public.geo_districts AS x
     LEFT JOIN public.geo_municipalities AS m ON m.id = x.municipality_id
     LEFT JOIN public.geo_departments AS d ON d.id = m.department_id
     WHERE d.id IS NULL) AS distritos_sin_jerarquia;

SELECT d.codigo AS codigo_departamento, d.nombre AS departamento,
       count(DISTINCT m.id) AS municipios, count(x.id) AS distritos
FROM public.geo_departments AS d
LEFT JOIN public.geo_municipalities AS m ON m.department_id = d.id
LEFT JOIN public.geo_districts AS x ON x.municipality_id = m.id
GROUP BY d.id
ORDER BY d.codigo;

SELECT d.codigo AS codigo_departamento, d.nombre AS departamento,
       m.codigo AS codigo_municipio, m.nombre AS municipio,
       x.codigo AS codigo_distrito, x.nombre AS distrito
FROM public.geo_departments AS d
LEFT JOIN public.geo_municipalities AS m ON m.department_id = d.id
LEFT JOIN public.geo_districts AS x ON x.municipality_id = m.id
ORDER BY d.codigo, m.codigo, x.codigo;

COMMIT;
