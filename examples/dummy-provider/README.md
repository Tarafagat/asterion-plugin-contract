# dummy-fs-provider — el Plugin de Referencia

Este es el "estándar de oro" del Asterion Plugin Contract: un plugin real,
completo y funcionando, que cualquiera puede leer para entender cómo se ve
un plugin bien hecho. En vez de aprovisionar servidores en la nube (lo que
costaría dinero real solo para poder probarlo), "aprovisiona" archivos de
texto en un directorio local — mismo patrón de CRUD + permisos + health +
lifecycle que tendría un plugin de verdad, sin la cuenta de AWS.

## Qué demuestra

- **`plugin.yaml` completo**: `contract_version`, `language`, `api`
  (con `openapi.yaml` real en `api/`), `permissions` declarados de verdad
  (`filesystem: ["./data"]` — es lo único que este plugin toca), `resources`
  (`files`, con las 5 operaciones CRUD) y `actions` (`wipe`).
- **`config_schema`**: un campo opcional (`data_dir`) con default, para
  mostrar cómo se lee con `pdk.ConfigString`.
- **Uso del PDK**: `pdk.NewLogger`, `pdk.HealthHandler`, `pdk.WriteError`,
  `pdk.Config()` — nada de esto es obligatorio para cumplir el contrato,
  pero es el camino más corto en Go.
- **Tests reales** (`main_test.go`): crean, leen, actualizan, borran y
  vacían archivos contra un directorio temporal — corren con `go test ./...`,
  sin nada externo.
- **Manejo de input no confiable**: `safeName` rechaza nombres con path
  traversal (`../`) — un recordatorio de que "el storage es solo texto
  plano" no exime de tratar el input como si viniera de afuera, porque
  viene de afuera.

## Probarlo

```
go build -o dummy-fs-provider .
go test ./...

asterion plugin validate .        # valida plugin.yaml + openapi.yaml + el schema JSON de files
asterion plugin install .         # instálalo desde este path local
asterion plugin start dummy-fs-provider
asterion plugin dev dummy-fs-provider   # confirma que la API real responde lo que plugin.yaml declara
```
