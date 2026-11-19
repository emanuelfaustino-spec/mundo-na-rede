package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var publicHost = regexp.MustCompile("(?i)^[a-z0-9](?:[a-z0-9.-]*[a-z0-9])?$")

func validPublicAddress(address string) bool {
	if address == "" {
		return true
	}
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return false
	}
	number, err := strconv.Atoi(port)
	if err != nil || number < 1 || number > 65535 {
		return false
	}
	if ip := net.ParseIP(host); ip != nil {
		return ip.IsGlobalUnicast() && !ip.IsPrivate()
	}
	return len(host) <= 253 && strings.Contains(host, ".") && publicHost.MatchString(host) && !strings.Contains(host, "..") && !strings.HasSuffix(strings.ToLower(host), ".local")
}
func validateProfile(p *Profile) error {
	p.Name = strings.TrimSpace(p.Name)
	p.PublicJava = strings.TrimSpace(p.PublicJava)
	p.PublicBedrock = strings.TrimSpace(p.PublicBedrock)
	switch p.Engine {
	case "bedrock", "pnx", "java", "both":
	default:
		return errors.New("motor inválido")
	}
	if p.Name == "" || len(p.Name) > 80 || strings.ContainsAny(p.Name, "\r\n\\") ||
		p.Port < 1024 || p.Port > 65534 || (p.Engine == "both" && p.Port == crossplayPort) ||
		p.Players < 1 || p.Players > 100 || p.Memory < 512 || p.Memory > 32768 ||
		(p.Mode != "survival" && p.Mode != "creative") {
		return errors.New("revise o nome, a porta e os limites do servidor")
	}
	if len(p.Seed) > 64 || strings.ContainsAny(p.Seed, "\r\n\\") {
		return errors.New("seed inválida")
	}
	switch p.Difficulty {
	case "", "peaceful", "easy", "normal", "hard":
	default:
		return errors.New("dificuldade inválida")
	}
	switch p.Preset {
	case "", "light", "balanced", "reach":
	default:
		return errors.New("perfil de distância inválido")
	}
	if (p.ViewDistance != 0 && (p.ViewDistance < 5 || p.ViewDistance > 64)) ||
		(p.TickDistance != 0 && (p.TickDistance < 4 || p.TickDistance > 12)) {
		return errors.New("use visão entre 5 e 64 chunks e simulação entre 4 e 12")
	}
	if !validPublicAddress(p.PublicJava) || !validPublicAddress(p.PublicBedrock) {
		return errors.New("use endereço público e porta, por exemplo jogar.exemplo.com:25565")
	}
	return nil
}
func recommendedPlayers() int {
	_, available := memoryMiB()
	if available <= 0 {
		return 4
	}
	return max(2, min(12, runtime.NumCPU()*2, available/384))
}
func (m *manager) updateWorld(w http.ResponseWriter, r *http.Request, p Profile, active bool) {
	var next Profile
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8192)).Decode(&next); err != nil {
		fail(w, err)
		return
	}
	next.ID = p.ID
	next.Engine = p.Engine
	// The seed is read only after creation. Changing server.properties later
	// does not safely replace an existing Bedrock world.
	next.Seed = p.Seed
	if err := validateProfile(&next); err != nil {
		fail(w, err)
		return
	}
	if active && (next.Name != p.Name || next.Port != p.Port || next.Mode != p.Mode || next.Memory != p.Memory || next.Players != p.Players || next.Seed != p.Seed || next.Difficulty != p.Difficulty || next.Preset != p.Preset || next.ViewDistance != p.ViewDistance || next.TickDistance != p.TickDistance) {
		fail(w, errors.New("pare o servidor para alterar as opções do mundo"))
		return
	}
	if p.Engine == "java" {
		next.PublicBedrock = ""
	} else if p.Engine != "both" {
		next.PublicJava = ""
	}
	dir := m.dir(p.ID)
	tmp, err := os.CreateTemp(dir, "profile-update-")
	if err != nil {
		fail(w, err)
		return
	}
	defer os.Remove(tmp.Name())
	if err = errors.Join(json.NewEncoder(tmp).Encode(next), tmp.Close()); err != nil {
		fail(w, err)
		return
	}
	if err = os.Rename(tmp.Name(), filepath.Join(dir, "profile.json")); err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, next)
}

type worldExtension struct {
	Name    string `json:"name"`
	Kind    string `json:"kind"`
	ID      string `json:"id,omitempty"`
	Version string `json:"version,omitempty"`
}

func listExtensions(game, engine string, advanced bool) ([]worldExtension, error) {
	out := []worldExtension{}
	for _, folder := range []string{"resource_packs", "behavior_packs", "plugins", "mods"} {
		entries, err := os.ReadDir(filepath.Join(game, folder))
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			if engine == "bedrock" && !advanced && (folder == "resource_packs" || folder == "behavior_packs") && !strings.HasPrefix(entry.Name(), "mnr-") {
				continue
			}
			if strings.HasPrefix(entry.Name(), "pacote-anterior-") || strings.HasPrefix(entry.Name(), "extensao-anterior-") {
				continue
			}
			if folder == "mods" || folder == "plugins" {
				if !strings.EqualFold(filepath.Ext(entry.Name()), ".jar") || entry.IsDir() {
					continue
				}
			} else if !entry.IsDir() && filepath.Ext(entry.Name()) != ".zip" && filepath.Ext(entry.Name()) != ".mcpack" {
				continue
			}
			label := entry.Name()
			if entry.IsDir() && (folder == "resource_packs" || folder == "behavior_packs") {
				raw, err := os.ReadFile(filepath.Join(game, folder, entry.Name(), "manifest.json"))
				var pack struct{ Header struct{ Name string } }
				if err == nil && json.Unmarshal(raw, &pack) == nil && pack.Header.Name != "" {
					label = pack.Header.Name
				}
			}
			item := worldExtension{Name: label, Kind: folder}
			if engine == "bedrock" && strings.HasPrefix(entry.Name(), "mnr-") {
				if pack, err := readPack(filepath.Join(game, folder, entry.Name())); err == nil {
					item.ID = pack.Ref.ID
					var ver []uint
					if json.Unmarshal(pack.Ref.Version, &ver) == nil && len(ver) == 3 {
						item.Version = fmt.Sprintf("%d.%d.%d", ver[0], ver[1], ver[2])
					}
				}
			}
			out = append(out, item)
			if len(out) > 5000 {
				return nil, fmt.Errorf("há extensões demais para exibir")
			}
		}
	}
	return out, nil
}

type catalogHit struct {
	Title       string `json:"title"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
}

func searchCatalog(w http.ResponseWriter, r *http.Request, p Profile) {
	kind := r.URL.Query().Get("kind")
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if len(query) > 140 || (kind != "mod" && kind != "plugin") ||
		(p.Engine != "java" && !(p.Engine == "both" && kind == "plugin")) {
		fail(w, errors.New("catálogo indisponível para este motor"))
		return
	}
	facets, _ := json.Marshal([][]string{{"project_type:" + kind}, {"server_side:required", "server_side:optional"}})
	params := url.Values{"query": {query}, "limit": {"8"}, "facets": {string(facets)}}
	var payload struct {
		Hits []catalogHit `json:"hits"`
	}
	client := &http.Client{Timeout: 8 * time.Second, CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }}
	if err := downloadMetadata(client, r.Context(), "https://api.modrinth.com/v2/search?"+params.Encode(), &payload); err != nil {
		fail(w, err)
		return
	}
	if payload.Hits == nil {
		payload.Hits = []catalogHit{}
	}
	writeJSON(w, payload.Hits)
}
