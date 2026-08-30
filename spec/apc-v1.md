# Asterion Plugin Contract — v1 (`asterion.plugin/v1`)

> "Un plugin de Asterion es una aplicación independiente. Asterion no
> necesita conocer cómo está implementada; solamente necesita conocer y
> validar el Asterion Plugin Contract."

Este documento es la especificación. La implementación de referencia (lo
que realmente se ejecuta) es el paquete Go [`apc`](../apc/manifest.go) —
si algo acá y el código alguna vez difieren, el código manda y este
documento tiene un bug.

## 1. Principio y forma general

Un plugin es un **proceso HTTP separado**, en cualquier lenguaje, que
Asterion instala, arranca, detiene y administra desde afuera — nunca código
que se carga dentro del proceso de Asterion. Todo lo que Asterion necesita
saber de un plugin está declarado en un único archivo, `plugin.yaml`, en la
raíz de su repositorio. Instalar un plugin nunca ejecuta nada del repo
salvo leer ese archivo.

```
mi-plugin/
├── plugin.yaml          # el manifiesto — obligatorio
├── api/
│   └── openapi.yaml       # opcional: contrato HTTP completo
├── resources/
│   └── schemas/            # opcional: JSON Schema por resource
├── migrations/               # opcional: si permissions.database es true
├── src/                        # el código del plugin — opaco para Asterion
├── tests/
└── README.md
```

## 2. Qué garantiza este contrato hoy (`declared` / `implemented` / `enforced`)

El resto de este documento describe muchas capacidades — pero "está en el
manifiesto" no dice nada por sí solo de qué tan real es esa capacidad hoy.
Esta sección existe para que esa distinción nunca quede implícita: cada
capacidad de este contrato está en exactamente uno de estos tres niveles,
y toda sección de más abajo que introduzca una capacidad nueva debe decir
cuál.

- **`declared`** — Existe en el schema. Asterion lo parsea, valida su
  forma, y lo guarda/pasa de un lado a otro. Por sí solo, no garantiza
  ningún comportamiento en tiempo de ejecución — es una promesa del autor
  del plugin, no un hecho verificado.
- **`implemented`** — Además de estar declarado, alguna herramienta real
  de Asterion lee este campo y hace algo concreto con él: lo muestra en el
  dashboard, dirige un flujo del CLI, genera una UI, dispara una llamada
  HTTP real. Sigue sin ser una garantía — un plugin puede declarar algo
  falso y nada lo detecta.
- **`enforced`** — Asterion activamente previene la violación. No es
  informativo: es un límite real que un plugin no puede cruzar aunque lo
  intente (una instalación rechazada, un arranque que nunca se da por
  exitoso, un dato que nunca vuelve a texto plano).

Los tres niveles son acumulativos pero independientes: una capacidad puede
estar `declared` sin estar `implemented`, e `implemented` sin estar
`enforced`. Ninguno implica el siguiente — hay que verificarlo caso por
caso contra el código real (`apc/`, `asterion-core/internal/plugins`,
`asterion-core/frontend-core`), nunca asumirlo por dónde vive el campo en
el schema.

