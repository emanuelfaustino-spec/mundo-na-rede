package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBedrockAddonAttachesBothKinds(t *testing.T) {
	dir := t.TempDir()
	src, game := filepath.Join(dir, "unpacked"), filepath.Join(dir, "game")
	for i, kind := range []string{"resources", "data"} {
		pack := filepath.Join(src, kind)
		if err := os.MkdirAll(pack, 0700); err != nil {
			t.Fatal(err)
		}
		uuid := []string{"11111111-1111-1111-1111-111111111111", "22222222-2222-2222-2222-222222222222"}[i]
		manifest := `{"header":{"uuid":"` + uuid + `","version":[1,2,3]},"modules":[{"type":"` + kind + `"}]}`
		if err := os.WriteFile(filepath.Join(pack, "manifest.json"), []byte(manifest), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := installBedrockAddon(src, game, "auto"); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if err := attachBedrockPacks(game); err != nil {
			t.Fatal(err)
		}
	}
	world := filepath.Join(game, "worlds", "Mundo")
	for _, group := range []string{"resource", "behavior"} {
		data, err := os.ReadFile(filepath.Join(world, "world_"+group+"_packs.json"))
		if err != nil || strings.Count(string(data), `"pack_id"`) != 1 {
			t.Fatalf("%s: %s (%v)", group, data, err)
		}
	}
}

func TestBedrockCombinedPackInstallsInBothFolders(t *testing.T) {
	dir := t.TempDir()
	src, game := filepath.Join(dir, "unpacked"), filepath.Join(dir, "game")
	pack := filepath.Join(src, "combined")
	if err := os.MkdirAll(pack, 0700); err != nil {
		t.Fatal(err)
	}
	const id = "33333333-3333-3333-3333-333333333333"
	manifest := `{"header":{"uuid":"` + id + `","version":[1,0,0]},"modules":[{"type":"resources"},{"type":"data"}]}`
	if err := os.WriteFile(filepath.Join(pack, "manifest.json"), []byte(manifest), 0600); err != nil {
		t.Fatal(err)
	}
	if err := installBedrockAddon(src, game, "auto"); err != nil {
		t.Fatal(err)
	}
	for _, folder := range []string{"resource_packs", "behavior_packs"} {
		if _, err := os.Stat(filepath.Join(game, folder, "mnr-"+id, "manifest.json")); err != nil {
			t.Fatalf("%s: %v", folder, err)
		}
	}
}
