package apc

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ValidateDir hace lo mismo que LoadManifest (leer + defaultear + validar
// estructuralmente plugin.yaml) y además confirma que los archivos que el
// manifiesto referencia por ruta relativa (api.openapi, resources[].schema)
// existen de verdad y son al menos parseables en su formato — sin llegar a
// validar un OpenAPI 3 o un JSON Schema completos, que queda para una
// versión futura del validador. Es lo que corre `asterion plugin validate`.
func ValidateDir(dir string) (Manifest, error) {
	m, err := LoadManifest(dir)
	if err != nil {
		return Manifest{}, err
	}

	if m.API != nil && m.API.OpenAPI != "" {
		p := filepath.Join(dir, m.API.OpenAPI)
		data, err := os.ReadFile(p)
		if err != nil {
			return Manifest{}, fmt.Errorf("plugin.yaml declara api.openapi=%q pero no pude leer %s: %w", m.API.OpenAPI, p, err)
		}
		var probe any
		if err := yaml.Unmarshal(data, &probe); err != nil {
			return Manifest{}, fmt.Errorf("%s no es YAML válido: %w", p, err)
		}
	}

	for _, r := range m.Resources {
		if r.Schema == "" {
			continue
		}
		p := filepath.Join(dir, r.Schema)
		data, err := os.ReadFile(p)
		if err != nil {
			return Manifest{}, fmt.Errorf("resource %q declara schema=%q pero no pude leer %s: %w", r.Name, r.Schema, p, err)
		}
		var probe any
		if err := json.Unmarshal(data, &probe); err != nil {
			return Manifest{}, fmt.Errorf("%s no es JSON válido: %w", p, err)
		}
	}

	return m, nil
}
