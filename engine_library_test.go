package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPaperCacheSharedBetweenProfiles(t *testing.T) {
	root := t.TempDir()
	m := &manager{root: root}
	cached := filepath.Join(root, "engines", "paper", "26.3")
	if err := os.MkdirAll(cached, 0700); err != nil {
		t.Fatal(err)
	}
	original := []byte("paper-fixture")
	if err := os.WriteFile(filepath.Join(cached, "server.jar"), original, 0600); err != nil {
		t.Fatal(err)
	}
	if m.cachedPaper() == "" {
		t.Fatal("motor Paper em cache não foi encontrado")
	}
	for _, name := range []string{"alpha", "bravo"} {
		game := filepath.Join(root, name)
		if err := os.MkdirAll(game, 0700); err != nil {
			t.Fatal(err)
		}
		p := Profile{Engine: "java"}
		if err := m.ensureCachedEngine(p, game); err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(filepath.Join(game, "server.jar"))
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != string(original) {
			t.Fatal("motor copiado incorretamente")
		}
	}
}
func TestExistingServerEngineNotOverwritten(t *testing.T) {
	root := t.TempDir()
	game := filepath.Join(root, "server")
	if err := os.MkdirAll(game, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(game, "server.jar"), []byte("custom"), 0600); err != nil {
		t.Fatal(err)
	}
	m := &manager{root: root}
	if err := m.ensureCachedEngine(Profile{Engine: "java"}, game); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(game, "server.jar"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "custom" {
		t.Fatal("motor importado foi alterado")
	}
}
