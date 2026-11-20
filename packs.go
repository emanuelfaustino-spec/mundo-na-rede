package main

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var packUUID = regexp.MustCompile("(?i)^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$")

type packReference struct {
	ID      string          `json:"pack_id"`
	Version json.RawMessage `json:"version"`
}
type bedrockPack struct {
	Folders []string
	Path    string
	Ref     packReference
}

func readPack(dir string) (bedrockPack, error) {
	var out bedrockPack
	out.Path = dir
	data, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		return out, err
	}
	var manifest struct {
		Header struct {
			UUID    string
			Version json.RawMessage
		}
		Modules []struct{ Type string }
	}
	if err = json.Unmarshal(data, &manifest); err != nil {
		return out, err
	}
	if !packUUID.MatchString(manifest.Header.UUID) {
		return out, errors.New("UUID inválido no manifest.json")
	}
	var ver []uint
	if err = json.Unmarshal(manifest.Header.Version, &ver); err != nil || len(ver) != 3 {
		return out, errors.New("use um pacote com versão composta de três números")
	}
	out.Ref = packReference{ID: strings.ToLower(manifest.Header.UUID), Version: manifest.Header.Version}
	for _, module := range manifest.Modules {
		switch module.Type {
		case "resources":
			out.addFolder("resource_packs")
		case "data", "script":
			out.addFolder("behavior_packs")
		}
	}
	if len(out.Folders) == 0 {
		return out, errors.New("nenhum módulo de recursos ou comportamento no manifest.json")
	}
	return out, nil
}

func (p *bedrockPack) addFolder(folder string) {
	for _, existing := range p.Folders {
		if existing == folder {
			return
		}
	}
	p.Folders = append(p.Folders, folder)
}

func (p bedrockPack) hasFolder(folder string) bool {
	for _, existing := range p.Folders {
		if existing == folder {
			return true
		}
	}
	return false
}
func expandNestedPacks(root string) error {
	entries, err := os.ReadDir(root)
	if err != nil {
		return err
	}
	if _, err = os.Stat(filepath.Join(root, "manifest.json")); err == nil {
		return nil
	}
	var total uint64
	for _, entry := range entries {
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".mcpack") {
			continue
		}
		archive := filepath.Join(root, entry.Name())
		zr, err := zip.OpenReader(archive)
		if err != nil {
			return err
		}
		for _, file := range zr.File {
			if file.UncompressedSize64 > uint64(maxArchive)-total {
				zr.Close()
				return errors.New("pacotes internos excedem 2 GiB")
			}
			total += file.UncompressedSize64
		}
		zr.Close()
		dest, err := os.MkdirTemp(root, "nested-pack-")
		if err != nil {
			return err
		}
		if err = unpack(archive, dest); err != nil {
			return err
		}
	}
	return nil
}
func installBedrockAddon(src, game, folder string) error {
	if err := expandNestedPacks(src); err != nil {
		return err
	}
	packs := []bedrockPack{}
	err := filepath.WalkDir(src, func(dir string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			return nil
		}
		if _, err = os.Stat(filepath.Join(dir, "manifest.json")); errors.Is(err, os.ErrNotExist) {
			return nil
		}
		pack, err := readPack(dir)
		if err != nil {
			return err
		}
		if folder != "auto" && !pack.hasFolder(folder) {
			return errors.New("o tipo do pacote não corresponde à seleção")
		}
		packs = append(packs, pack)
		return filepath.SkipDir
	})
	if err != nil {
		return err
	}
	if len(packs) == 0 {
		return errors.New("nenhum manifest.json encontrado no add-on")
	}
	seen := map[string]bool{}
	for _, pack := range packs {
		if seen[pack.Ref.ID] {
			return errors.New("o add-on contém UUIDs duplicados")
		}
		seen[pack.Ref.ID] = true
	}
	for _, pack := range packs {
		installed := pack.Path
		for index, folder := range pack.Folders {
			dest := filepath.Join(game, folder, "mnr-"+pack.Ref.ID)
			if err = os.MkdirAll(filepath.Dir(dest), 0700); err != nil {
				return err
			}
			source := installed
			if index > 0 {
				source = filepath.Join(filepath.Dir(pack.Path), "pack-copy-"+randomHex(4))
				if err = copyMissing(installed, source); err != nil {
					return err
				}
			}
			if err = replaceWithBackup(source, dest, "pacote-anterior"); err != nil {
				return err
			}
			installed = dest
		}
	}
	return nil
}
func attachBedrockPacks(game string) error {
	for _, group := range []struct{ folder, index string }{
		{"resource_packs", "world_resource_packs.json"}, {"behavior_packs", "world_behavior_packs.json"},
	} {
		root := filepath.Join(game, group.folder)
		entries, err := os.ReadDir(root)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return err
		}
		var packs []packReference
		for _, entry := range entries {
			if !entry.IsDir() || !strings.HasPrefix(entry.Name(), "mnr-") {
				continue
			}
			pack, err := readPack(filepath.Join(root, entry.Name()))
			if err != nil {
				return err
			}
			packs = append(packs, pack.Ref)
		}
		if len(packs) == 0 {
			continue
		}
		world := filepath.Join(game, "worlds", "Mundo")
		if err = os.MkdirAll(world, 0700); err != nil {
			return err
		}
		index := filepath.Join(world, group.index)
		data, err := os.ReadFile(index)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		var refs []map[string]json.RawMessage
		if len(data) > 0 {
			if err = json.Unmarshal(data, &refs); err != nil {
				return err
			}
		}
		for _, pack := range packs {
			id, _ := json.Marshal(pack.ID)
			found := false
			for _, ref := range refs {
				var existing string
				json.Unmarshal(ref["pack_id"], &existing)
				if strings.EqualFold(existing, pack.ID) {
					ref["version"] = pack.Version
					found = true
					break
				}
			}
			if !found {
				refs = append(refs, map[string]json.RawMessage{"pack_id": id, "version": pack.Version})
			}
		}
		data, err = json.MarshalIndent(refs, "", "  ")
		if err != nil {
			return err
		}
		if err = os.WriteFile(index, data, 0600); err != nil {
			return err
		}
	}
	return nil
}
func checkExtensionJar(path, folder string) error {
	jar, err := zip.OpenReader(path)
	if err != nil {
		return errors.New("JAR inválido")
	}
	defer jar.Close()
	for _, entry := range jar.File {
		switch entry.Name {
		case "plugin.yml", "paper-plugin.yml":
			if folder == "plugins" {
				return nil
			}
		case "fabric.mod.json", "quilt.mod.json", "META-INF/mods.toml", "META-INF/neoforge.mods.toml", "mcmod.info":
			if folder == "mods" {
				return nil
			}
		}
	}
	return errors.New("o JAR não contém um manifesto reconhecido para esse tipo de extensão")
}

