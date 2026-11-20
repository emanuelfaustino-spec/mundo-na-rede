package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPruneBackupsWithFewerThanRetention(t *testing.T) {
	for _, count := range []int{0, 1, backupRetention} {
		t.Run(string(rune('0'+count)), func(t *testing.T) {
			dir := t.TempDir()
			for i := 0; i < count; i++ {
				name := "mundo-nexus-20261120-101010-00000" + string(rune('0'+i)) + ".zip"
				if err := os.WriteFile(filepath.Join(dir, name), []byte("test"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if err := pruneLocalBackups(dir); err != nil {
				t.Fatal(err)
			}
			files, err := os.ReadDir(dir)
			if err != nil {
				t.Fatal(err)
			}
			if len(files) != count {
				t.Fatalf("esperava %d arquivos, encontrou %d", count, len(files))
			}
		})
	}
}
