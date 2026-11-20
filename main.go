package main

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/skip2/go-qrcode"
	"gopkg.in/yaml.v3"
)

//go:embed ui/*
var assets embed.FS

const maxArchive = 2 << 30

var ansiCodes = regexp.MustCompile("\x1b\\[[0-?]*[ -/]*[@-~]")
var profileID = regexp.MustCompile("^[a-f0-9]{16}$")

type Profile struct {
	ID                  string          `json:"id"`
	Name                string          `json:"name"`
	Engine              string          `json:"engine"`
	Port                int             `json:"port"`
	Players             int             `json:"players"`
	Memory              int             `json:"memory"`
	Mode                string          `json:"mode"`
	Seed                string          `json:"seed,omitempty"`
	Difficulty          string          `json:"difficulty,omitempty"`
	Preset              string          `json:"preset,omitempty"`
	ViewDistance        int             `json:"viewDistance,omitempty"`
	TickDistance        int             `json:"tickDistance,omitempty"`
	AllowCheats         bool            `json:"allowCheats,omitempty"`
	ForceGamemode       bool            `json:"forceGamemode,omitempty"`
	DefaultPermission   string          `json:"defaultPermission,omitempty"`
	TexturePackRequired bool            `json:"texturePackRequired,omitempty"`
	LANVisibility       bool            `json:"lanVisibility,omitempty"`
	LANVisibilitySet    bool            `json:"lanVisibilitySet,omitempty"`
	Rules               map[string]bool `json:"rules,omitempty"`
	PublicJava          string          `json:"publicJava,omitempty"`
	PublicBedrock       string          `json:"publicBedrock,omitempty"`
}
type process struct {
	cmd           *exec.Cmd
	commands      chan string
	done          chan struct{}
	started       time.Time
	restart       bool
	ready         bool
	players       map[string]bool
	banned        map[string]bool
	accessWarning bool
	sample        cpuSample
	state         string
	lines         []string
}
type manager struct {
	root     string
	progress sync.Map
	drive    *driveBackup
	token    string
	mu       sync.Mutex
	running  map[string]*process
	lastRun  map[string]time.Time
}

