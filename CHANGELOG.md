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
