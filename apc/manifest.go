// Package apc es el Asterion Plugin Contract: la definición canónica de
// qué es un plugin.yaml válido y qué necesita declarar un plugin para que
// Asterion pueda instalarlo, arrancarlo, validarlo y administrarlo — sin
// importar en qué lenguaje esté escrito. asterion-core no reimplementa
// nada de esto: internal/plugins.Manifest es un alias directo de apc.Manifest
// (mismo criterio que asterion-lab con LabState/store.LabState), así que
// esta es la única definición del contrato en todo el ecosistema.
//
// El contrato extiende, no reemplaza, el plugin.yaml que ya existía: name,
// version, start, port, health_path y config_schema siguen significando
// exactamente lo mismo. Todo lo nuevo (contract_version, language, api,
// permissions, resources, actions, events) es opcional — un plugin.yaml de
// antes de que este paquete existiera sigue siendo válido hoy.
package apc

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// ContractVersion es la versión del Asterion Plugin Contract que entiende
// este código. Un plugin.yaml que no declara contract_version se asume en
// esta versión (compatibilidad con manifiestos pre-APC); uno que declara
// una versión que no está en supportedContractVersions se rechaza en vez de
// asumir que es compatible.
const ContractVersion = "asterion.plugin/v1"

var supportedContractVersions = map[string]bool{
	ContractVersion: true,
}

var namePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{1,63}$`)

// IsValidName confirma si name cumple las reglas de identidad de un plugin
// (usado tanto por Manifest.Validate() como por internal/plugins al derivar
// un nombre de una URL de repo en el momento de instalar).
func IsValidName(name string) bool {
	return namePattern.MatchString(name)
}

// StartSpec es el comando que levanta el proceso del plugin. Command debe
// ser un binario resoluble en PATH — Asterion nunca interpreta un script,
// solo lo ejecuta vía exec.Command.
type StartSpec struct {
	Command string   `yaml:"command" json:"command"`
	Args    []string `yaml:"args,omitempty" json:"args,omitempty"`
}

// ConfigField describe un dato que el plugin necesita para funcionar.
// frontend-core renderiza un formulario genérico a partir de esta lista.
type ConfigField struct {
	Key      string `yaml:"key" json:"key"`
	Label    string `yaml:"label" json:"label"`
	Type     string `yaml:"type" json:"type"` // string | number | bool
	Secret   bool   `yaml:"secret,omitempty" json:"secret,omitempty"`
	Required bool   `yaml:"required,omitempty" json:"required,omitempty"`
	Default  string `yaml:"default,omitempty" json:"default,omitempty"`
}

// LanguageSpec identifica en qué está escrito el plugin — el dashboard y
// el marketplace lo muestran, y asterion-core lo usa para decidir CÓMO
// prepararlo (`asterion plugin build`/`plugin start --build`/`plugin
// system apply --build`): Name="go" compila con `go build`; Name="python"
// sincroniza (o crea, si hace falta) un virtualenv e instala
// Requirements ahí. Sigue sin decidir CÓMO EJECUTARLO — eso lo sigue
// diciendo siempre start.command, sin excepción.
//
// Venv/Requirements son opcionales — sin declararlos, un plugin Python se
// sigue infiriendo por convención a partir de start.command (si es
// "./backend/venv/bin/python", el venv es "backend/venv" y
// requirements.txt vive en "backend/requirements.txt", como ya hace
// asterion-sii) — declararlos explícito es para cuando el propio layout
// del plugin no seas esa convención.
type LanguageSpec struct {
	Name    string `yaml:"name,omitempty" json:"name,omitempty"`
	Version string `yaml:"version,omitempty" json:"version,omitempty"`
	// Venv es la ruta (relativa a la raíz del plugin) donde crear/buscar
	// el virtualenv de Python — ej. "backend/venv". Solo tiene efecto si
	// Name="python"; en cualquier otro lenguaje se ignora.
	Venv string `yaml:"venv,omitempty" json:"venv,omitempty"`
	// Requirements es la ruta (relativa a la raíz del plugin) al
	// requirements.txt a instalar en Venv — ej. "backend/requirements.txt".
	// Solo tiene efecto junto con Venv/Name="python".
	Requirements string `yaml:"requirements,omitempty" json:"requirements,omitempty"`
}

