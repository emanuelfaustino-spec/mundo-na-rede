package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPlayerAccessAndLocalBan(t *testing.T) {
	game := t.TempDir()
	if err := updateAllowedPlayers(game, "Alex", true); err != nil {
		t.Fatal(err)
	}
	mode, allowed := bedrockAccess(game)
	if mode != "approved" || len(allowed) != 1 || allowed[0] != "Alex" {
		t.Fatalf("allowlist inesperada: mode=%q players=%v", mode, allowed)
	}
	if err := updateAllowedPlayers(game, "aLeX", false); err != nil {
		t.Fatal(err)
	}
	_, allowed = bedrockAccess(game)
	if len(allowed) != 0 {
		t.Fatalf("jogador ainda permitido: %v", allowed)
	}

	banned := map[string]bool{"alex": true}
	if err := saveBannedPlayers(game, banned); err != nil {
		t.Fatal(err)
	}
	if !bannedPlayers(game)["alex"] {
		t.Fatal("banimento não foi salvo")
	}
	if _, err := os.Stat(filepath.Join(game, ".mnr-banned.json")); err != nil {
		t.Fatal(err)
	}
}