| Capacidad | `declared` | `implemented` | `enforced` | Detalle |
|---|:---:|:---:|:---:|---|
| `contract_version` | ✅ | ✅ | ✅ | Versión no reconocida → instalación rechazada (`apc.Manifest.Validate`, §12) |
| `name`/`version`/`port`/`start.command` (forma estructural) | ✅ | ✅ | ✅ | `Validate()` rechaza un manifiesto que no cumple la forma obligatoria — nunca llega a instalarse |
| `health_path` | ✅ | ✅ | ✅ | Asterion sondea este endpoint después de `start`; sin una respuesta 2xx, el arranque no se da por exitoso (§10) |
| `config_schema[].secret` | ✅ | ✅ | ✅ | Cifrado real AES-256-GCM (`internal/secretbox`) — el valor en texto plano nunca se vuelve a mostrar (§4) |
| `config_schema` (resto de campos) | ✅ | ✅ | — | Genera el formulario del dashboard; no valida que el plugin realmente use lo que declaró (§4) |
| `resources[].crud` | ✅ | ✅ | — | Controla qué botones ofrece el dashboard y qué llama `asterion plugin dev`; no impide que la API real del plugin acepte otra operación si se la llama directo (§6) |
| `actions` | ✅ | ✅ | — | Igual que `resources[].crud`, como botones sueltos (§7) |
| `api.openapi`, `resources[].schema` | ✅ | ✅ (parcial) | — | `asterion plugin validate` confirma que el archivo existe y es YAML/JSON *sintácticamente* válido — no que sea un OpenAPI 3 o JSON Schema semánticamente correcto (queda para una versión futura del validador, ver el comentario en `apc/validate.go`) |
| `permissions.*` | ✅ | ✅ | ❌ | Se muestra en el dashboard ("Permisos declarados", `frontend-core/src/App.tsx`); el plugin corre igual como proceso normal con los privilegios completos del usuario que lo instaló — nada lo hace cumplir a nivel de sistema operativo (§8) |
| `events.publishes`/`events.subscribes` | ✅ | ❌ | ❌ | Puramente de ida y vuelta por el schema — confirmado: cero referencias a estos campos en todo `asterion-core`, ningún bus de eventos existe todavía (§9) |

Cuando el estado real de una fila cambie (por ejemplo, el día que exista
un bus de eventos real, o un validador de OpenAPI 3 completo), esta tabla
se actualiza en el mismo commit que el código — nunca al revés. Si esta
tabla y el comportamiento real alguna vez difieren, es un bug de este
documento, mismo criterio que el resto del spec.

## 3. El manifiesto (`plugin.yaml`)

Todos los campos nuevos de esta sección son **opcionales** — un
`plugin.yaml` que solo tiene `name`/`version`/`start`/`port` (el formato
anterior a este contrato) sigue siendo válido.

| Campo | Obligatorio | Descripción |
|---|---|---|
| `name` | sí | Identificador único. `^[a-z0-9][a-z0-9_-]{1,63}$` |
| `version` | sí | Versión semántica del plugin (no del contrato) |
| `description`, `author`, `license`, `repo` | no | Metadata |
| `contract_version` | no (default `asterion.plugin/v1`) | Qué versión de este contrato implementa |
| `language.name`, `language.version` | no | Informativo — nunca decide cómo se ejecuta el plugin |
| `start.command`, `start.args` | `start.command` sí | Cómo arrancar el proceso |
| `port` | sí (`0` = auto) | Puerto TCP en loopback; `0` deja que Asterion elija uno libre |
| `health_path` | no (default `/health`) | Endpoint que confirma que el proceso levantó de verdad |
| `config_schema` | no | Qué configuración necesita — genera el formulario en el dashboard |
| `api.base_path`, `api.openapi` | no | Dónde vive la API y, opcionalmente, su OpenAPI |
| `permissions` | no | Qué necesita tocar — declarativo (§8) |
| `resources` | no | Qué recursos administrables expone (§6) |
| `actions` | no | Qué operaciones no-CRUD expone (§7) |
| `events` | no | Qué eventos publica/consume — declarativo (§9) |

### Ejemplo completo

```yaml
name: dummy-fs-provider
version: "1.0.0"
description: "Plugin de referencia — aprovisiona archivos de texto locales"
author: "Asterion"
license: "Apache-2.0"
contract_version: "asterion.plugin/v1"

language:
  name: go
  version: "1.25"

start:
  command: ./dummy-fs-provider
port: 0
health_path: /health

config_schema:
  - key: data_dir
    label: "Directorio donde aprovisionar"
    type: string
    default: "./data"

api:
  base_path: /api/v1
  openapi: api/openapi.yaml

permissions:
  filesystem: ["./data"]

resources:
  - name: files
    endpoint: /files
    schema: resources/schemas/file.json
    primary_key: name
    crud: [create, read, update, delete, list]

actions:
  - name: wipe
    method: POST
    endpoint: /files/wipe

events:
  publishes: [file.created, file.updated, file.deleted]
```

