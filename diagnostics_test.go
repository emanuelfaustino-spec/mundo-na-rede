package main

import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestProbeLocalPortDetectsBusyPort(t *testing.T) {
	listener, err := net.Listen("tcp4", ":0")
	if err != nil {
		t.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	if err := probeLocalPort("tcp4", port); err == nil {
		t.Fatal("porta ocupada foi classificada como livre")
	}
	listener.Close()
	if err := probeLocalPort("tcp4", port); err != nil {
		t.Fatalf("porta liberada continua ocupada: %v", err)
	}
}

func TestPreflightMissingFiles(t *testing.T) {
	m := &manager{running: map[string]*process{}}
	p := Profile{ID: "1111222233334444", Engine: "bedrock", Port: 19132}
	r := m.preflight(p, filepath.Join(t.TempDir(), "missing"))
	if r.Ready {
		t.Fatal("arquivo do mundo ausente deveria impedir inicializacao")
	}
	errorsFound := 0
	for _, check := range r.Checks {
		if check.Level == "error" {
			errorsFound++
		}
	}
	if errorsFound < 2 {
		t.Fatalf("esperava erros de mundo e motor, obteve: %+v", r.Checks)
	}
}

func TestPreflightAPIRequiresLocalKey(t *testing.T) {
	root := t.TempDir()
	id := "1111222233334444"
	dir := filepath.Join(root, "servers", id)
	if err := os.MkdirAll(filepath.Join(dir, "game"), 0700); err != nil {
		t.Fatal(err)
	}
	p := Profile{ID: id, Engine: "bedrock", Port: 19132}
	raw, _ := json.Marshal(p)
	if err := os.WriteFile(filepath.Join(dir, "profile.json"), raw, 0600); err != nil {
		t.Fatal(err)
	}
	binary := "bedrock_server"
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	if err := os.WriteFile(filepath.Join(dir, "game", binary), []byte("test"), 0700); err != nil {
		t.Fatal(err)
	}
	m := &manager{root: root, token: "testtoken", running: map[string]*process{}}
	url := "/api/servers/" + id + "/preflight"
	denied := httptest.NewRecorder()
	m.api(denied, httptest.NewRequest(http.MethodGet, url, nil))
	if denied.Code != http.StatusForbidden {
		t.Fatalf("esperava 403, obteve %d", denied.Code)
	}
	request := httptest.NewRequest(http.MethodGet, url, nil)
	request.Header.Set("X-Local-Key", "testtoken")
	good := httptest.NewRecorder()
	m.api(good, request)
	if good.Code != http.StatusOK {
		t.Fatalf("esperava 200, obteve %d, corpo: %s", good.Code, good.Body.String())
	}
	var result preflightReport
	if err := json.Unmarshal(good.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if len(result.Checks) == 0 {
		t.Fatal("endpoint nao retornou verificacoes")
	}
}
