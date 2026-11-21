package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const mojangLinks = "https://net-secondary.web.minecraft-services.net/api/v1.0/download/links"

func (m *manager) installOfficial(r *http.Request, p Profile, game string) error {
	if p.Engine != "bedrock" || (runtime.GOOS != "windows" && runtime.GOOS != "linux") {
		return errors.New("instalação oficial disponível no Windows e no Ubuntu")
	}
	if !acceptedBDSTerms(r) {
		return errors.New("confirme os termos do download oficial antes de instalar")
	}
	if hasEngine(game, "bedrock") {
		return errors.New("este perfil já tem um motor Bedrock")
	}
	defer m.progress.Delete(p.ID)
	m.progress.Store(p.ID, "Verificando biblioteca")
	library := filepath.Join(m.root, "engines", "bedrock")
	if err := os.MkdirAll(library, 0700); err != nil {
		return err
	}
	version := r.URL.Query().Get("version")
	if version == "" {
		if versions := m.versions(); len(versions) > 0 {
			version = versions[0]
		}
	}
	if version == "legacy" {
		source := m.legacyEngine(p.ID)
		if source == "" {
			return errors.New("nenhum motor Bedrock foi encontrado nos outros perfis")
		}
		return copyBedrockRuntime(source, game)
	}
	if version != "" {
		if !bedrockVersion.MatchString(version) || !hasEngine(filepath.Join(library, version), "bedrock") {
			return errors.New("versão do BDS indisponível na biblioteca")
		}
		if err := copyMissing(filepath.Join(library, version), game); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(game, ".mnr-engine"), []byte(version), 0600)
	}
	kind, prefix := "serverBedrockLinux", "/bedrockdedicatedserver/bin-linux/"
	if runtime.GOOS == "windows" {
		kind, prefix = "serverBedrockWindows", "/bedrockdedicatedserver/bin-win/"
	}
	client := &http.Client{Timeout: 4 * time.Minute, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	get := func(address string) (*http.Response, error) {
		req, err := http.NewRequestWithContext(r.Context(), "GET", address, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", "Mozilla/5.0")
		response, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		if response.StatusCode != http.StatusOK {
			response.Body.Close()
			return nil, fmt.Errorf("download oficial indisponível (%d). Use a importação manual", response.StatusCode)
		}
		return response, nil
	}
	metadata, err := get(mojangLinks)
	if err != nil {
		return err
	}
	var listing struct {
		Result struct {
			Links []struct {
				Type string `json:"downloadType"`
				URL  string `json:"downloadUrl"`
			} `json:"links"`
		} `json:"result"`
	}
	err = json.NewDecoder(io.LimitReader(metadata.Body, 1<<20)).Decode(&listing)
	metadata.Body.Close()
	if err != nil {
		return err
	}
	var download string
	for _, link := range listing.Result.Links {
		if link.Type == kind {
			download = link.URL
			break
		}
	}
	parsed, err := url.Parse(download)
	if err != nil || parsed.Scheme != "https" || parsed.Host != "www.minecraft.net" || !strings.HasPrefix(parsed.Path, prefix) || !strings.HasSuffix(parsed.Path, ".zip") {
		return errors.New("endereço oficial inesperado. Use a importação manual")
	}
	version = strings.TrimSuffix(strings.TrimPrefix(filepath.Base(parsed.Path), "bedrock-server-"), ".zip")
	if !bedrockVersion.MatchString(version) {
		return errors.New("versão oficial do BDS inválida")
	}
	m.progress.Store(p.ID, "Baixando BDS")
	response, err := get(download)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	temp, err := os.MkdirTemp(library, "download-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(temp)
	archive := filepath.Join(temp, "bedrock.zip")
	output, err := os.Create(archive)
	if err != nil {
		return err
	}
	size, err := io.Copy(output, io.LimitReader(response.Body, maxArchive+1))
	if err = errors.Join(err, output.Close()); err != nil {
		return err
	}
	if size > maxArchive {
		return errors.New("o arquivo oficial excede 2 GiB")
	}
	extracted := filepath.Join(temp, "game")
	if err = os.Mkdir(extracted, 0700); err != nil {
		return err
	}
	m.progress.Store(p.ID, "Extraindo arquivos")
	if err = unpack(archive, extracted); err != nil {
		return err
	}
	if err = checkBDS(extracted); err != nil {
		return err
	}
	cached := filepath.Join(library, version)
	if err = os.Rename(extracted, cached); err != nil {
		return err
	}
	if err = copyMissing(cached, game); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(game, ".mnr-engine"), []byte(version), 0600)
}