Ver el plugin completo y funcionando en
[`examples/dummy-provider`](../examples/dummy-provider).

## 4. Configuración

Cada entrada de `config_schema` es un dato que el plugin necesita para
operar. Asterion genera el formulario de configuración automáticamente a
partir de esta lista — el plugin no escribe ni una línea de UI. Los campos
marcados `secret: true` se guardan cifrados (AES-256-GCM, clave local) y
nunca se vuelven a mostrar en texto plano.

El plugin recibe su configuración exclusivamente como variables de entorno
`ASTERION_PLUGIN_CONFIG_<CLAVE>` (en mayúsculas) al arrancar — nunca un
archivo, nunca un argumento de línea de comandos. Es el único canal; así un
plugin nunca tiene la tentación de loguear su propio archivo de secretos
por error.

## 5. API

La API del plugin es HTTP, en el `base_path` declarado (default: raíz).
`api.openapi`, si se declara, es una ruta relativa dentro del repo del
plugin a su contrato OpenAPI 3 completo — Asterion lo usa para mostrar
documentación interactiva, no para generar código todavía.

## 6. Resources

Un `resource` es algo que Asterion puede administrar en nombre del plugin
con una UI de tabla/formulario genérica: listar, crear, editar, borrar.
`crud` declara qué subconjunto de esas cinco operaciones (`create`, `read`,
`update`, `delete`, `list`) el endpoint realmente atiende — el dashboard
solo ofrece los botones que el plugin dijo que sabe responder. `schema`,
si se declara, apunta a un JSON Schema del recurso (usado para generar el
formulario de creación/edición con los tipos correctos).

## 7. Actions

No todo es CRUD. Una `action` es una operación con nombre (`issue_invoice`,
`restart_service`, `wipe`) sobre un método y endpoint HTTP específicos. El
dashboard las representa como botones con el nombre declarado, no como
filas de una tabla.

## 8. Permisos

