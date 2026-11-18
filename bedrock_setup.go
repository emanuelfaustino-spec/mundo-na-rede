package main

import (
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var bedrockVersion = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$`)
var gamerName = regexp.MustCompile(`^[\p{L}\p{N}][\p{L}\p{N} _-]{0,31}$`)

func (m *manager) versions() []string {
	entries, _ := os.ReadDir(filepath.Join(m.root, "engines", "bedrock"))
	versions := []string{}
	for _, e := range entries {
		if e.IsDir() && bedrockVersion.MatchString(e.Name()) && hasEngine(filepath.Join(m.root, "engines", "bedrock", e.Name()), "bedrock") {
			versions = append(versions, e.Name())
		}
	}
	sort.Slice(versions, func(i, j int) bool { return compareVersion(versions[i], versions[j]) > 0 })
	return versions
}
func (m *manager) legacyEngine(exclude string) string {
	entries, _ := os.ReadDir(filepath.Join(m.root, "servers"))
	for _, entry := range entries {
		if entry.Name() == exclude || !profileID.MatchString(entry.Name()) {
			continue
		}
		profile, err := m.load(entry.Name())
		if err != nil || profile.Engine != "bedrock" {
			continue
		}
		game := filepath.Join(m.dir(profile.ID), "game")
		if hasEngine(game, "bedrock") {
			return game
		}
	}
	return ""
}
func compareVersion(a, b string) int {
	aa, bb := strings.Split(a, "."), strings.Split(b, ".")
	for i := range aa {
		if len(aa[i]) != len(bb[i]) {
			return len(aa[i]) - len(bb[i])
		}
		if n := strings.Compare(aa[i], bb[i]); n != 0 {
			return n
		}
	}
	return 0
}
func hasEngine(dir, engine string) bool {
	if engine == "bedrock" {
		for _, name := range []string{"bedrock_server.exe", "bedrock_server"} {
			if st, err := os.Stat(filepath.Join(dir, name)); err == nil && st.Mode().IsRegular() {
				return true
			}
		}
		return false
	}
	st, err := os.Stat(filepath.Join(dir, "server.jar"))
	return err == nil && st.Mode().IsRegular()
}
func copyMissing(source, destination string) error {
	return filepath.WalkDir(source, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, relative)
		if entry.IsDir() {
			return os.MkdirAll(target, 0700)
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return errors.New("arquivo de motor inválido")
		}
		from, err := os.Open(path)
		if err != nil {
			return err
		}
		defer from.Close()
		to, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, info.Mode().Perm())
		if errors.Is(err, os.ErrExist) {
			return nil
		}
		if err != nil {
			return err
		}
		_, err = io.Copy(to, from)
		if err = errors.Join(err, to.Close()); err != nil {
			os.Remove(target)
		}
		return err
	})
}
func (m *manager) prepareAccess(w http.ResponseWriter, r *http.Request, p Profile, game string) error {
	var input struct {
		Mode    string
		Players []string
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8192)).Decode(&input); err != nil {
		return err
	}
	if input.Mode != "lan" && input.Mode != "approved" {
		return errors.New("escolha quem pode entrar")
	}
	if len(input.Players) > 100 || (input.Mode == "approved" && len(input.Players) == 0) {
		return errors.New("adicione pelo menos uma gamertag")
	}
	seen := map[string]bool{}
	entries := []map[string]any{}
	for _, name := range input.Players {
		name = strings.TrimSpace(name)
		key := strings.ToLower(name)
		if !gamerName.MatchString(name) || seen[key] {
			return errors.New("gamertag inválida ou repetida")
		}
		seen[key] = true
		entries = append(entries, map[string]any{"name": name, "ignoresPlayerLimit": false})
	}
	if input.Mode == "approved" {
		// Keep existing Xbox IDs and exemptions for players the administrator retains.
		old, _ := os.ReadFile(filepath.Join(game, "allowlist.json"))
		var previous []map[string]any
		json.Unmarshal(old, &previous)
		for i, entry := range entries {
			for _, prev := range previous {
				if name, ok := prev["name"].(string); ok && strings.EqualFold(name, entry["name"].(string)) {
					entries[i] = prev
					break
				}
			}
		}
		bytes, _ := json.MarshalIndent(entries, "", "  ")
		if err := os.WriteFile(filepath.Join(game, "allowlist.json"), bytes, 0600); err != nil {
			return err
		}
	}
	return os.WriteFile(filepath.Join(game, ".mnr-access"), []byte(input.Mode), 0600)
}
func bedrockAccess(game string) (string, []string) {
	marker, _ := os.ReadFile(filepath.Join(game, ".mnr-access"))
	raw, _ := os.ReadFile(filepath.Join(game, "allowlist.json"))
	var entries []struct{ Name string }
	json.Unmarshal(raw, &entries)
	names := []string{}
	for _, entry := range entries {
		if entry.Name != "" {
			names = append(names, entry.Name)
		}
	}
	return string(marker), names
}
