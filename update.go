package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// The Contents API is used instead of raw.githubusercontent.com because the
// latter can keep a branch file in a CDN cache for longer than a release check.
const updateManifestURL = "https://api.github.com/repos/emanuelfaustino-spec/mundo-na-rede-updates/contents/manifest.json"

type updateManifest struct {
	Version string `json:"version"`
	Notes   string `json:"notes"`
	URL     string `json:"url"`
	SHA256  string `json:"sha256"`
}

func versionNewer(candidate, current string) bool {
	parse := func(v string) ([3]int, bool) {
		var out [3]int
		parts := strings.Split(strings.TrimPrefix(strings.TrimSpace(v), "v"), ".")
		if len(parts) != 3 {
			return out, false
		}
		for i, part := range parts {
			n, err := strconv.Atoi(part)
			if err != nil || n < 0 {
				return out, false
			}
			out[i] = n
		}
		return out, true
	}
	a, okA := parse(candidate)
	b, okB := parse(current)
	if !okA || !okB {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return a[i] > b[i]
		}
	}
	return false
}

func fetchUpdate() (updateManifest, error) {
	var result updateManifest
	client := http.Client{Timeout: 12 * time.Second}
	request, err := http.NewRequest(http.MethodGet, updateManifestURL, nil)
	if err != nil {
		return result, err
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("Cache-Control", "no-cache")
	response, err := client.Do(request)
	if err != nil {
		return result, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return result, fmt.Errorf("canal de atualização respondeu %d", response.StatusCode)
	}
	var envelope struct {
		Content  string `json:"content"`
		Encoding string `json:"encoding"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 64<<10)).Decode(&envelope); err != nil {
		return result, err
	}
	if envelope.Encoding != "base64" {
		return result, errors.New("manifesto de atualização inválido")
	}
	data, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(envelope.Content, "\n", ""))
	if err != nil {
		return result, err
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return result, err
	}
	if !versionNewer(result.Version, appVersion) {
		return updateManifest{}, nil
	}
	if !strings.HasPrefix(result.URL, "https://") || len(result.SHA256) != 64 {
		return updateManifest{}, errors.New("manifesto de atualização inválido")
	}
	if _, err := hex.DecodeString(result.SHA256); err != nil {
		return updateManifest{}, errors.New("hash de atualização inválido")
	}
	return result, nil
}

func (m *manager) updateAPI(w http.ResponseWriter, r *http.Request) bool {
	if distributionChannel == "store" {
		switch {
		case r.URL.Path == "/api/update" && r.Method == http.MethodGet:
			writeJSON(w, map[string]any{"available": false, "version": appVersion, "managedByStore": true})
			return true
		case r.URL.Path == "/api/update/apply" && r.Method == http.MethodPost:
			fail(w, errors.New("atualizações desta instalação são gerenciadas pela Microsoft Store"))
			return true
		}
	}
	switch {
	case r.URL.Path == "/api/update" && r.Method == http.MethodGet:
		update, err := fetchUpdate()
		if err != nil {
			writeJSON(w, map[string]any{"available": false, "version": appVersion})
			return true
		}
		writeJSON(w, map[string]any{"available": update.Version != "", "version": update.Version, "notes": update.Notes})
		return true
	case r.URL.Path == "/api/update/apply" && r.Method == http.MethodPost:
		update, err := fetchUpdate()
		if err == nil {
			err = launchUpdate(update)
		}
		if err != nil {
			fail(w, err)
			return true
		}
		writeJSON(w, map[string]bool{"ok": true})
		go func() { time.Sleep(700 * time.Millisecond); os.Exit(0) }()
		return true
	}
	return false
}

func launchUpdate(update updateManifest) error {
	target, err := os.Executable()
	if err != nil {
		return err
	}
	target, err = filepath.Abs(target)
	if err != nil {
		return err
	}
	dir, err := os.MkdirTemp("", "MundoNexus-update-")
	if err != nil {
		return err
	}
	staged := filepath.Join(dir, "Mundo-Nexus-new.exe")
	response, err := (&http.Client{Timeout: 4 * time.Minute}).Get(update.URL)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("download de atualização respondeu %d", response.StatusCode)
	}
	file, err := os.OpenFile(staged, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0700)
	if err != nil {
		return err
	}
	hash := sha256.New()
	_, copyErr := io.Copy(io.MultiWriter(file, hash), io.LimitReader(response.Body, 250<<20))
	closeErr := file.Close()
	if copyErr != nil {
		return copyErr
	}
	if closeErr != nil {
		return closeErr
	}
	if !strings.EqualFold(hex.EncodeToString(hash.Sum(nil)), update.SHA256) {
		return errors.New("o hash SHA-256 da atualização não confere")
	}
	helper := filepath.Join(dir, "Mundo-Nexus-atualizador.exe")
	self, err := os.Open(target)
	if err != nil {
		return err
	}
	helperFile, err := os.OpenFile(helper, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0700)
	if err == nil {
		_, err = io.Copy(helperFile, self)
		helperFile.Close()
	}
	self.Close()
	if err != nil {
		return err
	}
	return exec.Command(helper, "--mnr-update", "--pid", strconv.Itoa(os.Getpid()), "--source", staged, "--target", target).Start()
}

func runUpdateHelper() bool {
	if len(os.Args) < 2 || os.Args[1] != "--mnr-update" {
		return false
	}
	var pid int
	var source, target string
	for i := 2; i+1 < len(os.Args); i += 2 {
		switch os.Args[i] {
		case "--pid":
			pid, _ = strconv.Atoi(os.Args[i+1])
		case "--source":
			source = os.Args[i+1]
		case "--target":
			target = os.Args[i+1]
		}
	}
	if pid <= 0 || source == "" || target == "" {
		return true
	}
	// The installed binary may still be locked while the previous process exits.
	if info, err := os.Stat(source); err != nil || !info.Mode().IsRegular() {
		return true
	}
	for i := 0; i < 120; i++ {
		if err := replaceExecutableWithRollback(source, target); err == nil {
			_ = exec.Command(target).Start()
			return true
		}
		time.Sleep(500 * time.Millisecond)
	}
	return true
}

// replaceExecutableWithRollback preserves a recoverable copy of the old binary.
// If the new binary cannot be moved into place, restore the previous version.
func replaceExecutableWithRollback(source, target string) error {
	backup := target + ".previous-update-" + randomHex(6)
	if err := os.Rename(target, backup); err != nil {
		return err
	}
	if err := os.Rename(source, target); err != nil {
		return errors.Join(err, os.Rename(backup, target))
	}
	return nil
}
