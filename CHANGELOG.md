# Changelog — Asterion Plugin Contract

Formato basado en [Keep a Changelog](https://keepachangelog.com/es-ES/1.1.0/).
Este proyecto todavía no tiene releases etiquetados en git.

## [Unreleased]

### Fixed
- **`ConfigField.IsSecret()`**: un campo declarado `type: "secret"` (la
  forma que emite el compilador de Asterion Language, `Contract.config(...,
  type="secret")`) se trataba como público en todo lugar que miraba
  directo el booleano `.Secret` — salía en claro en `plugin config show` y
  podía terminar en el `.env` que `asterion plugin export` arma para el
  frontend. `IsSecret()` es ahora el único predicado correcto (`f.Secret
  || f.Type == "secret"`); los cinco sitios que miraban `.Secret` sin pasar
  por acá (en `asterion-core`, en `asterion-graph-cognitive-architecture`,
  y en el `sdk/go/pdk` de este mismo repo) pasan a usarlo. §4 actualizada.
- **`schema/apc-v1.schema.json`** no listaba `"secret"` como valor válido
  de `config_schema[].type` (solo `string`/`number`/`bool`) — un
  `plugin.yaml` real, válido para `apc.Manifest`, fallaba contra este
  espejo. Corregido.

### Added
- **`services`** (`apc.ServiceSpec`, nueva §10): un plugin declara la
  infraestructura externa que necesita (una base `postgres`/`mysql`/
  `mariadb`, o un `redis`) y a qué claves de su propio `config_schema`
  volcar la conexión una vez resuelta (`maps_host`/`maps_port`/
  `maps_user`/`maps_password`/`maps_database`/`maps_url`).
  `Manifest.Validate()` exige nombres únicos, un `kind` soportado, que
  `redis` no declare `database`/`user` (no tiene ninguno de los dos), y
  que cada `maps_*` nombre una clave real de `config_schema` — un typo ahí
  dejaría al plugin silenciosamente sin configurar. Quién resuelve esto
  en la práctica (detectar el motor, crear base/usuario, o levantar un
  contenedor solo si se pide explícito) es `asterion plugin services` en
  `asterion-core`, no este paquete — ver su README y `internal/pluginsvc`.
  `schema/apc-v1.schema.json` ganó la propiedad correspondiente (no puede
  expresar la referencia cruzada a `config_schema`, eso sigue siendo
  responsabilidad exclusiva de `Validate()`).
- **`Contract.service(...)`** en Asterion Language (repo hermano) compila
  a lo de arriba — ver `asterion-language/spec/grammar.md` y
  `asterion-language/examples/plugin-services.asterion`.

### Changed
- **Renumeración de `spec/apc-v1.md`**: §10 "Servicios externos" es
  nueva (ver "Added" arriba) e insertada entre la vieja §9 (Eventos) y la
  vieja §10 (Health check) — todo de ahí en más se corrió +1 (Health
  check pasa a §11, Lifecycle a §12, Versionado a §13, Herramientas a
  §14, "Cómo encaja `asterion-language`" a §15). Mismo criterio que la
  renumeración anterior registrada más abajo en este changelog: el
  contenido no cambia de lugar por capricho, se referencia por número en
  varios puntos del documento y hay que mantenerlo consistente en el
  mismo commit que lo mueve.
- **`spec/apc-v1.md` gana una sección formal de `declared` / `implemented`
  / `enforced`** (nueva §2, todo lo posterior se corrió +1). Motivada por
  una revisión externa del contrato: varias capacidades (`permissions`,
  `events`) están descritas en el mismo nivel de detalle que otras que sí
  tienen efecto real (`contract_version`, `health_path`), sin que el
  documento distinguiera "está en el schema" de "Asterion realmente hace
  algo con esto" de "Asterion lo hace cumplir". La nueva sección define
  los tres niveles y una tabla que clasifica cada capacidad del contrato,
  verificada contra el código real (no contra lo que el spec anterior
  decía en prosa) — en particular, confirmado por grep que `permissions`
  y `events` nunca se leen en ningún `.go` de `asterion-core` más allá de
  pasar la struct de un lado a otro; `permissions` sí se muestra en el
  dashboard (`frontend-core/src/App.tsx`, "Permisos declarados"), `events`
  no se lee en absoluto todavía. §7 (Permisos) y §8 (Eventos, numeración
  vieja) ahora referencian la tabla en vez de repetir la explicación con
  otras palabras.
- **§13 (antes "Cómo encaja `asterion-language`") estaba desactualizada**:
  decía "todavía no existe" sobre un proyecto que para cuando se escribió
  este cambio ya tiene lexer/parser/semantic/CLI funcionando y compila
  manifiestos reales (`asterion plugin from-asterion`, repo hermano
  `asterion-language`). Reescrita para describir la integración real en
  vez de la aspiracional.

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
