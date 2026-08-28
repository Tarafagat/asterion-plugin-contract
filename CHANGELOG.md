# Changelog — Asterion Plugin Contract

Formato basado en [Keep a Changelog](https://keepachangelog.com/es-ES/1.1.0/).
Este proyecto todavía no tiene releases etiquetados en git.

## [Unreleased]

### Added
- Especificación v1 del Asterion Plugin Contract (`asterion.plugin/v1`):
  manifiesto extendido (`contract_version`, `language`, `api`, `permissions`,
  `resources`, `actions`, `events`) sobre el `plugin.yaml` que ya existía en
  `asterion-core` — retrocompatible, todos los campos nuevos son opcionales.
- Paquete Go `apc`: implementación de referencia del contrato, sin
  dependencias de terceros más allá de `yaml.v3`.
- JSON Schema del manifiesto (`schema/apc-v1.schema.json`), mantenido a
  mano como espejo del código Go, para tooling que no esté en Go.
- Plugin Development Kit para Go (`sdk/go/pdk`): config vía las mismas
  variables de entorno que ya inyecta `internal/plugins.Start`, logging
  JSON estándar, forma de error unificada, health check handler.
- Scaffolding (`sdk/go/scaffold`, usado por `asterion plugin init
  --language go`): genera un plugin en Go que ya cumple el contrato.
- Transformador API→plugin (`openapi`, usado por `asterion plugin
  from-openapi`): infiere `resources`/`actions` a partir de un OpenAPI 3
  existente agrupando por patrones CRUD.
- Plugin de Referencia (`examples/dummy-provider`): "aprovisiona" archivos
  de texto locales en vez de servidores reales — CRUD completo, permisos
  declarados, tests, `openapi.yaml` y JSON Schema reales.
- Plantilla de GitHub Actions para publicar un plugin (`templates/github-actions`).
- **`pdk.MountFrontend`**: un plugin puramente REST, sin ningún frontend
  propio, ya no queda "sin cara" — si `frontend/dist` no existe (o existe
  vacío), el propio Plugin Contract genera y sirve en `GET /` una página
  de documentación armada en el momento a partir del `plugin.yaml` del
  plugin (características, qué configuración necesita, tabla de
  endpoints con método/ruta/estructura) — el mismo contenido que ya
  muestra el panel de un plugin en el dashboard de `asterion-core`, pero
  servido directo por el proceso del plugin. Si `frontend/dist` sí existe,
  se sirve tal cual, sin cambios de comportamiento. `pdk.DefaultFrontendHandler`
  queda disponible aparte para quien quiera montarlo a mano. Adoptado por
  `asterion-mail-plugin-basic` y `asterion-firewall-analysis` como
  reemplazo de su fallback anterior (un JSON de "todavía no hay build").
