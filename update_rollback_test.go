package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReplaceExecutableWithRollbackSuccess(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "app.exe")
	source := filepath.Join(dir, "new.exe")
	if err := os.WriteFile(target, []byte("original"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(source, []byte("new version"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := replaceExecutableWithRollback(source, target); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "new version" {
		t.Fatalf("versao instalada incorreta: %s", data)
	}
	matches, err := filepath.Glob(target + ".previous-update-*")
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 1 {
		t.Fatalf("esperava 1 backup recuperavel: %v", matches)
	}
	old, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatal(err)
	}
	if string(old) != "original" {
		t.Fatal("backup da versao anterior foi alterado")
	}
}

func TestReplaceExecutableWithRollbackFailure(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "app.exe")
	if err := os.WriteFile(target, []byte("original"), 0600); err != nil {
		t.Fatal(err)
	}
	err := replaceExecutableWithRollback(filepath.Join(dir, "missing.exe"), target)
	if err == nil {
		t.Fatal("substituicao deveria falhar quando a origem esta ausente")
	}
	data, readErr := os.ReadFile(target)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(data) != "original" {
		t.Fatalf("arquivo original perdido: %s", data)
	}
	matches, globErr := filepath.Glob(target + ".previous-update-*")
	if globErr != nil {
		t.Fatal(globErr)
	}
	if len(matches) != 0 {
		t.Fatalf("rollback deixou copias desnecessarias: %s", strings.Join(matches, ", "))
	}
}
