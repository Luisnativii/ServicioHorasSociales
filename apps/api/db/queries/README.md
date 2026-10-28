`admin.sql` contiene consultas parametrizadas del bloque de identidad aceptado en #20 e implementado en #21. Crear/consultar/listar/cambiar estado se usan mediante el repositorio interno de administradores, sin endpoints de cuentas en esta entrega.

Desde la raíz: `bash scripts/sqlc.sh generate` y `bash scripts/sqlc.sh vet`. La herramienta corre en Docker con versión/digest fijados. El resultado en `internal/modules/admin/db` es generado: no editarlo a mano. El listado no consulta hashes y tiene orden estable/límite/offset.
