package main

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func unpack(input, target string) error {
	zr, err := zip.OpenReader(input)
	if err != nil {
		return errors.New("arquivo ZIP inválido")
	}
	defer zr.Close()
	var expanded uint64
	if len(zr.File) > 100000 {
		return errors.New("arquivo contém entradas demais")
	}
	for _, entry := range zr.File {
		name := strings.TrimSuffix(entry.Name, "/")
		if !fs.ValidPath(name) || name == "." || strings.ContainsAny(name, "\\:") {
			return errors.New("ZIP contém caminho inválido")
		}
		for _, part := range strings.Split(name, "/") {
			if strings.TrimRight(part, " .") != part || !filepath.IsLocal(part) {
				return errors.New("ZIP contém nome de arquivo inválido")
			}
		}
		dest := filepath.Join(target, filepath.FromSlash(name))
		if entry.FileInfo().IsDir() {
			if err = os.MkdirAll(dest, 0700); err != nil {
				return err
			}
			continue
		}
		if !entry.Mode().IsRegular() {
			return errors.New("ZIP contém um item que não é um arquivo comum")
		}
		if entry.UncompressedSize64 > uint64(3*maxArchive)-expanded {
			return errors.New("ZIP excede 6 GiB descompactado")
		}
		expanded += entry.UncompressedSize64
		if err = os.MkdirAll(filepath.Dir(dest), 0700); err != nil {
			return err
		}
		if err = extractEntry(entry, dest); err != nil {
			return err
		}
	}
	return nil
}
func extractEntry(entry *zip.File, destination string) error {
	reader, err := entry.Open()
	if err != nil {
		return err
	}
	defer reader.Close()
	writer, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(writer, reader)
	return errors.Join(copyErr, writer.Close())
}
func replaceWithBackup(source, dest, label string) error {
	if _, err := os.Stat(dest); errors.Is(err, os.ErrNotExist) {
		return os.Rename(source, dest)
	} else if err != nil {
		return err
	}
	previous := filepath.Join(filepath.Dir(dest), label+"-"+time.Now().UTC().Format("20060102-150405")+"-"+randomHex(3))
	if err := os.Rename(dest, previous); err != nil {
		return err
	}
	if err := os.Rename(source, dest); err != nil {
		return errors.Join(err, os.Rename(previous, dest))
	}
	return nil
}
func (m *manager) changeFiles(w http.ResponseWriter, r *http.Request, p Profile, game, action string) error {
	dir := m.dir(p.ID)
	tmp, err := os.MkdirTemp(dir, "upload-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	input := filepath.Join(tmp, "upload")
	limit := int64(maxArchive)
	if action == "config" {
		limit = 256 << 10
	}
	file, err := os.Create(input)
	if err != nil {
		return err
	}
	written, copyErr := io.Copy(file, http.MaxBytesReader(w, r.Body, limit))
	if err = errors.Join(copyErr, file.Close()); err != nil {
		return err
	}
	if written == 0 {
		return errors.New("o arquivo está vazio")
	}
	switch action {
	case "config":
		name := "server.properties"
		if p.Engine == "pnx" {
			name = "pnx.yml"
		}
		return replaceWithBackup(input, filepath.Join(game, name), "config-anterior")
	case "addon":
		folder := r.URL.Query().Get("folder")
		allowed := map[string]map[string]bool{
			"bedrock": {"auto": true, "resource_packs": true, "behavior_packs": true},
			"pnx":     {"resource_packs": true, "plugins": true},
			"java":    {"plugins": true, "mods": true},
			"both":    {"plugins": true},
		}
		if !allowed[p.Engine][folder] {
			return errors.New("pasta incompatível com este motor")
		}
		name := r.URL.Query().Get("name")
		if name == "." || !filepath.IsLocal(name) || strings.ContainsAny(name, "/\\:") || strings.TrimRight(name, " .") != name {
			return errors.New("nome de arquivo inválido")
		}
		if folder == "plugins" || folder == "mods" {
			if !strings.EqualFold(filepath.Ext(name), ".jar") {
				return errors.New("selecione um JAR confiável")
			}
			if err = checkExtensionJar(input, folder); err != nil {
				return err
			}
			if err = os.MkdirAll(filepath.Join(game, folder), 0700); err != nil {
				return err
			}
			return replaceWithBackup(input, filepath.Join(game, folder, name), "extensao-anterior")
		}
		extracted := filepath.Join(tmp, "pack")
		if err = unpack(input, extracted); err != nil {
			return err
		}
		if p.Engine == "bedrock" {
			return installBedrockAddon(extracted, game, folder)
		}
		if _, err = os.Stat(filepath.Join(extracted, "manifest.json")); err != nil {
			return errors.New("o pacote precisa conter manifest.json na raiz")
		}
		name = strings.TrimSuffix(name, filepath.Ext(name))
		if name == "" || name == "." || name == ".." || strings.TrimRight(name, " .") != name || !filepath.IsLocal(name) {
			return errors.New("nome de pacote inválido")
		}
		if err = os.MkdirAll(filepath.Join(game, folder), 0700); err != nil {
			return err
		}
		return replaceWithBackup(extracted, filepath.Join(game, folder, name), "pacote-anterior")
	case "install":
		if p.Engine == "both" {
			return errors.New("use o instalador do servidor misto")
		}
		if p.Engine != "bedrock" {
			zr, err := zip.OpenReader(input)
			if err != nil {
				return errors.New("JAR inválido")
			}
			zr.Close()
			if err = os.MkdirAll(game, 0700); err != nil {
				return err
			}
			return replaceWithBackup(input, filepath.Join(game, "server.jar"), "motor-anterior")
		}
	}
	extracted := filepath.Join(tmp, "extracted")
	if err = os.Mkdir(extracted, 0700); err != nil {
		return err
	}
	if err = unpack(input, extracted); err != nil {
		return err
	}
	switch action {
	case "install":
		entries, err := os.ReadDir(game)
		if err != nil {
			return err
		}
		if len(entries) > 0 {
			return errors.New("importe o BDS em um perfil novo. Faça backup antes de migrar o mundo para uma nova versão.")
		}
		if err = checkBDS(extracted); err != nil {
			return err
		}
		return replaceWithBackup(extracted, game, "instalacao-anterior")
	case "world":
		if p.Engine != "bedrock" {
			return errors.New("a importação pela interface está disponível para mundos Bedrock oficiais")
		}
		if _, err = os.Stat(filepath.Join(extracted, "level.dat")); err != nil {
			return errors.New("o ZIP ou .mcworld precisa conter level.dat e db/ na raiz")
		}
		if db, err := os.Stat(filepath.Join(extracted, "db")); err != nil || !db.IsDir() {
			return errors.New("pasta db/ não encontrada")
		}
		worlds := filepath.Join(game, "worlds")
		if err = os.MkdirAll(worlds, 0700); err != nil {
			return err
		}
		return replaceWithBackup(extracted, filepath.Join(worlds, "Mundo"), "mundo-anterior")
	case "restore":
		return restoreGame(p, extracted, game)
	}
	return errors.New("ação desconhecida")
}
func restoreGame(p Profile, extracted, game string) error {
	if p.Engine == "bedrock" {
		if err := checkBDS(extracted); err != nil {
			return err
		}
	} else if _, err := os.Stat(filepath.Join(extracted, "server.jar")); err != nil {
		return errors.New("backup sem server.jar")
	}
	return replaceWithBackup(extracted, game, "servidor-anterior")
}

func checkBDS(dir string) error {
	exe := "bedrock_server"
	if runtime.GOOS == "windows" {
		exe += ".exe"
	}
	p := filepath.Join(dir, exe)
	info, err := os.Stat(p)
	if err != nil || !info.Mode().IsRegular() {
		return fmt.Errorf("arquivo %s não encontrado na raiz do ZIP", exe)
	}
	return os.Chmod(p, 0700)
}
func (m *manager) backup(w http.ResponseWriter, game, id string) {
	folder := filepath.Join(m.dir(id), "backups")
	if err := os.MkdirAll(folder, 0700); err != nil {
		fail(w, err)
		return
	}
	name := "mundo-na-rede-" + time.Now().UTC().Format("20060102-150405") + "-" + randomHex(3) + ".zip"
	archive, err := os.OpenFile(filepath.Join(folder, name), os.O_CREATE|os.O_EXCL|os.O_RDWR, 0600)
	if err != nil {
		fail(w, err)
		return
	}
	defer archive.Close()
	if err = zipGame(archive, game); err != nil {
		archive.Close()
		os.Remove(archive.Name())
		fail(w, err)
		return
	}
	if err = pruneLocalBackups(folder); err != nil {
		fail(w, err)
		return
	}
	if _, err = archive.Seek(0, io.SeekStart); err != nil {
		fail(w, err)
		return
	}
	w.Header().Set("Content-Disposition", `attachment; filename="`+name+`"`)
	w.Header().Set("Content-Type", "application/zip")
	http.ServeContent(w, &http.Request{Method: "GET", Header: http.Header{}}, name, time.Time{}, archive)
}

func zipGame(archive *os.File, game string) error {
	var err error
	zw := zip.NewWriter(archive)
	err = filepath.WalkDir(game, func(file string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("arquivo especial: %s", entry.Name())
		}
		relative, err := filepath.Rel(game, file)
		if err != nil {
			return err
		}
		target, err := zw.Create(path.Clean(filepath.ToSlash(relative)))
		if err != nil {
			return err
		}
		source, err := os.Open(file)
		if err != nil {
			return err
		}
		_, err = io.Copy(target, source)
		return errors.Join(err, source.Close())
	})
	if err = errors.Join(err, zw.Close()); err != nil {
		return err
	}
	info, err := archive.Stat()
	if err != nil {
		return err
	}
	if info.Size() > maxArchive {
		return errors.New("backup exceeds 2 GiB")
	}
	return nil
}
