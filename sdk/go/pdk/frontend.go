package pdk

import (
	"html/template"
	"net/http"
	"os"
	"strings"

	"github.com/Tarafagat/asterion-plugin-contract/apc"
)

// MountFrontend registra en mux el frontend del plugin en "/". Si distDir
// existe y tiene contenido, se sirve tal cual — el build real que el autor
// haya hecho (React, Vue, lo que sea). Si no, el propio Plugin Contract se
// encarga de generar uno: un plugin puramente REST, sin ningún frontend
// propio, es un caso válido y esperado, y no tiene por qué quedar "sin
// cara" solo por no tener un dashboard hecho a mano — ver
// DefaultFrontendHandler.
func MountFrontend(mux *http.ServeMux, distDir string, manifest apc.Manifest) {
	if info, err := os.Stat(distDir); err == nil && info.IsDir() {
		if entries, err := os.ReadDir(distDir); err == nil && len(entries) > 0 {
			mux.Handle("GET /", http.FileServer(http.Dir(distDir)))
			return
		}
	}
	mux.HandleFunc("GET /", DefaultFrontendHandler(manifest))
}

// DefaultFrontendHandler arma, en cada request, una página de solo lectura
// a partir del manifest del propio plugin — sin estado, sin JavaScript,
// nada que ejecutar: es documentación autogenerada de lo que el plugin
// declaró en su plugin.yaml (config que necesita, resources, actions,
// permisos), el mismo contenido que el panel de Asterion ya arma del lado
// del dashboard, servido acá directo por el plugin para quien lo abra
// fuera de ese dashboard (o antes de que el autor decida escribir un
// frontend propio, si es que alguna vez hace falta).
func DefaultFrontendHandler(m apc.Manifest) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := defaultFrontendTemplate.Execute(w, m); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

var crudMethodDisplay = map[string]string{
	"list":   "GET",
	"read":   "GET /{id}",
	"create": "POST",
	"update": "PUT/PATCH /{id}",
	"delete": "DELETE /{id}",
}

var defaultFrontendFuncs = template.FuncMap{
	"crudMethod": func(op string) string {
		if v, ok := crudMethodDisplay[op]; ok {
			return v
		}
		return op
	},
	"join": func(items []string, sep string) string { return strings.Join(items, sep) },
	"basePath": func(m apc.Manifest) string {
		if m.API != nil {
			return m.API.BasePath
		}
		return ""
	},
}

var defaultFrontendTemplate = template.Must(template.New("default-frontend").Funcs(defaultFrontendFuncs).Parse(defaultFrontendHTML))

