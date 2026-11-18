package main

import (
	"archive/zip"
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
	"strings"
	"time"
)

const crossplayPort = 19132
const paperVersion = "1.21.11"

func downloadRequest(client *http.Client, ctx context.Context, address string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", address, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "MundoNaRede/development (local Minecraft server manager)")
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != 200 {
		response.Body.Close()
		return nil, fmt.Errorf("%s: HTTP %d", req.URL.Host, response.StatusCode)
	}
	return response, nil
}
func downloadMetadata(client *http.Client, ctx context.Context, address string, dest any) error {
	response, err := downloadRequest(client, ctx, address)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	return json.NewDecoder(io.LimitReader(response.Body, 8<<20)).Decode(dest)
}
func fetchJar(client *http.Client, ctx context.Context, address, dest, host, checksum string) error {
	parsed, err := url.Parse(address)
	if err != nil || parsed.Scheme != "https" || parsed.Host != host {
		return errors.New("origem inesperada do JAR")
	}
	response, err := downloadRequest(client, ctx, address)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	h := sha256.New()
	n, err := io.Copy(io.MultiWriter(f, h), io.LimitReader(response.Body, (180<<20)+1))
	if err = errors.Join(err, f.Close()); err != nil {
		return err
	}
	if n > 180<<20 || (checksum != "" && hex.EncodeToString(h.Sum(nil)) != checksum) {
		return errors.New("falha ao verificar o JAR")
	}
	if filepath.Base(dest) != "server.jar" && !strings.HasPrefix(filepath.Base(dest), "mojang_") {
		return checkExtensionJar(dest, "plugins")
	}
	return nil
}
func cachePaperDependency(client *http.Client, ctx context.Context, dir string) error {
	jar, err := zip.OpenReader(filepath.Join(dir, "server.jar"))
	if err != nil {
		return err
	}
	defer jar.Close()
	entry, err := jar.Open("META-INF/download-context")
	if err != nil {
		return err
	}
	raw, err := io.ReadAll(io.LimitReader(entry, 4096))
	entry.Close()
	if err != nil {
		return err
	}
	fields := strings.Split(strings.TrimSpace(string(raw)), "\t")
	if len(fields) != 3 || len(fields[0]) != 64 || fields[2] != "mojang_"+paperVersion+".jar" {
		return errors.New("dependência inesperada do Paper")
	}
	cache := filepath.Join(dir, "cache")
	if err = os.Mkdir(cache, 0700); err != nil {
		return err
	}
	return fetchJar(client, ctx, fields[1], filepath.Join(cache, fields[2]), "piston-data.mojang.com", fields[0])
}

func (m *manager) installCrossplay(r *http.Request, p Profile, game string) error {
	if p.Engine != "both" || r.Header.Get("X-Accept-BDS-Terms") != "yes" {
		return errors.New("confirme a EULA para instalar o servidor misto")
	}
	entries, err := os.ReadDir(game)
	if err != nil {
		return err
	}
	if len(entries) > 0 {
		return errors.New("use um perfil novo para instalar Java + Bedrock")
	}
	client := &http.Client{Timeout: 2 * time.Minute, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 4 || req.URL.Scheme != "https" || req.URL.Host != via[0].URL.Host {
			return http.ErrUseLastResponse
		}
		return nil
	}}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()
	var builds []struct {
		Channel   string
		Downloads map[string]struct {
			URL       string
			Checksums map[string]string
		}
	}
	if err = downloadMetadata(client, ctx, "https://fill.papermc.io/v3/projects/paper/versions/"+paperVersion+"/builds", &builds); err != nil {
		return err
	}
	temp, err := os.MkdirTemp(m.dir(p.ID), "crossplay-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(temp)
	plugins := filepath.Join(temp, "plugins")
	if err = os.Mkdir(plugins, 0700); err != nil {
		return err
	}
	paperFound := false
	for _, build := range builds {
		asset := build.Downloads["server:default"]
		if build.Channel != "STABLE" || asset.URL == "" || asset.Checksums["sha256"] == "" {
			continue
		}
		if err = fetchJar(client, ctx, asset.URL, filepath.Join(temp, "server.jar"), "fill-data.papermc.io", asset.Checksums["sha256"]); err != nil {
			return err
		}
		paperFound = true
		break
	}
	if !paperFound {
		return errors.New("nenhuma versão estável do Paper disponível")
	}
	if err = cachePaperDependency(client, ctx, temp); err != nil {
		return err
	}
	for _, plugin := range []struct{ project, name string }{
		{"geyser", "Geyser-Spigot.jar"}, {"floodgate", "Floodgate-Spigot.jar"},
	} {
		address := "https://download.geysermc.org/v2/projects/" + plugin.project + "/versions/latest/builds/latest/downloads/spigot"
		if err = fetchJar(client, ctx, address, filepath.Join(plugins, plugin.name), "download.geysermc.org", ""); err != nil {
			return err
		}
	}
	var versions []struct {
		Type  string `json:"version_type"`
		Files []struct {
			URL     string
			Primary bool
		}
	}
	params := url.Values{"loaders": {`["paper"]`}, "game_versions": {`["` + paperVersion + `"]`}}
	if err = downloadMetadata(client, ctx, "https://api.modrinth.com/v2/project/viaversion/version?"+params.Encode(), &versions); err != nil {
		return err
	}
	var viaURL string
	for _, version := range versions {
		if version.Type != "release" {
			continue
		}
		for _, file := range version.Files {
			if file.Primary || len(version.Files) == 1 {
				viaURL = file.URL
				break
			}
		}
		if viaURL != "" {
			break
		}
	}
	if viaURL == "" {
		return errors.New("ViaVersion compatível indisponível")
	}
	if err = fetchJar(client, ctx, viaURL, filepath.Join(plugins, "ViaVersion.jar"), "cdn.modrinth.com", ""); err != nil {
		return err
	}
	if err = os.WriteFile(filepath.Join(temp, "eula.txt"), []byte("eula=true\n"), 0600); err != nil {
		return err
	}
	return replaceWithBackup(temp, game, "instalacao-anterior")
}
