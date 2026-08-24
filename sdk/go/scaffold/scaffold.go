// Package scaffold genera la estructura mínima de un plugin en Go que ya
// cumple el Asterion Plugin Contract — lo que corre `asterion plugin init`.
// La comunidad no debería tener que armar plugin.yaml ni el wiring de
// health check a mano; este paquete resuelve eso una sola vez.
package scaffold

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/Tarafagat/asterion-plugin-contract/apc"
)

//go:embed templates/*.tmpl
var templates embed.FS

// Options son los datos que rellenan las plantillas.
type Options struct {
	Name        string
	Description string
	Author      string
}

// outputs mapea cada plantilla a su ruta de salida relativa al directorio
// del plugin nuevo.
var outputs = map[string]string{
	"plugin.yaml.tmpl": "plugin.yaml",
	"main.go.tmpl":     "main.go",
	"go.mod.tmpl":      "go.mod",
	"README.md.tmpl":   "README.md",
	"gitignore.tmpl":   ".gitignore",
}

// Generate escribe el scaffold completo en dir (que se crea si no existe).
// Falla si dir ya tiene un plugin.yaml — Generate nunca pisa un plugin
// existente.
func Generate(dir string, opts Options) error {
	if opts.Name == "" {
		return fmt.Errorf("scaffold: falta el nombre del plugin")
	}
	if !apc.IsValidName(opts.Name) {
		return fmt.Errorf("scaffold: %q no es un nombre de plugin válido — solo minúsculas, números, guiones y guión bajo", opts.Name)
	}
	if opts.Description == "" {
		opts.Description = fmt.Sprintf("Plugin de Asterion: %s", opts.Name)
	}
	if opts.Author == "" {
		opts.Author = "Tu nombre"
	}

	if _, err := os.Stat(filepath.Join(dir, "plugin.yaml")); err == nil {
		return fmt.Errorf("scaffold: %s ya tiene un plugin.yaml — no lo piso", dir)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	for tmplName, outRel := range outputs {
		if err := renderOne(dir, tmplName, outRel, opts); err != nil {
			return err
		}
	}
	return nil
}

func renderOne(dir, tmplName, outRel string, opts Options) error {
	tmpl, err := template.ParseFS(templates, "templates/"+tmplName)
	if err != nil {
		return fmt.Errorf("scaffold: plantilla %s: %w", tmplName, err)
	}
	outPath := filepath.Join(dir, outRel)
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}
	f, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("scaffold: no pude crear %s: %w", outPath, err)
	}
	defer f.Close()
	if err := tmpl.Execute(f, opts); err != nil {
		return fmt.Errorf("scaffold: no pude generar %s: %w", outPath, err)
	}
	return nil
}