// APISpec describe la API HTTP que expone el plugin. OpenAPI es una ruta
// relativa opcional a un archivo dentro del propio repo del plugin — si se
// declara, ValidateDir confirma que existe y que es YAML parseable (no que
// sea un OpenAPI 3 100% válido; eso queda para una versión futura).
type APISpec struct {
	BasePath string `yaml:"base_path,omitempty" json:"base_path,omitempty"`
	OpenAPI  string `yaml:"openapi,omitempty" json:"openapi,omitempty"`
}

// PermissionsSpec es lo que el plugin declara que necesita. En v1 es
// puramente declarativo — Asterion se lo muestra al usuario al instalar
// (consentimiento informado, como los permisos de una app), pero no lo
// hace cumplir técnicamente todavía: eso requiere correr el plugin dentro
// de un contenedor/VM de Asterion Lab en vez de como proceso raw, que es
// un cambio de arquitectura aparte. No fingir enforcement que no existe es
// a propósito — mismo criterio que ya documenta el README de asterion-core
// sobre la falta de sandboxing.
type PermissionsSpec struct {
	Network    []string `yaml:"network,omitempty" json:"network,omitempty"`
	Filesystem []string `yaml:"filesystem,omitempty" json:"filesystem,omitempty"`
	Database   bool     `yaml:"database,omitempty" json:"database,omitempty"`
	Secrets    bool     `yaml:"secrets,omitempty" json:"secrets,omitempty"`
}

// ResourceSpec declara un recurso que Asterion puede administrar en nombre
// del plugin. CRUD es el subconjunto de operaciones que el endpoint
// realmente soporta (create|read|update|delete|list) — el dashboard solo
// ofrece los botones que el plugin declaró que sabe atender.
type ResourceSpec struct {
	Name       string   `yaml:"name" json:"name"`
	Endpoint   string   `yaml:"endpoint" json:"endpoint"`
	Schema     string   `yaml:"schema,omitempty" json:"schema,omitempty"`
	PrimaryKey string   `yaml:"primary_key,omitempty" json:"primary_key,omitempty"`
	CRUD       []string `yaml:"crud,omitempty" json:"crud,omitempty"`
}

// ActionSpec declara una operación que no encaja en CRUD (ej. "emitir
// factura", "reiniciar servicio").
type ActionSpec struct {
	Name        string `yaml:"name" json:"name"`
	Method      string `yaml:"method" json:"method"`
	Endpoint    string `yaml:"endpoint" json:"endpoint"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
}

// EventsSpec es, en v1, puramente declarativo: documenta qué eventos
// publica/consume el plugin, pero no existe todavía un bus de eventos real
// en Asterion que los transporte. Se incluye desde ya en el contrato para
// que los manifiestos no tengan que cambiar de forma cuando ese bus exista
// — solo dejará de estar vacío de comportamiento.
type EventsSpec struct {
	Publishes  []string `yaml:"publishes,omitempty" json:"publishes,omitempty"`
	Subscribes []string `yaml:"subscribes,omitempty" json:"subscribes,omitempty"`
}

// Manifest es plugin.yaml completo. Port en 0 significa "que Asterion elija
// un puerto libre". HealthPath es lo que Asterion consulta después de
// arrancar el proceso para confirmar que levantó de verdad.
type Manifest struct {
	Name        string `yaml:"name" json:"name"`
	Version     string `yaml:"version" json:"version"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	Author      string `yaml:"author,omitempty" json:"author,omitempty"`
	License     string `yaml:"license,omitempty" json:"license,omitempty"`
	Repo        string `yaml:"repo,omitempty" json:"repo,omitempty"`

	// ContractVersion identifica qué versión del Asterion Plugin Contract
	// implementa este manifiesto. Se defaultea a ContractVersion si no se
	// declara — así un plugin.yaml de antes de que existiera este campo
	// sigue siendo válido.
	ContractVersion string `yaml:"contract_version,omitempty" json:"contract_version,omitempty"`

	Language *LanguageSpec `yaml:"language,omitempty" json:"language,omitempty"`

	Start        StartSpec     `yaml:"start" json:"start"`
	Port         int           `yaml:"port" json:"port"`
	HealthPath   string        `yaml:"health_path,omitempty" json:"health_path,omitempty"`
	ConfigSchema []ConfigField `yaml:"config_schema,omitempty" json:"config_schema,omitempty"`

	API         *APISpec         `yaml:"api,omitempty" json:"api,omitempty"`
	Permissions *PermissionsSpec `yaml:"permissions,omitempty" json:"permissions,omitempty"`
	Resources   []ResourceSpec   `yaml:"resources,omitempty" json:"resources,omitempty"`
	Actions     []ActionSpec     `yaml:"actions,omitempty" json:"actions,omitempty"`
	Events      *EventsSpec      `yaml:"events,omitempty" json:"events,omitempty"`
}

