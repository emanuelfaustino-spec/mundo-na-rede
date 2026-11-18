package main

import (
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

var localBackupName = regexp.MustCompile(`^mundo-na-rede-[0-9]{8}-[0-9]{6}-[a-f0-9]{6}\.zip$`)

func (m *manager) backupHistory(w http.ResponseWriter, r *http.Request, p Profile) {
	dir := filepath.Join(m.dir(p.ID), "backups")
	name := r.URL.Query().Get("name")
	if name != "" {
		if !localBackupName.MatchString(name) {
			http.NotFound(w, r)
			return
		}
		if r.Method == "POST" && r.URL.Query().Get("upload") == "1" {
			if err := m.drive.queueLocalBackup(p, filepath.Join(dir, name), name); err != nil {
				fail(w, err)
				return
			}
			writeJSON(w, map[string]bool{"ok": true})
			return
		}
		if r.Method == "POST" {
			if err := os.Remove(filepath.Join(dir, name)); err != nil {
				fail(w, err)
				return
			}
			writeJSON(w, map[string]bool{"ok": true})
			return
		}
		if r.Method == "GET" {
			file, err := os.Open(filepath.Join(dir, name))
			if err != nil {
				http.NotFound(w, r)
				return
			}
			defer file.Close()
			w.Header().Set("Content-Type", "application/zip")
			w.Header().Set("Content-Disposition", `attachment; filename="`+name+`"`)
			http.ServeContent(w, r, name, time.Time{}, file)
			return
		}
	}
	if r.Method != "GET" {
		w.WriteHeader(405)
		return
	}
	entries, err := os.ReadDir(dir)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		fail(w, err)
		return
	}
	type item struct {
		Name string    `json:"name"`
		Size int64     `json:"size"`
		Date time.Time `json:"date"`
	}
	list := []item{}
	for _, entry := range entries {
		if !localBackupName.MatchString(entry.Name()) || !entry.Type().IsRegular() {
			continue
		}
		if st, err := entry.Info(); err == nil {
			list = append(list, item{entry.Name(), st.Size(), st.ModTime()})
		}
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Date.After(list[j].Date) })
	writeJSON(w, list)
}

func (d *driveBackup) queueLocalBackup(p Profile, path, name string) error {
	d.mu.Lock()
	if !d.cloudReady() || d.account.Email == "" {
		d.mu.Unlock()
		return errors.New("conecte o Google Drive primeiro")
	}
	queue := filepath.Join(d.root, "drive-queue", d.accountKey())
	d.mu.Unlock()
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	if err = os.MkdirAll(queue, 0700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(queue, "history-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	_, err = io.Copy(tmp, file)
	if err = errors.Join(err, tmp.Close()); err != nil {
		return err
	}
	dest := filepath.Join(queue, "MundoNaRede-"+p.Engine+"-"+p.ID+"-"+strings.TrimPrefix(name, "mundo-na-rede-"))
	if err = os.Rename(tmp.Name(), dest); err != nil {
		return err
	}
	d.mu.Lock()
	d.status = "pending"
	d.mu.Unlock()
	select {
	case d.queued <- struct{}{}:
	default:
	}
	return nil
}
