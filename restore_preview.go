package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type restorePreview struct {
	Token   string `json:"token"`
	World   bool   `json:"world"`
	Config  bool   `json:"config"`
	Engine  bool   `json:"engine"`
	Version string `json:"version"`
	Packs   int    `json:"packs"`
	Size    int64  `json:"size"`
}

func (m *manager) stageRestore(w http.ResponseWriter, r *http.Request, p Profile) (restorePreview, error) {
	info := restorePreview{Token: randomHex(8)}
	root := filepath.Join(m.dir(p.ID), "restore-preview")
	if err := os.RemoveAll(root); err != nil {
		return info, err
	}
	if err := os.MkdirAll(root, 0700); err != nil {
		return info, err
	}
	ready := false
	defer func() {
		if !ready {
			os.RemoveAll(root)
		}
	}()
	zipPath := filepath.Join(root, "backup.zip")
	file, err := os.Create(zipPath)
	if err != nil {
		return info, err
	}
	m.progress.Store(p.ID, "Recebendo backup")
	size, err := io.Copy(file, http.MaxBytesReader(w, r.Body, maxArchive))
	if err = errors.Join(err, file.Close()); err != nil {
		return info, err
	}
	info.Size = size
	m.progress.Store(p.ID, "Validando backup")
	game := filepath.Join(root, "game")
	if err = unpack(zipPath, game); err != nil {
		return info, err
	}
	if err = os.Remove(zipPath); err != nil {
		return info, err
	}
	regular := func(name string) bool {
		v, e := os.Stat(filepath.Join(game, name))
		return e == nil && v.Mode().IsRegular()
	}
	info.Engine = hasEngine(game, p.Engine)
	if p.Engine == "bedrock" && info.Engine {
		if err = checkBDS(game); err != nil {
			return info, errors.New("o motor incluído é de outra plataforma")
		}
	}
	info.Config = regular("server.properties") || regular("pnx.yml")
	info.World = regular("level.dat") || regular(filepath.Join("Mundo", "level.dat"))
	_ = filepath.WalkDir(filepath.Join(game, "worlds"), func(path string, entry fs.DirEntry, err error) error {
		if err == nil && !entry.IsDir() && entry.Name() == "level.dat" {
			info.World = true
			return fs.SkipAll
		}
		return nil
	})
	if info.World && regular("level.dat") && p.Engine == "bedrock" {
		world := filepath.Join(root, "world")
		if err = os.Rename(game, world); err != nil {
			return info, err
		}
		if err = os.MkdirAll(filepath.Join(game, "worlds"), 0700); err != nil {
			return info, err
		}
		if err = os.Rename(world, filepath.Join(game, "worlds", "Mundo")); err != nil {
			return info, err
		}
		info.Config, info.Engine = false, false
	}
	data, _ := os.ReadFile(filepath.Join(game, ".mnr-engine"))
	if bedrockVersion.Match(data) {
		info.Version = string(data)
	}
	packs, _ := listExtensions(game, p.Engine, false)
	info.Packs = len(packs)
	if !info.World && !info.Config && !info.Engine {
		return info, errors.New("o ZIP não contém um mundo, configuração ou motor reconhecido")
	}
	if p.Engine != "bedrock" && !info.Engine {
		return info, errors.New("o backup deste tipo de servidor precisa incluir o motor")
	}
	data, _ = json.Marshal(info)
	if err = os.WriteFile(filepath.Join(root, "preview.json"), data, 0600); err != nil {
		return info, err
	}
	ready = true
	return info, nil
}
func (m *manager) applyRestore(w http.ResponseWriter, r *http.Request, p Profile, game string) error {
	var input struct{ Token, Version string }
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8192)).Decode(&input); err != nil {
		return err
	}
	root := filepath.Join(m.dir(p.ID), "restore-preview")
	preview := filepath.Join(root, "preview.json")
	st, err := os.Stat(preview)
	if err != nil || time.Since(st.ModTime()) > time.Hour {
		return errors.New("a validação expirou. Selecione o ZIP de novo")
	}
	bytes, err := os.ReadFile(preview)
	if err != nil {
		return err
	}
	var info restorePreview
	if json.Unmarshal(bytes, &info) != nil || input.Token != info.Token {
		return errors.New("a validação do backup não corresponde à seleção")
	}
	extracted := filepath.Join(root, "game")
	m.progress.Store(p.ID, "Restaurando arquivos")
	if !info.Engine && p.Engine == "bedrock" && input.Version != "" {
		if input.Version == "legacy" {
			source := m.legacyEngine(p.ID)
			if source == "" {
				return errors.New("nenhum motor Bedrock foi encontrado nos outros perfis")
			}
			if err = copyBedrockRuntime(source, extracted); err != nil {
				return err
			}
		} else {
			if !bedrockVersion.MatchString(input.Version) {
				return errors.New("versão do BDS inválida")
			}
			library := filepath.Join(m.root, "engines", "bedrock", input.Version)
			if !hasEngine(library, "bedrock") {
				return errors.New("versão do BDS indisponível na biblioteca")
			}
			if err = copyMissing(library, extracted); err != nil {
				return err
			}
			if err = os.WriteFile(filepath.Join(extracted, ".mnr-engine"), []byte(input.Version), 0600); err != nil {
				return err
			}
		}
	}
	if p.Engine == "bedrock" && !info.Engine && input.Version == "" && hasEngine(game, "bedrock") {
		if err = copyBedrockRuntime(game, extracted); err != nil {
			return err
		}
	}
	if err = replaceWithBackup(extracted, game, "servidor-anterior"); err != nil {
		return err
	}
	os.RemoveAll(root)
	return nil
}
func copyBedrockRuntime(source, dest string) error {
	entries, err := os.ReadDir(source)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			if name == "resource_packs" || name == "behavior_packs" {
				packs, err := os.ReadDir(filepath.Join(source, name))
				if err != nil {
					return err
				}
				if err := os.MkdirAll(filepath.Join(dest, name), 0700); err != nil {
					return err
				}
				for _, pack := range packs {
					if strings.HasPrefix(pack.Name(), "mnr-") || strings.HasPrefix(pack.Name(), "pacote-anterior-") {
						continue
					}
					if err := copyMissing(filepath.Join(source, name, pack.Name()), filepath.Join(dest, name, pack.Name())); err != nil {
						return err
					}
				}
			}
			if name == "definitions" || name == "structures" || name == "treatments" {
				if err = copyMissing(filepath.Join(source, name), filepath.Join(dest, name)); err != nil {
					return err
				}
			}
			continue
		}
		ext := strings.ToLower(filepath.Ext(name))
		if name == "bedrock_server" || name == "bedrock_server.exe" || name == ".mnr-engine" || ext == ".dll" || ext == ".so" {
			if err = copyMissing(filepath.Join(source, name), filepath.Join(dest, name)); err != nil {
				return fmt.Errorf("%s: %w", name, err)
			}
		}
	}
	return nil
}