func randomHex(bytes int) string {
	key := make([]byte, bytes)
	if _, err := rand.Read(key); err != nil {
		panic(err)
	}
	return hex.EncodeToString(key)
}
func (m *manager) dir(id string) string { return filepath.Join(m.root, "servers", id) }
func (m *manager) load(id string) (Profile, error) {
	var p Profile
	if !profileID.MatchString(id) {
		return p, errors.New("servidor inválido")
	}
	data, err := os.ReadFile(filepath.Join(m.dir(id), "profile.json"))
	if err != nil {
		return p, err
	}
	return p, json.Unmarshal(data, &p)
}
func writeJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
func fail(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusBadRequest)
	writeJSON(w, map[string]string{"error": err.Error()})
}
func (m *manager) api(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if subtle.ConstantTimeCompare([]byte(r.Header.Get("X-Local-Key")), []byte(m.token)) != 1 {
		w.WriteHeader(http.StatusForbidden)
		writeJSON(w, map[string]string{"error": "Sessão local inválida. Reabra o aplicativo."})
		return
	}
	if m.updateAPI(w, r) {
		return
	}
	if r.Method == "GET" && strings.HasSuffix(r.URL.Path, "/catalog") {
		route := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/servers/"), "/")
		if len(route) != 2 {
			http.NotFound(w, r)
			return
		}
		m.mu.Lock()
		profile, err := m.load(route[0])
		m.mu.Unlock()
		if err != nil {
			http.NotFound(w, r)
			return
		}
		searchCatalog(w, r, profile)
		return
	}
	if r.Method == "GET" && strings.HasSuffix(r.URL.Path, "/progress") {
		id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/servers/"), "/progress")
		if !profileID.MatchString(id) {
			http.NotFound(w, r)
			return
		}
		phase, _ := m.progress.Load(id)
		writeJSON(w, map[string]any{"phase": phase})
		return
	}
	if strings.HasPrefix(r.URL.Path, "/api/account") {
		if r.URL.Path == "/api/account/backups/delete" && r.Method == "POST" {
			var input struct {
				ID string `json:"id"`
			}
			if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&input); err != nil {
				fail(w, err)
				return
			}
			ctx, cancel := context.WithTimeout(r.Context(), 35*time.Second)
			defer cancel()
			if err := m.drive.deleteCopy(ctx, input.ID); err != nil {
				fail(w, err)
				return
			}
			writeJSON(w, map[string]bool{"ok": true})
			return
		}
		if r.URL.Path == "/api/account/backups" && r.Method == "GET" {
			ctx, cancel := context.WithTimeout(r.Context(), 35*time.Second)
			defer cancel()
			copies, err := m.drive.listCopies(ctx)
			if err != nil {
				fail(w, err)
				return
			}
			writeJSON(w, copies)
			return
		}
		m.drive.handle(w, r)
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if r.URL.Path == "/api/servers" {
		if r.Method == "GET" {
			entries, err := os.ReadDir(filepath.Join(m.root, "servers"))
			if err != nil {
				fail(w, err)
				return
			}
			list := []Profile{}
			for _, entry := range entries {
				if p, err := m.load(entry.Name()); err == nil {
					list = append(list, p)
				}
			}
			total, available := memoryMiB()
			writeJSON(w, map[string]any{"servers": list, "ips": localIPs(), "os": runtime.GOOS, "cores": runtime.NumCPU(), "totalMemory": total, "availableMemory": available, "recommendedPlayers": recommendedPlayers()})
			return
		}
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var p Profile
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8192)).Decode(&p); err != nil {
			fail(w, err)
			return
		}
		p.PublicJava, p.PublicBedrock = "", ""
		if p.Engine == "bedrock" {
			p.LANVisibility, p.LANVisibilitySet = true, true
		}
		if err := validateProfile(&p); err != nil {
			fail(w, err)
			return
		}
		p.ID = randomHex(8)
		dir := m.dir(p.ID)
		if err := os.MkdirAll(filepath.Join(dir, "game"), 0700); err != nil {
			fail(w, err)
			return
		}
		data, _ := json.MarshalIndent(p, "", "  ")
		if err := os.WriteFile(filepath.Join(dir, "profile.json"), data, 0600); err != nil {
			os.RemoveAll(dir)
			fail(w, err)
			return
		}
		writeJSON(w, p)
		return
	}
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/servers/"), "/")
	if len(parts) != 2 {
		http.NotFound(w, r)
		return
	}
	p, err := m.load(parts[0])
	if err != nil {
		http.NotFound(w, r)
		return
	}
	action := parts[1]
	if action == "backup-history" {
		m.backupHistory(w, r, p)
		return
	}
	if action == "schedules" && (r.Method == "GET" || r.Method == "POST") {
		m.configureSchedule(w, r, p.ID)
		return
	}
	game := filepath.Join(m.dir(p.ID), "game")
	proc := m.running[p.ID]
	active := proc != nil && proc.state != "parado" && proc.state != "falhou"
	if action == "invite" && r.Method == "GET" {
		ip := r.URL.Query().Get("ip")
		if !slices.Contains(localIPs(), ip) {
			fail(w, errors.New("endereço local indisponível"))
			return
		}
		qr, err := qrcode.Encode(net.JoinHostPort(ip, strconv.Itoa(p.Port)), qrcode.Medium, 220)
		if err != nil {
			fail(w, err)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		w.Write(qr)
		return
	}
	if action == "cover" {
		m.worldCover(w, r, p)
		return
	}
	if action == "players" && r.Method == "GET" {
		writeJSON(w, playerDetails(game, proc))
		return
	}
	if action == "world-details" && r.Method == "GET" {
		writeJSON(w, map[string]any{"bytes": worldSize(game), "distances": configuredDistances(game)})
		return
	}
	if action == "world-settings" && r.Method == "POST" {
		m.updateWorld(w, r, p, active)
		return
	}
	if r.Method == "GET" {
		if action == "extensions" {
			items, err := listExtensions(game, p.Engine, r.URL.Query().Get("advanced") == "1")
			if err != nil {
				fail(w, err)
				return
			}
			writeJSON(w, items)
			return
		}
		if action == "config" {
			name := "server.properties"
			if p.Engine == "pnx" {
				name = "pnx.yml"
			}
			file, readErr := os.Open(filepath.Join(game, name))
			if errors.Is(readErr, os.ErrNotExist) {
				writeJSON(w, map[string]string{"name": name, "text": ""})
				return
			}
			if readErr != nil {
				fail(w, readErr)
				return
			}
			defer file.Close()
			raw, readErr := io.ReadAll(io.LimitReader(file, (256<<10)+1))
			if readErr != nil {
				fail(w, readErr)
				return
			}
			if len(raw) > 256<<10 {
				fail(w, errors.New("arquivo de configuração excede o limite do editor"))
				return
			}
			writeJSON(w, map[string]string{"name": name, "text": string(raw)})
			return
		}
		if action == "backup" {
			if active {
				fail(w, errors.New("pare o servidor para criar um backup consistente"))
				return
			}
			m.backup(w, game, p.ID)
			return
		}
		if action != "status" {
			http.NotFound(w, r)
			return
		}
		state := "parado"
		var lines []string
		var seconds int64
		if proc != nil {
			state = proc.state
			lines = proc.lines
			if active {
				seconds = int64(time.Since(proc.started).Seconds())
			}
		}
		_, engineError := m.command(p, game)
		result := map[string]any{"state": state, "logs": lines, "seconds": seconds, "installed": engineError == nil}
		if p.Engine == "bedrock" {
			mode, players := bedrockAccess(game)
			version, _ := os.ReadFile(filepath.Join(game, ".mnr-engine"))
			result["access"] = mode
			result["approved"] = players
			result["versions"] = m.versions()
			result["legacyMotor"] = m.legacyEngine(p.ID) != ""
			result["version"] = string(version)
		}
		if proc != nil {
			result["players"] = len(proc.players)
			result["accessWarning"] = proc.accessWarning
			if active {
				result["performance"] = proc.measurements()
			}
		}
		total, available := memoryMiB()
		result["memoryTotal"] = total
		result["memoryUsed"] = max(0, total-available)
		writeJSON(w, result)
		return
	}
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	switch action {
	case "delete":
		if active {
			err = errors.New("pare o servidor antes de excluir")
		} else {
			err = os.RemoveAll(m.dir(p.ID))
		}
	case "install", "restore", "world", "addon", "config":
		if active {
			fail(w, errors.New("pare o servidor antes de editar os arquivos"))
			return
		}
		err = m.changeFiles(w, r, p, game, action)
	case "restore-preview", "restore-apply":
		if active {
			err = errors.New("pare o servidor antes de editar os arquivos")
			break
		}
		defer m.progress.Delete(p.ID)
		if action == "restore-apply" {
			err = m.applyRestore(w, r, p, game)
			break
		}
		preview, stageErr := m.stageRestore(w, r, p)
		if stageErr != nil {
			err = stageErr
			break
		}
		writeJSON(w, preview)
		return
	case "remove-addon":
		if active {
			err = errors.New("pare o servidor antes de editar os arquivos")
		} else {
			err = m.removeBedrockPack(w, r, p, game)
		}
	case "player-action":
		err = m.playerAction(w, r, p, proc)
	case "access":
		if active || p.Engine != "bedrock" {
			err = errors.New("pare o servidor Bedrock para mudar o acesso")
		} else {
			err = m.prepareAccess(w, r, p, game)
		}
	case "cloud-restore":
		if active {
			err = errors.New("pare o servidor antes de editar os arquivos")
		} else {
			err = m.restoreCloud(r, p, game)
		}
	case "crossplay":
		if active {
			err = errors.New("pare o servidor primeiro")
		} else {
			err = m.installCrossplay(r, p, game)
		}
	case "official":
		if active {
			err = errors.New("pare o servidor primeiro")
		} else {
			err = m.installOfficial(r, p, game)
		}
	case "start":
		if active {
			err = errors.New("servidor já está ativo")
		} else {
			err = m.start(p, game)
		}
	case "stop", "restart":
		if !active || proc.state == "encerrando" {
			err = errors.New("servidor indisponível")
		} else {
			err = sendCommand(proc, "stop")
			if err == nil {
				proc.state = "encerrando"
				proc.restart = action == "restart"
			}
		}
	case "accept-eula":
		if active || p.Engine != "java" {
			err = errors.New("disponível para Java com o servidor parado")
		} else {
			err = os.WriteFile(filepath.Join(game, "eula.txt"), []byte("eula=true\n"), 0600)
		}
	case "force":
		if !active {
			err = errors.New("servidor indisponível")
		} else {
			proc.restart = false
			err = proc.cmd.Process.Kill()
		}
	case "command":
		if !active || proc.state == "encerrando" {
			err = errors.New("servidor indisponível")
			break
		}
		data, readErr := io.ReadAll(http.MaxBytesReader(w, r.Body, 1024))
		if readErr != nil {
			err = readErr
		} else {
			err = sendCommand(proc, strings.TrimSpace(string(data)))
		}
	default:
		http.NotFound(w, r)
		return
	}
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, map[string]bool{"ok": true})
}
func sendCommand(p *process, command string) error {
	if command == "" || strings.ContainsAny(command, "\r\n\x00") {
		return errors.New("envie um comando por vez")
	}
	select {
	case <-p.done:
		return errors.New("processo encerrado")
	case p.commands <- command:
		return nil
	default:
		return errors.New("o console já tem comandos pendentes")
	}
}

