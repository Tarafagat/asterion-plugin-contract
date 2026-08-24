// Package openapi implementa el "transformador de API en plugin": dado un
// openapi.yaml de una API que ya existe, infiere un plugin.yaml de partida
// (resources + actions) para que crear un plugin de Asterion a partir de
// una API propia sea copiar y pegar en vez de escribir todo a mano.
//
// Es deliberadamente heurístico, no un parser OpenAPI 3 completo: agrupa
// paths por su primer segmento, y distingue "colección" (/x → list/create),
// "miembro" (/x/{id} → read/update/delete) de "acción" (cualquier otra
// forma, ej. /x/{id}/emitir). Cubre el caso común razonablemente bien; lo
// que no adivine bien, el autor lo ajusta a mano en el plugin.yaml
// resultante — el objetivo es ahorrar tecleo, no reemplazar criterio.
package openapi

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/Tarafagat/asterion-plugin-contract/apc"
)

var httpVerbs = []string{"get", "post", "put", "patch", "delete"}

type doc struct {
	Info struct {
		Title string `yaml:"title"`
	} `yaml:"info"`
	Paths map[string]map[string]operation `yaml:"paths"`
}

type operation struct {
	OperationID string `yaml:"operationId"`
	Summary     string `yaml:"summary"`
}

type pathKind int

const (
	kindAction pathKind = iota
	kindCollection
	kindMember
)

func isParam(seg string) bool {
	return strings.HasPrefix(seg, "{") && strings.HasSuffix(seg, "}")
}

func splitPath(path string) []string {
	trimmed := strings.Trim(path, "/")
	if trimmed == "" {
		return nil
	}
	return strings.Split(trimmed, "/")
}

func classify(segs []string) pathKind {
	switch {
	case len(segs) == 1 && !isParam(segs[0]):
		return kindCollection
	case len(segs) == 2 && !isParam(segs[0]) && isParam(segs[1]):
		return kindMember
	default:
		return kindAction
	}
}

var nonSlug = regexp.MustCompile(`[^a-z0-9_-]+`)
var repeatedDash = regexp.MustCompile(`-{2,}`)

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "-")
	// Cualquier carácter fuera de a-z0-9_- (acentos, ñ, etc.) se vuelve
	// separador, no se borra — "Facturación" tiene que dar "facturaci-n" o
	// "facturacion", nunca "facturacin" (que junta mal dos letras).
	s = nonSlug.ReplaceAllString(s, "-")
	s = repeatedDash.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-_")
	if s == "" {
		return "mi-plugin"
	}
	if len(s) > 64 {
		s = s[:64]
	}
	return s
}

func addCRUD(r *apc.ResourceSpec, op string) {
	for _, existing := range r.CRUD {
		if existing == op {
			return
		}
	}
	r.CRUD = append(r.CRUD, op)
}

func actionName(op operation, segs []string, verb string) string {
	if op.OperationID != "" {
		return slugify(op.OperationID)
	}
	var static []string
	for _, s := range segs {
		if !isParam(s) {
			static = append(static, s)
		}
	}
	return slugify(verb + "-" + strings.Join(static, "-"))
}

// Infer lee un openapi.yaml y devuelve un Manifest de partida.
func Infer(openapiPath string) (apc.Manifest, error) {
	data, err := os.ReadFile(openapiPath)
	if err != nil {
		return apc.Manifest{}, fmt.Errorf("no pude leer %s: %w", openapiPath, err)
	}
	var d doc
	if err := yaml.Unmarshal(data, &d); err != nil {
		return apc.Manifest{}, fmt.Errorf("%s no es YAML válido: %w", openapiPath, err)
	}
	if len(d.Paths) == 0 {
		return apc.Manifest{}, fmt.Errorf("%s no declara ningún path bajo 'paths' — no hay nada que inferir", openapiPath)
	}

	resources := map[string]*apc.ResourceSpec{}
	var resourceOrder []string
	var actions []apc.ActionSpec

	for path, methods := range d.Paths {
		segs := splitPath(path)
		if len(segs) == 0 {
			continue
		}
		kind := classify(segs)
		if kind == kindAction {
			for _, verb := range httpVerbs {
				op, ok := methods[verb]
				if !ok {
					continue
				}
				actions = append(actions, apc.ActionSpec{
					Name:        actionName(op, segs, verb),
					Method:      strings.ToUpper(verb),
					Endpoint:    path,
					Description: op.Summary,
				})
			}
			continue
		}

		base := segs[0]
		r, ok := resources[base]
		if !ok {
			r = &apc.ResourceSpec{Name: slugify(base), Endpoint: "/" + base, PrimaryKey: "id"}
			resources[base] = r
			resourceOrder = append(resourceOrder, base)
		}
		switch kind {
		case kindCollection:
			if _, ok := methods["get"]; ok {
				addCRUD(r, "list")
			}
			if _, ok := methods["post"]; ok {
				addCRUD(r, "create")
			}
		case kindMember:
			if _, ok := methods["get"]; ok {
				addCRUD(r, "read")
			}
			if _, ok := methods["put"]; ok {
				addCRUD(r, "update")
			}
			if _, ok := methods["patch"]; ok {
				addCRUD(r, "update")
			}
			if _, ok := methods["delete"]; ok {
				addCRUD(r, "delete")
			}
		}
	}

	sort.Strings(resourceOrder)
	resourceList := make([]apc.ResourceSpec, 0, len(resourceOrder))
	for _, base := range resourceOrder {
		resourceList = append(resourceList, *resources[base])
	}
	sort.Slice(actions, func(i, j int) bool { return actions[i].Name < actions[j].Name })

	name := slugify(d.Info.Title)
	return apc.Manifest{
		Name:            name,
		Version:         "0.1.0",
		Description:     d.Info.Title,
		ContractVersion: apc.ContractVersion,
		Start:           apc.StartSpec{Command: "./" + name},
		HealthPath:      "/health",
		Resources:       resourceList,
		Actions:         actions,
	}, nil
}
