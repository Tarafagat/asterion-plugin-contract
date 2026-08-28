package pdk

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Tarafagat/asterion-plugin-contract/apc"
)

func sampleManifest() apc.Manifest {
	return apc.Manifest{
		Name:            "sample-rest-plugin",
		Version:         "1.0.0",
		Description:     "Un plugin puramente REST, sin frontend propio.",
		Author:          "Alguien",
		License:         "MIT",
		ContractVersion: apc.ContractVersion,
		HealthPath:      "/health",
		ConfigSchema: []apc.ConfigField{
			{Key: "api_key", Label: "API key", Type: "string", Required: true, Secret: true},
		},
		API: &apc.APISpec{BasePath: "/api/v1"},
		Permissions: &apc.PermissionsSpec{
			Network: []string{"internet"},
		},
		Resources: []apc.ResourceSpec{
			{Name: "widgets", Endpoint: "/widgets", PrimaryKey: "id", CRUD: []string{"list", "create"}},
		},
		Actions: []apc.ActionSpec{
			{Name: "sync", Method: "POST", Endpoint: "/sync", Description: "Sincroniza con el proveedor externo"},
		},
	}
}

func TestDefaultFrontendHandler_RendersManifestContent(t *testing.T) {
	m := sampleManifest()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	DefaultFrontendHandler(m)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	ct := rec.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		t.Fatalf("Content-Type = %q, want text/html", ct)
	}

	body := rec.Body.String()
	wantContains := []string{
		"sample-rest-plugin",       // nombre en el título/header
		"Un plugin puramente REST", // descripción
		"api_key",                  // config schema
		"/api/v1/widgets",          // resource con base_path aplicado
		"GET",                      // crud "list" -> GET
		"POST",                     // crud "create" -> POST
		"/api/v1/sync",             // action con base_path aplicado
		"sync",                     // nombre de la action
	}
	for _, want := range wantContains {
		if !strings.Contains(body, want) {
			t.Errorf("la página generada no contiene %q\n--- body ---\n%s", want, body)
		}
	}
}

func TestDefaultFrontendHandler_EscapesUntrustedManifestFields(t *testing.T) {
	m := sampleManifest()
	m.Description = `<script>alert(1)</script>`
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	DefaultFrontendHandler(m)(rec, req)

	body := rec.Body.String()
	if strings.Contains(body, "<script>alert(1)</script>") {
		t.Fatalf("la descripción del manifest se insertó sin escapar — riesgo de XSS:\n%s", body)
	}
	if !strings.Contains(body, "&lt;script&gt;") {
		t.Fatalf("esperaba la descripción escapada por html/template, no apareció:\n%s", body)
	}
}

// TestMountFrontend_FallsBackWhenNoDist confirma el comportamiento
// completo que van a usar los plugins reales: si distDir no existe (o
// existe vacío), MountFrontend sirve la página generada en vez de un
// FileServer roto.
func TestMountFrontend_FallsBackWhenNoDist(t *testing.T) {
	mux := http.NewServeMux()
	MountFrontend(mux, filepath.Join(t.TempDir(), "no-existe"), sampleManifest())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "sample-rest-plugin") {
		t.Fatalf("esperaba la página generada, salió otra cosa:\n%s", rec.Body.String())
	}
}

// TestMountFrontend_ServesRealDistWhenPresent confirma que un plugin CON
// frontend propio sigue funcionando exactamente igual que antes — el
// fallback nunca debe pisar un build real.
func TestMountFrontend_ServesRealDistWhenPresent(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("<h1>build real</h1>"), 0o644); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	MountFrontend(mux, dir, sampleManifest())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "build real") {
		t.Fatalf("esperaba el index.html real servido tal cual, salió otra cosa:\n%s", rec.Body.String())
	}
}
