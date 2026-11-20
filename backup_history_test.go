package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestBackupNamesAcceptNewAndLegacy(t *testing.T) {
	for _, name := range []string{
		"mundo-na-rede-20261120-101010-abcdef.zip",
		"mundo-nexus-20261120-101010-abcdef.zip",
		"mundo-nexus-manual-20261120-101010-abcdef.zip",
		"mundo-nexus-auto-20261120-101010-abcdef.zip",
	} {
		if !localBackupName.MatchString(name) {
			t.Errorf("nome de backup valido recusado: %s", name)
		}
	}
	for _, name := range []string{
		"../mundo-nexus-20261120-101010-abcdef.zip",
		"mundo-nexus-20261120-101010-abcdef.exe",
		"mundo-nexus-invalid.zip",
	} {
		if localBackupName.MatchString(name) {
			t.Errorf("nome de backup invalido aceito: %s", name)
		}
	}
}

func TestPruneBackupsPreservesLegacyAndNewest(t *testing.T) {
	dir := t.TempDir()
	names := []string{}
	for i := 0; i < 9; i++ {
		prefix := "mundo-nexus-auto-"
		name := prefix + "20261120-101010-" + fmt.Sprintf("%06x", i) + ".zip"
		file := filepath.Join(dir, name)
		if err := os.WriteFile(file, []byte("archive"), 0600); err != nil {
			t.Fatal(err)
		}
		at := time.Now().Add(time.Duration(i) * time.Minute)
		if err := os.Chtimes(file, at, at); err != nil {
			t.Fatal(err)
		}
		names = append(names, name)
	}
	if err := pruneLocalBackups(dir); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != backupRetention {
		t.Fatalf("esperava %d backups, encontrou %d", backupRetention, len(entries))
	}
	for _, name := range names[:2] {
		if _, err := os.Stat(filepath.Join(dir, name)); !os.IsNotExist(err) {
			t.Errorf("backup antigo deveria ser descartado: %s", name)
		}
	}
	for _, name := range names[2:] {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Errorf("backup recente nao foi preservado: %s: %v", name, err)
		}
	}
}

func TestManualBackupsAreNotPruned(t *testing.T) {
	dir := t.TempDir()
	for i := 0; i < 12; i++ {
		name := fmt.Sprintf("mundo-nexus-manual-20261120-101010-%06x.zip", i)
		if err := os.WriteFile(filepath.Join(dir, name), []byte("manual"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	legacy := "mundo-na-rede-20261120-101010-ffffff.zip"
	if err := os.WriteFile(filepath.Join(dir, legacy), []byte("legacy"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := pruneLocalBackups(dir); err != nil {
		t.Fatal(err)
	}
	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 13 {
		t.Fatalf("backup manual foi removido: restaram %d", len(files))
	}
}
