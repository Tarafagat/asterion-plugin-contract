package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Tarafagat/asterion-plugin-contract/sdk/go/pdk"
)

func newTestServer(t *testing.T) *server {
	t.Helper()
	return &server{dataDir: t.TempDir(), logger: pdk.NewLogger("test")}
}

func TestCreateReadUpdateDelete(t *testing.T) {
	s := newTestServer(t)

	body, _ := json.Marshal(fileResource{Name: "hola.txt", Content: "mundo"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/files", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	s.create(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create: esperaba 201, obtuve %d (%s)", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/files/hola.txt", nil)
	req.SetPathValue("name", "hola.txt")
	rec = httptest.NewRecorder()
	s.read(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("read: esperaba 200, obtuve %d", rec.Code)
	}
	var got fileResource
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Content != "mundo" {
		t.Fatalf("read: contenido inesperado %q", got.Content)
	}

	body, _ = json.Marshal(fileResource{Content: "mundo actualizado"})
	req = httptest.NewRequest(http.MethodPut, "/api/v1/files/hola.txt", bytes.NewReader(body))
	req.SetPathValue("name", "hola.txt")
	rec = httptest.NewRecorder()
	s.update(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("update: esperaba 200, obtuve %d (%s)", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodDelete, "/api/v1/files/hola.txt", nil)
	req.SetPathValue("name", "hola.txt")
	rec = httptest.NewRecorder()
	s.delete(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("delete: esperaba 204, obtuve %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/files/hola.txt", nil)
	req.SetPathValue("name", "hola.txt")
	rec = httptest.NewRecorder()
	s.read(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("read tras delete: esperaba 404, obtuve %d", rec.Code)
	}
}

func TestCreateRejectsUnsafeName(t *testing.T) {
	s := newTestServer(t)
	body, _ := json.Marshal(fileResource{Name: "../escape", Content: "x"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/files", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	s.create(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("esperaba 400 para un name con path traversal, obtuve %d", rec.Code)
	}
}

func TestWipe(t *testing.T) {
	s := newTestServer(t)
	for _, name := range []string{"a.txt", "b.txt"} {
		body, _ := json.Marshal(fileResource{Name: name, Content: "x"})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/files", bytes.NewReader(body))
		rec := httptest.NewRecorder()
		s.create(rec, req)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/files/wipe", nil)
	rec := httptest.NewRecorder()
	s.wipe(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("wipe: esperaba 200, obtuve %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	s.list(rec, httptest.NewRequest(http.MethodGet, "/api/v1/files", nil))
	var out []fileResource
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if len(out) != 0 {
		t.Fatalf("esperaba 0 archivos tras wipe, quedaron %d", len(out))
	}
}

func TestHealth(t *testing.T) {
	s := newTestServer(t)
	status, detail := s.health()
	if status != pdk.Healthy {
		t.Fatalf("esperaba healthy, obtuve %s (%s)", status, detail)
	}
}