var crudOps = map[string]bool{"create": true, "read": true, "update": true, "delete": true, "list": true}
var httpMethods = map[string]bool{"GET": true, "POST": true, "PUT": true, "PATCH": true, "DELETE": true}

// applyDefaults completa los campos opcionales que tienen un valor por
// defecto conocido. Se llama una sola vez, en LoadManifest, antes de
// Validate — así el manifiesto que termina guardado en state.json ya tiene
// todo resuelto, y el resto del código (process.go, etc.) puede leer los
// campos directo sin repetir la lógica de default en cada lugar que los usa.
func (m *Manifest) applyDefaults() {
	if m.ContractVersion == "" {
		m.ContractVersion = ContractVersion
	}
	if m.HealthPath == "" {
		m.HealthPath = "/health"
	}
}

// Validate confirma que el manifiesto tiene lo mínimo para poder operarlo.
// Es una validación puramente estructural (no toca disco) — corta en el
// primer error, mensajes pensados para leerse tal cual en una terminal.
func (m Manifest) Validate() error {
	if !namePattern.MatchString(m.Name) {
		return fmt.Errorf("plugin.yaml: 'name' inválido (%q) — solo minúsculas, números, guiones y guión bajo, 2-64 caracteres", m.Name)
	}
	if m.Version == "" {
		return fmt.Errorf("plugin.yaml: falta 'version'")
	}
	if m.Start.Command == "" {
		return fmt.Errorf("plugin.yaml: falta 'start.command' — cómo arrancar el proceso del plugin")
	}
	if m.ContractVersion != "" && !supportedContractVersions[m.ContractVersion] {
		return fmt.Errorf("plugin.yaml: contract_version %q no soportada — esta versión de Asterion entiende %q", m.ContractVersion, ContractVersion)
	}
	for _, f := range m.ConfigSchema {
		if f.Key == "" {
			return fmt.Errorf("plugin.yaml: config_schema tiene un campo sin 'key'")
		}
	}

	seenResources := map[string]bool{}
	for _, r := range m.Resources {
		if r.Name == "" {
			return fmt.Errorf("plugin.yaml: resources tiene un elemento sin 'name'")
		}
		if seenResources[r.Name] {
			return fmt.Errorf("plugin.yaml: hay dos resources llamados %q — los nombres deben ser únicos", r.Name)
		}
		seenResources[r.Name] = true
		if r.Endpoint == "" {
			return fmt.Errorf("resource %q: falta 'endpoint'", r.Name)
		}
		for _, op := range r.CRUD {
			if !crudOps[op] {
				return fmt.Errorf("resource %q: crud %q no reconocido (create|read|update|delete|list)", r.Name, op)
			}
		}
	}

	seenActions := map[string]bool{}
	for _, a := range m.Actions {
		if a.Name == "" {
			return fmt.Errorf("plugin.yaml: actions tiene un elemento sin 'name'")
		}
		if seenActions[a.Name] {
			return fmt.Errorf("plugin.yaml: hay dos actions llamadas %q — los nombres deben ser únicos", a.Name)
		}
		seenActions[a.Name] = true
		if a.Endpoint == "" {
			return fmt.Errorf("action %q: falta 'endpoint'", a.Name)
		}
		if !httpMethods[strings.ToUpper(a.Method)] {
			return fmt.Errorf("action %q: method %q no soportado (GET|POST|PUT|PATCH|DELETE)", a.Name, a.Method)
		}
	}

	return nil
}

// LoadManifest lee, defaultea y valida plugin.yaml en la raíz de dir.
func LoadManifest(dir string) (Manifest, error) {
	path := filepath.Join(dir, "plugin.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return Manifest{}, fmt.Errorf("no encontré %s — todo plugin de Asterion necesita un plugin.yaml en la raíz del repo: %w", path, err)
	}
	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return Manifest{}, fmt.Errorf("%s no es YAML válido: %w", path, err)
	}
	m.applyDefaults()
	if err := m.Validate(); err != nil {
		return Manifest{}, err
	}
	return m, nil
}
