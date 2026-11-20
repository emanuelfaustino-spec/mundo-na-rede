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
		prefix := "mundo-nexus-"
		if i%2 == 0 {
			prefix = "mundo-na-rede-"
		}
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
