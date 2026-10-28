`admin/` implementa persistencia del único rol ADMIN con múltiples cuentas. `repository.go` usa consultas generadas de `db/`, normaliza/valida correo y traduce errores sin revelar SQL. El hash se excluye de JSON y no se consulta en listados. Los handlers futuros deben usar DTOs explícitos, nunca serializar modelos SQL.

No hay todavía hashing/CLI de altas, JWT, login ni módulos de negocio. `scripts/check-admin.sh` prueba el repositorio con PostgreSQL real en hs-test y rollback de datos sintéticos; una prueba ordinaria sin el indicador explícito no conecta a la base de desarrollo.
