# Asterion Plugin Contract

El contrato obligatorio que cualquier plugin de Asterion debe cumplir para
ser instalado, ejecutado y administrado por
[asterion-core](https://github.com/Tarafagat/asterion-core). Un plugin es
una aplicación HTTP independiente, en cualquier lenguaje — Asterion no
necesita conocer cómo está implementada, solo necesita conocer y validar
este contrato.

**Especificación completa**: [`spec/apc-v1.md`](spec/apc-v1.md).

## Qué hay en este repo

| Carpeta | Qué es |
|---|---|
| [`spec/`](spec) | La especificación en prosa, v1 |
| [`schema/`](schema) | JSON Schema del manifiesto, para tooling no-Go |
| [`apc/`](apc) | Implementación de referencia en Go: tipos, `LoadManifest`, `Validate` — la fuente de verdad real |
| [`sdk/go/pdk/`](sdk/go/pdk) | Plugin Development Kit para Go: config, logging, errores, health check |
| [`sdk/go/scaffold/`](sdk/go/scaffold) | Generador usado por `asterion plugin init --language go` |
| [`openapi/`](openapi) | El transformador API→plugin (`asterion plugin from-openapi`) |
| [`examples/dummy-provider/`](examples/dummy-provider) | El Plugin de Referencia — funcionando, con tests |
| [`templates/github-actions/`](templates/github-actions) | Plantilla de CI para publicar un plugin |

## Cómo se conecta con `asterion-core`

`internal/plugins.Manifest` en `asterion-core` es un alias directo de
`apc.Manifest` — no hay dos definiciones del contrato, `asterion-core`
importa este módulo igual que ya importa
[`asterion-lab`](https://github.com/Tarafagat/asterion-lab): como carpeta
hermana, vía `replace` en `go.mod`, porque todavía no está publicado en
ningún registry.

```
asterion/
├── asterion-core/
├── asterion-lab/
└── asterion-plugin-contract/   ← este repo, hermano de los otros dos
```

Todo el ciclo de vida de un plugin se administra desde el CLI de
`asterion-core` (`asterion plugin install/start/stop/validate/init/dev/...`)
— este repo no tiene binario propio, es la especificación más las
herramientas que la implementan.

## Empezar a crear un plugin

```
asterion plugin init --language go --dir mi-plugin
cd mi-plugin
go build -o mi-plugin .
asterion plugin validate .
```

O, si ya tenés una API REST propia:

```
asterion plugin from-openapi openapi.yaml --out mi-plugin
```

Para ver un plugin completo, con permisos, resources, actions y tests
reales, mirá [`examples/dummy-provider`](examples/dummy-provider).

## Licencia

Apache License 2.0 — igual que `asterion-core`, `asterion-lab` y
`asterion-shared`: cuanto más fácil sea construir sobre este contrato, más
rápido crece el ecosistema de plugins.
