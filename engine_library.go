package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

const paperAPI = "https://fill.papermc.io/v3/projects/paper"
const paperAgent = "MundoNexus/0.6.0 (https://github.com/emanuelfaustino-spec/mundo-na-rede)"

func (m *manager) cachedPaper() string {
	root := filepath.Join(m.root, "engines", "paper")
	dirs, _ := os.ReadDir(root)
	versions := []string{}
	for _, d := range dirs {
		if d.IsDir() {
			p := filepath.Join(root, d.Name(), "server.jar")
			if info, err := os.Stat(p); err == nil && info.Mode().IsRegular() {
				versions = append(versions, d.Name())
			}
		}
	}
	sort.Slice(versions, func(i, j int) bool { return paperVersionNewer(versions[i], versions[j]) })
	if len(versions) == 0 {
		return ""
	}
	return filepath.Join(root, versions[0], "server.jar")
}

func paperVersionNewer(a, b string) bool {
	aa, bb := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < max(len(aa), len(bb)); i++ {
		av, bv := 0, 0
		if i < len(aa) {
			av, _ = strconv.Atoi(aa[i])
		}
		if i < len(bb) {
			bv, _ = strconv.Atoi(bb[i])
		}
		if av != bv {
			return av > bv
		}
	}
	return a > b
}

func (m *manager) cachedEngineAvailable(p Profile) bool {
	switch p.Engine {
	case "bedrock":
		return len(m.versions()) > 0
	case "java":
		if m.cachedPaper() == "" {
			return false
		}
		_, err := javaExecutable(m.root)
		return err == nil
	default:
		return false
	}
}

func (m *manager) ensureCachedEngine(p Profile, game string) error {
	if hasEngine(game, p.Engine) {
		return nil
	}
	switch p.Engine {
	case "bedrock":
		versions := m.versions()
		if len(versions) == 0 {
			return nil
		}
		if err := copyMissing(filepath.Join(m.root, "engines", "bedrock", versions[0]), game); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(game, ".mnr-engine"), []byte(versions[0]), 0600)
	case "java":
		jar := m.cachedPaper()
		if jar == "" {
			return nil
		}
		from, err := os.Open(jar)
		if err != nil {
			return err
		}
		defer from.Close()
		dest := filepath.Join(game, "server.jar")
		out, err := os.OpenFile(dest, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if errors.Is(err, os.ErrExist) {
			return nil
		}
		if err != nil {
			return err
		}
		_, err = io.Copy(out, io.LimitReader(from, 256<<20))
		if err = errors.Join(err, out.Close()); err != nil {
			os.Remove(dest)
			return err
		}
	}
	return nil
}

func (m *manager) engineLibraryAPI(w http.ResponseWriter, r *http.Request) bool {
	if r.URL.Path != "/api/engines" && r.URL.Path != "/api/engines/prepare" {
		return false
	}
	if r.URL.Path == "/api/engines" && r.Method == http.MethodGet {
		_, javaErr := javaExecutable(m.root)
		writeJSON(w, map[string]any{
			"bedrockReady":          len(m.versions()) > 0,
			"paperReady":            m.cachedPaper() != "",
			"javaReady":             javaErr == nil,
			"needsConsent":          true,
			"portableJavaSupported": runtime.GOOS == "windows",
		})
		return true
	}
	if r.URL.Path != "/api/engines/prepare" || r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return true
	}
	if r.Header.Get("X-Accept-Minecraft-Terms") != "yes" {
		fail(w, errors.New("confirme os termos oficiais antes do download"))
		return true
	}
	m.engineInstallMu.Lock()
	defer m.engineInstallMu.Unlock()
	var err error
	switch r.URL.Query().Get("kind") {
	case "bedrock":
		if len(m.versions()) == 0 {
			var temp string
			temp, err = os.MkdirTemp(m.root, "prepare-bedrock-")
			if err == nil {
				defer os.RemoveAll(temp)
				err = os.MkdirAll(filepath.Join(temp, "game"), 0700)
				if err == nil {
					err = m.installOfficial(r, Profile{ID: "library", Engine: "bedrock"}, filepath.Join(temp, "game"))
				}
			}
		}
	case "paper":
		if m.cachedPaper() == "" {
			err = m.downloadPaper(r.Context())
		}
	case "java":
		err = m.installPortableJava(r.Context())
	default:
		err = errors.New("escolha Bedrock ou Paper")
	}
	if err != nil {
		fail(w, err)
		return true
	}
	_, javaErr := javaExecutable(m.root)
	writeJSON(w, map[string]any{"ok": true, "bedrockReady": len(m.versions()) > 0, "paperReady": m.cachedPaper() != "", "javaReady": javaErr == nil})
	return true
}

