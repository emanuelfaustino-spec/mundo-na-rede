package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLocalPanelKeyPersistsAndRecovers(t *testing.T) {
	root := t.TempDir()
	first, err := localPanelKey(root)
	if err != nil {
		t.Fatal(err)
	}
	second, err := localPanelKey(root)
	if err != nil {
		t.Fatal(err)
	}
	if first != second || len(first) != 64 {
		t.Fatalf("session key is not stable: %q / %q", first, second)
	}
	path := filepath.Join(root, "panel-session.key")
	if err := os.WriteFile(path, []byte("broken"), 0600); err != nil {
		t.Fatal(err)
	}
	if repaired, err := localPanelKey(root); err != nil || len(repaired) != 64 || repaired == "broken" {
		t.Fatalf("invalid key was not recovered: %q, %v", repaired, err)
	}
}