func (m *manager) removeBedrockPack(w http.ResponseWriter, r *http.Request, p Profile, game string) error {
	var input struct{ ID, Kind string }
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&input); err != nil {
		return err
	}
	if p.Engine != "bedrock" || !packUUID.MatchString(input.ID) {
		return errors.New("selecione um add-on importado")
	}
	indexName := ""
	switch input.Kind {
	case "resource_packs":
		indexName = "world_resource_packs.json"
	case "behavior_packs":
		indexName = "world_behavior_packs.json"
	default:
		return errors.New("tipo de pacote inválido")
	}
	source := filepath.Join(game, input.Kind, "mnr-"+strings.ToLower(input.ID))
	if _, err := readPack(source); err != nil {
		return err
	}
	index := filepath.Join(game, "worlds", "Mundo", indexName)
	raw, err := os.ReadFile(index)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	var refs []map[string]json.RawMessage
	if len(raw) > 0 {
		if err = json.Unmarshal(raw, &refs); err != nil {
			return err
		}
	}
	keep := []map[string]json.RawMessage{}
	for _, ref := range refs {
		var id string
		json.Unmarshal(ref["pack_id"], &id)
		if !strings.EqualFold(id, input.ID) {
			keep = append(keep, ref)
		}
	}
	removed := filepath.Join(m.dir(p.ID), "removed-packs")
	if err = os.MkdirAll(removed, 0700); err != nil {
		return err
	}
	target := filepath.Join(removed, filepath.Base(source)+"-"+randomHex(4))
	if err = os.Rename(source, target); err != nil {
		return err
	}
	if len(keep) == len(refs) {
		return nil
	}
	contents, _ := json.MarshalIndent(keep, "", "  ")
	temp, err := os.CreateTemp(filepath.Dir(index), "pack-index-")
	if err == nil {
		defer os.Remove(temp.Name())
		_, err = temp.Write(contents)
		if err = errors.Join(err, temp.Close()); err == nil {
			err = os.Rename(temp.Name(), index)
		}
	}
	if err != nil {
		return errors.Join(err, os.Rename(target, source))
	}
	return nil
}