const defaultFrontendHTML = `<!doctype html>
<html lang="es">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>{{.Name}}</title>
<style>
:root{--void:#0a0e2a;--navy:#111a3d;--navy-light:#182352;--gold:#d4a017;--gold-light:#f0c04a}
*{box-sizing:border-box}
body{margin:0;background:var(--void);color:#f5f6fa;font-family:"Segoe UI",system-ui,-apple-system,sans-serif}
.topbar{display:flex;align-items:center;justify-content:space-between;padding:1rem 1.5rem;border-bottom:1px solid rgba(255,255,255,.08);flex-wrap:wrap;gap:.5rem}
.topbar h1{font-size:1rem;letter-spacing:.08em;text-transform:uppercase;color:var(--gold-light);margin:0}
.badge{font-size:.7rem;padding:.2rem .6rem;border-radius:999px;border:1px solid rgba(212,160,23,.4);color:var(--gold)}
main{max-width:52rem;margin:0 auto;padding:1.5rem;display:grid;gap:1.25rem}
.hint{color:rgba(255,255,255,.55);font-size:.85rem;line-height:1.5}
.card{background:var(--navy-light);border:1px solid rgba(255,255,255,.06);border-radius:1rem;padding:1.5rem}
.section-title{font-size:.8rem;text-transform:uppercase;letter-spacing:.08em;color:var(--gold);margin:0 0 .75rem}
table{width:100%;border-collapse:collapse;font-size:.85rem}
tr{border-top:1px solid rgba(255,255,255,.06)}
tr:first-child{border-top:none}
th,td{text-align:left;padding:.5rem .4rem;vertical-align:top}
thead th{color:var(--gold);font-size:.7rem;text-transform:uppercase;letter-spacing:.05em;border-bottom:1px solid rgba(255,255,255,.1)}
.detail-table th{color:rgba(255,255,255,.5);font-weight:500;width:11rem;white-space:nowrap}
.method{font-family:monospace;font-size:.78rem;color:var(--gold-light);white-space:nowrap}
.path{font-family:monospace;font-size:.78rem;color:rgba(255,255,255,.75);word-break:break-all}
</style>
</head>
<body>
<div class="topbar">
  <h1>{{.Name}}</h1>
  <span class="badge">v{{.Version}} · frontend generado por el Plugin Contract</span>
</div>
<main>
  <p class="hint">
    Este plugin no trae un frontend propio — esta página se generó sola a partir de su plugin.yaml
    (Asterion Plugin Contract {{.ContractVersion}}), documentando qué necesita para configurarse y qué
    endpoints expone. Es de solo lectura: no ejecuta nada.
  </p>

  <div class="card">
    <p class="section-title">Características</p>
    <table class="detail-table">
      <tbody>
        {{if .Description}}<tr><th>Descripción</th><td>{{.Description}}</td></tr>{{end}}
        {{if .Author}}<tr><th>Autor</th><td>{{.Author}}</td></tr>{{end}}
        {{if .License}}<tr><th>Licencia</th><td>{{.License}}</td></tr>{{end}}
        {{if .Language}}<tr><th>Lenguaje</th><td>{{.Language.Name}}{{if .Language.Version}} {{.Language.Version}}{{end}}</td></tr>{{end}}
        <tr><th>Health check</th><td>{{.HealthPath}}</td></tr>
        {{if .Permissions}}
        <tr><th>Permisos declarados</th><td>
          {{if .Permissions.Network}}red: {{join .Permissions.Network ", "}} · {{end}}
          {{if .Permissions.Filesystem}}filesystem: {{join .Permissions.Filesystem ", "}} · {{end}}
          {{if .Permissions.Database}}base de datos · {{end}}
          {{if .Permissions.Secrets}}secretos{{end}}
        </td></tr>
        {{end}}
      </tbody>
    </table>
  </div>

  {{if .ConfigSchema}}
  <div class="card">
    <p class="section-title">Configuración que necesita</p>
    <table>
      <thead><tr><th>Clave</th><th>Etiqueta</th><th>Tipo</th><th>Obligatorio</th><th>Secreto</th></tr></thead>
      <tbody>
        {{range .ConfigSchema}}
        <tr>
          <td class="path">{{.Key}}</td>
          <td>{{.Label}}</td>
          <td>{{.Type}}</td>
          <td>{{if .Required}}sí{{else}}no{{end}}</td>
          <td>{{if .Secret}}sí{{else}}no{{end}}</td>
        </tr>
        {{end}}
      </tbody>
    </table>
  </div>
  {{end}}

  {{if or .Resources .Actions}}
  <div class="card">
    <p class="section-title">Endpoints</p>
    <table>
      <thead><tr><th>Método</th><th>Ruta</th><th>Qué es</th></tr></thead>
      <tbody>
        {{$base := basePath .}}
        {{range .Resources}}
          {{$r := .}}
          {{range .CRUD}}
          <tr>
            <td class="method">{{crudMethod .}}</td>
            <td class="path">{{$base}}{{$r.Endpoint}}</td>
            <td class="hint">recurso <strong>{{$r.Name}}</strong>{{if $r.PrimaryKey}} · clave: {{$r.PrimaryKey}}{{end}}</td>
          </tr>
          {{end}}
        {{end}}
        {{range .Actions}}
        <tr>
          <td class="method">{{.Method}}</td>
          <td class="path">{{$base}}{{.Endpoint}}</td>
          <td class="hint">acción <strong>{{.Name}}</strong>{{if .Description}} — {{.Description}}{{end}}</td>
        </tr>
        {{end}}
      </tbody>
    </table>
  </div>
  {{end}}
</main>
</body>
</html>
`