func paperGet(ctx context.Context, client *http.Client, address string, destination any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", paperAgent)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("serviço Paper indisponível (%d)", resp.StatusCode)
	}
	return json.NewDecoder(io.LimitReader(resp.Body, 2<<20)).Decode(destination)
}

func (m *manager) downloadPaper(ctx context.Context) error {
	client := &http.Client{Timeout: 4 * time.Minute}
	var versions struct {
		Versions map[string][]string `json:"versions"`
	}
	if err := paperGet(ctx, client, paperAPI, &versions); err != nil {
		return err
	}
	all := []string{}
	for _, v := range versions.Versions {
		all = append(all, v...)
	}
	sort.Slice(all, func(i, j int) bool { return paperVersionNewer(all[i], all[j]) })
	if len(all) == 0 {
		return errors.New("nenhuma versão do Paper disponível")
	}
	for _, version := range all {
		var builds []struct {
			Channel   string `json:"channel"`
			Downloads map[string]struct {
				URL       string `json:"url"`
				Checksums struct {
					SHA256 string `json:"sha256"`
				} `json:"checksums"`
			} `json:"downloads"`
		}
		endpoint := paperAPI + "/versions/" + url.PathEscape(version) + "/builds"
		if err := paperGet(ctx, client, endpoint, &builds); err != nil {
			continue
		}
		for _, build := range builds {
			if build.Channel != "STABLE" {
				continue
			}
			d, exists := build.Downloads["server:default"]
			if !exists || len(d.Checksums.SHA256) != 64 {
				continue
			}
			parsed, parseErr := url.Parse(d.URL)
			if parseErr != nil || parsed.Scheme != "https" || parsed.Host != "fill-data.papermc.io" {
				continue
			}
			dir := filepath.Join(m.root, "engines", "paper", version)
			if err := os.MkdirAll(dir, 0700); err != nil {
				return err
			}
			tmp, err := os.CreateTemp(dir, "paper-*.tmp")
			if err != nil {
				return err
			}
			defer os.Remove(tmp.Name())
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, d.URL, nil)
			if err != nil {
				tmp.Close()
				return err
			}
			req.Header.Set("User-Agent", paperAgent)
			resp, err := client.Do(req)
			if err != nil {
				tmp.Close()
				return err
			}
			if resp.StatusCode != http.StatusOK {
				resp.Body.Close()
				tmp.Close()
				return fmt.Errorf("download do Paper indisponível (%d)", resp.StatusCode)
			}
			hash := sha256.New()
			count, cpErr := io.Copy(io.MultiWriter(tmp, hash), io.LimitReader(resp.Body, (256<<20)+1))
			resp.Body.Close()
			cpErr = errors.Join(cpErr, tmp.Close())
			if cpErr != nil {
				return cpErr
			}
			if count > 256<<20 {
				return errors.New("arquivo Paper muito grande")
			}
			if !strings.EqualFold(hex.EncodeToString(hash.Sum(nil)), d.Checksums.SHA256) {
				return errors.New("o hash SHA-256 do Paper não confere")
			}
			return os.Rename(tmp.Name(), filepath.Join(dir, "server.jar"))
		}
	}
	return errors.New("nenhuma versão estável do Paper disponível no momento")
}
