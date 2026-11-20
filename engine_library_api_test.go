package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestEngineLibraryReadAndConsent(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "engines", "bedrock"), 0700); err != nil {
		t.Fatal(err)
	}
	m := &manager{root: root, token: "unit-test"}
	authorized := httptest.NewRequest(http.MethodGet, "/api/engines", nil)
	authorized.Header.Set("X-Local-Key", "unit-test")
	response := httptest.NewRecorder()
	m.api(response, authorized)
	if response.Code != http.StatusOK {
		t.Fatalf("consulta de motores HTTP %d: %s", response.Code, response.Body.String())
	}
	unauthed := httptest.NewRecorder()
	m.api(unauthed, httptest.NewRequest(http.MethodGet, "/api/engines", nil))
	if unauthed.Code != http.StatusForbidden {
		t.Fatalf("endpoint aceitou visitante sem chave: %d", unauthed.Code)
	}
	prepare := httptest.NewRequest(http.MethodPost, "/api/engines/prepare?kind=paper", nil)
	prepare.Header.Set("X-Local-Key", "unit-test")
	denied := httptest.NewRecorder()
	m.api(denied, prepare)
	if denied.Code != http.StatusBadRequest {
		t.Fatalf("download sem aceite autorizado: %d", denied.Code)
	}
}