// applyBedrockRules runs only after BDS reports that the world is ready. BDS
// stores gamerules inside the world, so applying them at startup also makes
// the selected rules survive the next launch without editing level.dat.
func applyBedrockRules(p *process, rules map[string]bool) {
	for _, rule := range []string{
		"commandblockoutput", "commandblocksenabled", "dodaylightcycle",
		"doentitydrops", "dofiretick", "doimmediaterespawn", "doinsomnia",
		"domobloot", "domobspawning", "dotiledrops", "doweathercycle",
		"drowningdamage", "falldamage", "firedamage", "freezedamage",
		"keepinventory", "mobgriefing", "naturalregeneration", "pvp",
		"sendcommandfeedback", "showcoordinates", "showdeathmessages", "tntexplodes",
	} {
		value, configured := rules[rule]
		if !configured {
			continue
		}
		if err := sendCommand(p, "gamerule "+rule+" "+strconv.FormatBool(value)); err != nil {
			return
		}
	}
}
func (m *manager) command(p Profile, game string) (*exec.Cmd, error) {
	var cmd *exec.Cmd
	if p.Engine == "bedrock" {
		name := "bedrock_server"
		if runtime.GOOS == "windows" {
			name += ".exe"
		}
		if runtime.GOOS != "windows" && runtime.GOOS != "linux" {
			return nil, errors.New("BDS disponível para Windows e Ubuntu")
		}
		path := filepath.Join(game, name)
		if _, err := os.Stat(path); err != nil {
			return nil, errors.New("importe o ZIP oficial do BDS para este sistema")
		}
		cmd = exec.Command(path)
		if runtime.GOOS == "linux" {
			cmd.Env = append(os.Environ(), "LD_LIBRARY_PATH="+game)
		}
	} else {
		java, err := exec.LookPath("java")
		if err != nil {
			return nil, errors.New("instale Java compatível e coloque o executável no PATH")
		}
		jar := filepath.Join(game, "server.jar")
		if _, err = os.Stat(jar); err != nil {
			return nil, errors.New("importe o JAR do motor escolhido")
		}
		args := []string{"-Xms256M", fmt.Sprintf("-Xmx%dM", p.Memory), "-jar", jar}
		if p.Engine == "java" || p.Engine == "both" {
			args = append(args, "nogui")
		}
		cmd = exec.Command(java, args...)
		if p.Engine == "both" {
			cmd.Env = append(os.Environ(), "GEYSER_JAVA_AUTH_TYPE=floodgate", "GEYSER_BEDROCK_PORT=19132", "GEYSER_BEDROCK_CLONE_REMOTE_PORT=false")
		}
	}
	cmd.Dir = game
	hideServerWindow(cmd)
	return cmd, nil
}

