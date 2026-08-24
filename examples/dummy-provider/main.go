// Command dummy-fs-provider es el "Plugin de Referencia" del Asterion
// Plugin Contract: en vez de aprovisionar servidores reales (que costarían
// dinero de verdad solo para probar el contrato), aprovisiona archivos de
// texto en un directorio local. Implementa el resource "files" completo
// (create/read/update/delete/list) y la action "wipe" tal como los declara
// plugin.yaml — es el ejemplo que se linkea desde toda la documentación
// para mostrar cómo se ve un plugin real y funcionando.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"github.com/Tarafagat/asterion-plugin-contract/sdk/go/pdk"
)

type fileResource struct {
	Name    string `json:"name"`
	Content string `json:"content"`
	Size    int    `json:"size"`
}

// safeName evita path traversal (../, /) — un plugin de verdad recibe
// input de afuera y tiene que tratarlo como no confiable, aunque el
// "storage" real acá sean archivos de texto sin ninguna consecuencia grave.
var safeName = regexp.MustCompile(`^[a-zA-Z0-9_.-]{1,128}$`)

type server struct {
	dataDir string
	logger  *pdk.Logger
}

func main() {
	cfg := pdk.Config()
	dataDir := pdk.ConfigString(cfg, "data_dir", "./data")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		log.Fatalf("no pude crear data_dir %s: %v", dataDir, err)
	}

	s := &server{dataDir: dataDir, logger: pdk.NewLogger(pdk.Name())}

	port := pdk.Port()
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", pdk.HealthHandler(s.health))
	mux.HandleFunc("GET /api/v1/files", s.list)
	mux.HandleFunc("POST /api/v1/files", s.create)
	mux.HandleFunc("GET /api/v1/files/{name}", s.read)
	mux.HandleFunc("PUT /api/v1/files/{name}", s.update)
	mux.HandleFunc("DELETE /api/v1/files/{name}", s.delete)
	mux.HandleFunc("POST /api/v1/files/wipe", s.wipe)

	s.logger.Info("arrancando", map[string]any{"port": port, "data_dir": dataDir})
	if err := http.ListenAndServe("127.0.0.1:"+port, mux); err != nil {
		log.Fatal(err)
	}
}

func (s *server) health() (pdk.Status, string) {
	if _, err := os.Stat(s.dataDir); err != nil {
		return pdk.Unhealthy, "no puedo acceder a data_dir: " + err.Error()
	}
	return pdk.Healthy, ""
}

func (s *server) path(name string) string {
	return filepath.Join(s.dataDir, name)
}

func (s *server) list(w http.ResponseWriter, r *http.Request) {
	entries, err := os.ReadDir(s.dataDir)
	if err != nil {
		pdk.WriteError(w, http.StatusInternalServerError, "no pude leer data_dir", err.Error())
		return
	}
	out := make([]fileResource, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		out = append(out, fileResource{Name: e.Name(), Size: int(info.Size())})
	}
	writeJSON(w, out)
}

func (s *server) create(w http.ResponseWriter, r *http.Request) {
	var in fileResource
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		pdk.WriteError(w, http.StatusBadRequest, "cuerpo inválido", err.Error())
		return
	}
	if !safeName.MatchString(in.Name) {
		pdk.WriteError(w, http.StatusBadRequest, "'name' inválido", "solo letras, números, punto, guion y guión bajo")
		return
	}
	if err := os.WriteFile(s.path(in.Name), []byte(in.Content), 0o644); err != nil {
		pdk.WriteError(w, http.StatusInternalServerError, "no pude escribir el archivo", err.Error())
		return
	}
	s.logger.Info("file.created", map[string]any{"name": in.Name})
	in.Size = len(in.Content)
	w.WriteHeader(http.StatusCreated)
	writeJSON(w, in)
}

func (s *server) read(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	data, err := os.ReadFile(s.path(name))
	if err != nil {
		pdk.WriteError(w, http.StatusNotFound, fmt.Sprintf("no existe %q", name), "")
		return
	}
	writeJSON(w, fileResource{Name: name, Content: string(data), Size: len(data)})
}

func (s *server) update(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if !safeName.MatchString(name) {
		pdk.WriteError(w, http.StatusBadRequest, "'name' inválido", "")
		return
	}
	var in fileResource
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		pdk.WriteError(w, http.StatusBadRequest, "cuerpo inválido", err.Error())
		return
	}
	if _, err := os.Stat(s.path(name)); err != nil {
		pdk.WriteError(w, http.StatusNotFound, fmt.Sprintf("no existe %q", name), "")
		return
	}
	if err := os.WriteFile(s.path(name), []byte(in.Content), 0o644); err != nil {
		pdk.WriteError(w, http.StatusInternalServerError, "no pude escribir el archivo", err.Error())
		return
	}
	s.logger.Info("file.updated", map[string]any{"name": name})
	writeJSON(w, fileResource{Name: name, Content: in.Content, Size: len(in.Content)})
}

func (s *server) delete(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if err := os.Remove(s.path(name)); err != nil {
		pdk.WriteError(w, http.StatusNotFound, fmt.Sprintf("no existe %q", name), "")
		return
	}
	s.logger.Info("file.deleted", map[string]any{"name": name})
	w.WriteHeader(http.StatusNoContent)
}

func (s *server) wipe(w http.ResponseWriter, r *http.Request) {
	entries, err := os.ReadDir(s.dataDir)
	if err != nil {
		pdk.WriteError(w, http.StatusInternalServerError, "no pude leer data_dir", err.Error())
		return
	}
	removed := 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if err := os.Remove(s.path(e.Name())); err == nil {
			removed++
		}
	}
	s.logger.Info("files.wiped", map[string]any{"removed": removed})
	writeJSON(w, map[string]int{"removed": removed})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
