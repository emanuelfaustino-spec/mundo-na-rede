package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func bedrockTuning(p Profile) (int, int) {
	switch p.Preset {
	case "light":
		return 10, 4
	case "balanced":
		return 16, 4
	case "reach":
		return 24, 6
	default:
		return 0, 0
	}
}

func (m *manager) worldCover(w http.ResponseWriter, r *http.Request, p Profile) {
	cover := filepath.Join(m.dir(p.ID), "cover")
	if r.Method == "GET" {
		data, err := os.ReadFile(cover)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", http.DetectContentType(data))
		w.Write(data)
		return
	}
	if r.Method != "POST" {
		w.WriteHeader(405)
		return
	}
	data, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 4<<20))
	if err != nil {
		fail(w, errors.New("a capa deve ter até 4 MiB"))
		return
	}
	cfg, kind, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || (kind != "jpeg" && kind != "png") || cfg.Width < 1 || cfg.Height < 1 || cfg.Width > 4096 || cfg.Height > 4096 {
		fail(w, errors.New("selecione uma imagem JPEG ou PNG de até 4096 × 4096"))
		return
	}
	if _, _, err = image.Decode(bytes.NewReader(data)); err != nil {
		fail(w, err)
		return
	}
	tmp, err := os.CreateTemp(m.dir(p.ID), "cover-")
	if err != nil {
		fail(w, err)
		return
	}
	defer os.Remove(tmp.Name())
	_, err = tmp.Write(data)
	if err = errors.Join(err, tmp.Close()); err == nil {
		err = os.Rename(tmp.Name(), cover)
	}
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, map[string]bool{"ok": true})
}
func playerDetails(game string, proc *process) map[string]any {
	mode, allowed := bedrockAccess(game)
	online := []string{}
	banned := []string{}
	recent := map[string]time.Time{}
	data, _ := os.ReadFile(filepath.Join(game, ".mnr-players"))
	json.Unmarshal(data, &recent)
	if proc != nil {
		for name := range proc.players {
			online = append(online, name)
		}
	}
	for name := range bannedPlayers(game) {
		banned = append(banned, name)
	}
	sort.Strings(online)
	sort.Strings(allowed)
	sort.Strings(banned)
	return map[string]any{"mode": mode, "allowed": allowed, "online": online, "banned": banned, "recent": recent}
}

func bannedPath(game string) string { return filepath.Join(game, ".mnr-banned.json") }

// Bedrock does not expose a portable on-disk ban list. Keep the app's list
// separately, and eject a listed player whenever BDS reports a connection.
func bannedPlayers(game string) map[string]bool {
	items := []string{}
	data, _ := os.ReadFile(bannedPath(game))
	_ = json.Unmarshal(data, &items)
	out := map[string]bool{}
	for _, name := range items {
		if gamerName.MatchString(name) {
			out[strings.ToLower(name)] = true
		}
	}
	return out
}

func saveBannedPlayers(game string, players map[string]bool) error {
	items := []string{}
	for name := range players {
		items = append(items, name)
	}
	sort.Strings(items)
	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(bannedPath(game), data, 0600)
}

func updateAllowedPlayers(game, name string, allow bool) error {
	mode, current := bedrockAccess(game)
	entries := []map[string]any{}
	for _, player := range current {
		if !strings.EqualFold(player, name) || allow {
			entries = append(entries, map[string]any{"name": player, "ignoresPlayerLimit": false})
		}
	}
	if allow {
		found := false
		for _, entry := range entries {
			if strings.EqualFold(entry["name"].(string), name) {
				found = true
			}
		}
		if !found {
			entries = append(entries, map[string]any{"name": name, "ignoresPlayerLimit": false})
		}
	}
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	if err = os.WriteFile(filepath.Join(game, "allowlist.json"), data, 0600); err != nil {
		return err
	}
	if mode == "" {
		return os.WriteFile(filepath.Join(game, ".mnr-access"), []byte("approved"), 0600)
	}
	return nil
}
func recordPlayer(game, name string) {
	if name == "" || len(name) > 100 {
		return
	}
	path := filepath.Join(game, ".mnr-players")
	players := map[string]time.Time{}
	data, _ := os.ReadFile(path)
	json.Unmarshal(data, &players)
	if players == nil {
		players = map[string]time.Time{}
	}
	if len(players) >= 300 {
		var oldest string
		for n := range players {
			if oldest == "" || players[n].Before(players[oldest]) {
				oldest = n
			}
		}
		delete(players, oldest)
	}
	players[name] = time.Now()
	data, _ = json.Marshal(players)
	os.WriteFile(path, data, 0600)
}
func (m *manager) playerAction(w http.ResponseWriter, r *http.Request, p Profile, proc *process) error {
	var input struct{ Action, Name string }
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&input); err != nil {
		return err
	}
	if p.Engine != "bedrock" || !gamerName.MatchString(input.Name) {
		return errors.New("jogador inválido ou motor incompatível")
	}
	switch input.Action {
	case "allow":
		return updateAllowedPlayers(filepath.Join(m.dir(p.ID), "game"), input.Name, true)
	case "unallow":
		return updateAllowedPlayers(filepath.Join(m.dir(p.ID), "game"), input.Name, false)
	case "ban", "unban":
		game := filepath.Join(m.dir(p.ID), "game")
		banned := bannedPlayers(game)
		key := strings.ToLower(input.Name)
		if input.Action == "ban" {
			banned[key] = true
		} else {
			delete(banned, key)
		}
		if err := saveBannedPlayers(game, banned); err != nil {
			return err
		}
		if proc != nil {
			proc.banned = banned
			if input.Action == "ban" && proc.state == "online" && proc.players[input.Name] {
				return sendCommand(proc, fmt.Sprintf(`kick "%s" Você foi banido deste mundo`, input.Name))
			}
		}
		return nil
	case "kick", "op", "deop":
		if proc == nil || proc.state != "online" {
			return errors.New("inicie o servidor para aplicar essa ação")
		}
		return sendCommand(proc, fmt.Sprintf(`%s "%s"`, input.Action, input.Name))
	default:
		return errors.New("ação desconhecida")
	}
}
func worldSize(game string) int64 {
	var total int64
	filepath.WalkDir(filepath.Join(game, "worlds"), func(_ string, e fs.DirEntry, err error) error {
		if err == nil && !e.IsDir() && e.Type().IsRegular() {
			if info, err := e.Info(); err == nil {
				total += info.Size()
			}
		}
		return nil
	})
	return total
}
func configuredDistances(game string) map[string]string {
	out := map[string]string{}
	data, _ := os.ReadFile(filepath.Join(game, "server.properties"))
	for _, line := range strings.Split(string(data), "\n") {
		key, value, ok := strings.Cut(line, "=")
		if ok && (key == "view-distance" || key == "tick-distance") {
			out[key] = strings.TrimSpace(value)
		}
	}
	return out
}
