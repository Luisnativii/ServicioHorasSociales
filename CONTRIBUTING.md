# Contribuir al Portal de Horas Sociales

El código y las contribuciones se trabajan en `develop`. `main` permanece reservada con un único README.

1. Crear un issue mediante las plantillas y definir sus criterios de aceptación.
2. Seleccionar exactamente una etiqueta `type:` y agregar `status:ready` cuando se decida empezar.
3. Usar la rama que crea la automatización: `tipo/NUMERO-descripcion`.
4. Ejecutar las comprobaciones adecuadas dentro de Docker y subir los cambios a esa rama.
5. Abrir PR hacia `develop`, con título convencional terminado en `(#NUMERO)`, línea `Closes #NUMERO` y una sola etiqueta `version:patch`, `version:minor` o `version:major`.
6. Resolver comentarios y pasar `GitFlow` y `Docker` antes de usar squash merge.

La [guía completa](docs/GITFLOW.md) explica ejemplos, etiquetas, protecciones y versionado. No se suben `.env`, secretos, compilados ni dependencias locales. No se reescriben tags ni se fuerza el historial de las ramas compartidas.