type logWriter struct {
	manager *manager
	process *process
	profile Profile
	pending string
}

func (l *logWriter) Write(data []byte) (int, error) {
	l.manager.mu.Lock()
	defer l.manager.mu.Unlock()
	lines := strings.Split(l.pending+strings.ReplaceAll(string(data), "\r", ""), "\n")
	l.pending = lines[len(lines)-1]
	if len(l.pending) > 8192 {
		l.pending = l.pending[len(l.pending)-8192:]
	}
	for _, line := range lines[:len(lines)-1] {
		line = ansiCodes.ReplaceAllString(line, "")
		if len(line) > 8192 {
			line = line[:8192]
		}
		l.process.lines = append(l.process.lines, line)
		if _, name, ok := strings.Cut(line, "Player connected: "); ok {
			player := strings.TrimSpace(strings.SplitN(name, ",", 2)[0])
			l.process.players[player] = true
			recordPlayer(l.process.cmd.Dir, player)
			if l.process.banned[strings.ToLower(player)] {
				// The local ban list is enforced immediately after BDS reports entry.
				// sendCommand is non-blocking and the console writer runs under m.mu.
				go sendCommand(l.process, fmt.Sprintf(`kick "%s" Você foi banido deste mundo`, player))
			}
		}
		if _, name, ok := strings.Cut(line, "Player disconnected: "); ok {
			delete(l.process.players, strings.TrimSpace(strings.SplitN(name, ",", 2)[0]))
		}
		if strings.Contains(strings.ToLower(line), "allow list is enabled but contains no entries") {
			l.process.accessWarning = true
		}
		lower := strings.ToLower(line)
		if l.process.state == "iniciando" && (strings.Contains(lower, "server started") || strings.Contains(lower, "done (") || strings.Contains(lower, "done!")) {
			l.process.state = "online"
			l.process.ready = true
			if l.profile.Engine == "bedrock" && len(l.profile.Rules) > 0 {
				go applyBedrockRules(l.process, l.profile.Rules)
			}
		}
	}
	if len(l.process.lines) > 350 {
		l.process.lines = append([]string(nil), l.process.lines[len(l.process.lines)-350:]...)
	}
	return len(data), nil
}
func (m *manager) start(p Profile, game string) error {
	for id, active := range m.running {
		if id == p.ID || active.state == "parado" || active.state == "falhou" {
			continue
		}
		other, err := m.load(id)
		if err != nil {
			continue
		}
		if (p.Engine == "both" && (other.Port == crossplayPort || other.Engine == "both" || (other.Engine == "bedrock" && other.Port+1 == crossplayPort))) ||
			(other.Engine == "both" && (p.Port == crossplayPort || (p.Engine == "bedrock" && p.Port+1 == crossplayPort))) ||
			p.Port == other.Port || (p.Engine == "bedrock" && p.Port+1 == other.Port) || (other.Engine == "bedrock" && other.Port+1 == p.Port) {
			return fmt.Errorf("porta em uso pelo servidor %s", other.Name)
		}
	}
	cmd, err := m.command(p, game)
	if err != nil {
		return err
	}
	if err = configure(p, game); err != nil {
		return err
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	proc := &process{cmd: cmd, commands: make(chan string, 32), done: make(chan struct{}), started: time.Now(), state: "iniciando", players: map[string]bool{}, banned: bannedPlayers(game)}
	writer := &logWriter{manager: m, process: proc, profile: p}
	cmd.Stdout = writer
	cmd.Stderr = writer
	if err = cmd.Start(); err != nil {
		stdin.Close()
		return err
	}
	m.running[p.ID] = proc
	go func() {
		for {
			select {
			case <-proc.done:
				return
			case command := <-proc.commands:
				if _, err := fmt.Fprintln(stdin, command); err != nil {
					return
				}
			}
		}
	}()
	go func() {
		err := cmd.Wait()
		m.mu.Lock()
		defer m.mu.Unlock()
		if writer.pending != "" {
			proc.lines = append(proc.lines, ansiCodes.ReplaceAllString(writer.pending, ""))
		}
		proc.state = "parado"
		if err != nil {
			proc.state = "falhou"
			proc.lines = append(proc.lines, err.Error())
		}
		if err == nil && proc.ready {
			m.drive.capture(p, game)
		}
		close(proc.done)
		if proc.restart && err == nil {
			if restartErr := m.start(p, game); restartErr != nil {
				proc.state = "falhou"
				proc.lines = append(proc.lines, restartErr.Error())
			}
		}
	}()
	return nil
}
func configurePNX(p Profile, game string) error {
	path := filepath.Join(game, "pnx.yml")
	raw, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	settings := map[string]any{}
	if len(raw) > 0 {
		if err = yaml.Unmarshal(raw, &settings); err != nil {
			return fmt.Errorf("pnx.yml inválido: %w", err)
		}
	}
	if settings == nil {
		settings = map[string]any{}
	}
	groups := map[string]map[string]any{
		"settings":          {"port": p.Port, "maxPlayers": p.Players, "motd": p.Name},
		"gameplay-settings": {"gamemode": map[string]int{"survival": 0, "creative": 1}[p.Mode]},
	}
	for group, values := range groups {
		entries, _ := settings[group].(map[string]any)
		if entries == nil {
			entries = map[string]any{}
			settings[group] = entries
		}
		for key, value := range values {
			entries[key] = value
		}
	}
	data, err := yaml.Marshal(settings)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func configure(p Profile, game string) error {
	if p.Engine == "pnx" {
		return configurePNX(p, game)
	}
	path := filepath.Join(game, "server.properties")
	settings := map[string]string{"server-port": strconv.Itoa(p.Port), "max-players": strconv.Itoa(p.Players), "gamemode": p.Mode, "level-name": "Mundo"}
	if p.Engine == "bedrock" {
		if err := attachBedrockPacks(game); err != nil {
			return err
		}
		mode, approved := bedrockAccess(game)
		if mode == "approved" && len(approved) == 0 {
			return errors.New("adicione pelo menos uma gamertag antes de iniciar")
		}
		if mode != "" || len(approved) == 0 {
			settings["allow-list"] = strconv.FormatBool(mode == "approved")
		}
		if p.Seed != "" {
			settings["level-seed"] = p.Seed
		}
		if p.Difficulty != "" {
			settings["difficulty"] = p.Difficulty
		}
		if view, tick := bedrockTuning(p); view != 0 {
			settings["view-distance"] = strconv.Itoa(view)
			settings["tick-distance"] = strconv.Itoa(tick)
		}
		settings["server-name"] = p.Name
		settings["server-portv6"] = strconv.Itoa(p.Port + 1)
		permission := p.DefaultPermission
		if permission == "" {
			permission = "member"
		}
		settings["allow-cheats"] = strconv.FormatBool(p.AllowCheats)
		settings["force-gamemode"] = strconv.FormatBool(p.ForceGamemode)
		settings["default-player-permission-level"] = permission
		settings["texturepack-required"] = strconv.FormatBool(p.TexturePackRequired)
		lanVisibility := true
		if p.LANVisibilitySet {
			lanVisibility = p.LANVisibility
		}
		settings["enable-lan-visibility"] = strconv.FormatBool(lanVisibility)
		// RakNet is the LAN transport understood by the normal Bedrock server list.
		// New BDS packages may default to NetherNet, which can start without opening
		// the classic UDP game ports exposed by this application.
		settings["transport"] = "raknet"
	} else {
		settings["motd"] = p.Name
		settings["pause-when-empty-seconds"] = "900"
	}
	raw, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	var result []string
	for _, line := range strings.Split(strings.TrimRight(string(raw), "\r\n"), "\n") {
		key, _, _ := strings.Cut(line, "=")
		if _, ok := settings[strings.TrimSpace(key)]; !ok && line != "" {
			result = append(result, strings.TrimRight(line, "\r"))
		}
	}
	for _, key := range []string{"server-name", "motd", "server-port", "server-portv6", "max-players", "gamemode", "level-name", "pause-when-empty-seconds", "allow-list", "level-seed", "difficulty", "view-distance", "tick-distance", "allow-cheats", "force-gamemode", "default-player-permission-level", "texturepack-required", "enable-lan-visibility", "transport"} {
		if value, ok := settings[key]; ok {
			result = append(result, key+"="+value)
		}
	}
	return os.WriteFile(path, []byte(strings.Join(result, "\n")+"\n"), 0600)
}
func localIPs() []string {
	result := []string{}
	addrs, _ := net.InterfaceAddrs()
	for _, addr := range addrs {
		ip, _, err := net.ParseCIDR(addr.String())
		if err == nil && ip.To4() != nil && !ip.IsLoopback() && !ip.IsLinkLocalUnicast() {
			result = append(result, ip.String())
		}
	}
	return result
}
func main() {
	if runUpdateHelper() {
		return
	}
	cache, err := os.UserConfigDir()
	if err != nil {
		cache = "."
	}
	root := flag.String("data", filepath.Join(cache, "MundoNaRede"), "pasta de dados")
	port := flag.Int("panel-port", 37654, "porta local do painel")
	noBrowser := flag.Bool("no-browser", false, "exibe o endereço sem abrir o navegador")
	flag.Parse()
	absolute, err := filepath.Abs(*root)
	if err != nil {
		log.Fatal(err)
	}
	if err = os.MkdirAll(filepath.Join(absolute, "servers"), 0700); err != nil {
		log.Fatal(err)
	}
	m := &manager{root: absolute, token: randomHex(32), running: make(map[string]*process), lastRun: make(map[string]time.Time)}
	done := make(chan struct{})
	go m.scheduleLoop(done)
	defer close(done)
	listener, err := net.Listen("tcp4", fmt.Sprintf("127.0.0.1:%d", *port))
	if err != nil {
		log.Fatal(err)
	}
	host := listener.Addr().String()
	m.drive = newDrive(absolute, "http://"+host+"/oauth/google")
	driveCtx, cancelDrive := context.WithCancel(context.Background())
	defer cancelDrive()
	go m.drive.loop(driveCtx)
	mux := http.NewServeMux()
	ui, _ := fs.Sub(assets, "ui")
	mux.Handle("/", http.FileServer(http.FS(ui)))
	mux.HandleFunc("/api/", m.api)
	mux.HandleFunc("/oauth/google", m.drive.callback)
	server := &http.Server{ReadHeaderTimeout: 5 * time.Second, Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Host != host {
			http.Error(w, "Endereço inválido", 403)
			return
		}
		w.Header().Set("Content-Security-Policy", "default-src 'self'; img-src 'self' blob:; frame-ancestors 'none'; base-uri 'none'; form-action 'none'")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		mux.ServeHTTP(w, r)
	})}
	url := "http://" + host + "/#" + m.token
	fmt.Println("Mundo na Rede")
	fmt.Println(url)
	if !*noBrowser {
		switch runtime.GOOS {
		case "windows":
			go openWindowsPanel(url)
		case "android":
			go exec.Command("termux-open-url", url).Run()
		default:
			go exec.Command("xdg-open", url).Run()
		}
	}
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signals
		server.Close()
	}()
	if err = server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Print(err)
	}
	m.mu.Lock()
	var remaining []*process
	for _, proc := range m.running {
		if proc.state != "parado" && proc.state != "falhou" {
			proc.restart = false
			remaining = append(remaining, proc)
			sendCommand(proc, "stop")
		}
	}
	m.mu.Unlock()
	for _, proc := range remaining {
		select {
		case <-proc.done:
		case <-time.After(15 * time.Second):
			proc.cmd.Process.Kill()
			<-proc.done
		}
	}
}