`permissions` es lo que el plugin **declara** que necesita: qué hosts de
red, qué rutas de filesystem, si necesita una base de datos, si maneja
secretos. Estado real, ver tabla de §2: **`declared` + `implemented`, NO
`enforced`.** Asterion se lo muestra al usuario en el panel del plugin
("Permisos declarados") y lo deja registrado, pero no lo hace cumplir a
nivel de sistema operativo: un plugin corre como proceso normal, con los
mismos privilegios del usuario que lo instaló. Forzar esto de verdad
requeriría correr el plugin dentro de un contenedor o VM (ver [Asterion
Lab](https://github.com/Tarafagat/asterion-lab)) en vez de como proceso
directo — un cambio de arquitectura mayor, fuera del alcance de v1, pero
una evolución natural del contrato más adelante.

## 9. Eventos

`events.publishes`/`events.subscribes` documentan qué eventos de dominio
emite o consume el plugin (`invoice.created`, `customer.created`, etc.).
Estado real, ver tabla de §2: **solo `declared`.** No existe todavía un
bus de eventos en Asterion que los transporte, ni código alguno que lea
estos campos. Se incluye desde ahora en el contrato para que un
`plugin.yaml` escrito hoy no tenga que cambiar de forma el día que ese bus
exista — pero un plugin no debe asumir que declarar esto tiene ningún
efecto todavía.

## 10. Health check

Todo plugin debe responder en `health_path` (default `/health`) con un
código HTTP 2xx cuando está sano. `pdk.HealthHandler` (Go) devuelve además
un cuerpo `{"status": "healthy"|"degraded"|"unhealthy", "detail": "..."}`
— Asterion hoy solo mira el código HTTP para decidir si el arranque tuvo
éxito, pero el cuerpo queda disponible para una vista de estado más
detallada en el dashboard.

## 11. Lifecycle

Asterion administra un plugin instalado con estas operaciones (ver
`asterion plugin --help`):

| Comando | Qué hace |
|---|---|
| `install <repo>` | Clona el repo, valida `plugin.yaml`, registra el plugin |
| `config set` / `config show` | Guarda/muestra la configuración (cifrada) |
| `start` | Arranca el proceso, espera el health check |
| `status` | Reconcilia el estado guardado contra si el proceso sigue vivo |
| `stop` | Envía `SIGTERM` |
| `remove` | Detiene, borra el repo clonado y la configuración |
| `connect --project <id>` | Vincula el plugin a un proyecto de Asterion Cloud |

No hay un paso de "update" separado en v1: reinstalar (`remove` + `install`
de nuevo) cubre ese caso hasta que exista una necesidad real de distinguir
"actualizar" de "reinstalar".

## 12. Versionado y compatibilidad

`contract_version` identifica qué versión de este documento implementa el
manifiesto. Si no se declara, se asume `asterion.plugin/v1` (compatibilidad
con manifiestos de antes de que este campo existiera). Si se declara una
versión que Asterion no reconoce, la instalación se rechaza explícitamente
en vez de asumir que es compatible — la lista de versiones soportadas vive
en `apc.ContractVersion` y crece a medida que existan versiones futuras del
contrato.

## 13. Herramientas

- **`asterion plugin init --language go`** — scaffolding: genera un plugin
  en Go que ya cumple el contrato (health check + un resource + una action
  de ejemplo), listo para reemplazar la lógica de negocio.
- **`asterion plugin validate <dir>`** — valida `plugin.yaml` estructuralmente
  y confirma que los archivos que referencia (`api.openapi`,
  `resources[].schema`) existen y son parseables (ver la nota de §2 sobre
  qué tan profunda es esa validación hoy).
- **`asterion plugin dev <dir>`** — sandbox de pruebas local: arranca el
  plugin con su propio `start.command`, espera el health check, y hace
  llamadas de descubrimiento (no destructivas) a cada `resource`/`action`
  declarado, reportando si la API real coincide con lo que el manifiesto
  promete.
- **`asterion plugin from-openapi <openapi.yaml>`** — transformador: si ya
  tenés una API REST propia, infiere un `plugin.yaml` de partida
  (resources/actions) a partir de su OpenAPI, agrupando por patrones CRUD.
  Heurístico, no perfecto — ahorra el primer borrador, no reemplaza el
  criterio del autor.
- **`asterion plugin from-asterion <archivo.asterion>`** — alternativa sin
  heurística a lo anterior: el autor declara cada campo del contrato
  explícito con llamadas `Contract.<verbo>(...)` en Asterion Language, sin
  nada que adivinar. Ver §14.
- **[`sdk/go/pdk`](../sdk/go/pdk)** — utilidades para plugins en Go: config,
  logging estándar, forma de error unificada, health handler.

## 14. Cómo encaja `asterion-language`

Este contrato es la especificación que [`asterion-language`](https://github.com/Tarafagat/asterion-language)
*apunta*, no lo contrario — el paquete `pluginmanifest` de ese repo compila
un `.asterion` con llamadas `Contract.define/config/resource/action/...` a
un `apc.Manifest` real (el mismo tipo Go de este repo, sin una segunda
definición del contrato), invocado como `asterion plugin from-asterion`
(§13). El contrato en sí no depende de que exista un lenguaje propio, y
nunca lo hará: cualquier lenguaje que pueda hablar HTTP puede implementar
un plugin de Asterion — `asterion-language` es una forma más directa de
*escribir* el manifiesto, no un requisito para *ser* un plugin válido.
