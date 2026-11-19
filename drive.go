package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"golang.org/x/oauth2"
)

const googleDrive = "https://www.googleapis.com"

type driveAccount struct {
	Name  string        `json:"name"`
	Email string        `json:"email"`
	Token *oauth2.Token `json:"token"`
}
type driveBackup struct {
	mu           sync.Mutex
	root         string
	redirect     string
	remote       string
	config       oauth2.Config
	account      driveAccount
	pendingState string
	verifier     string
	expires      time.Time
	status       string
	queued       chan struct{}
	client       *http.Client
}

func newDrive(root, redirect string) *driveBackup {
	d := &driveBackup{root: root, redirect: redirect, queued: make(chan struct{}, 1), client: &http.Client{Timeout: 3 * time.Minute, CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }}}
	if remote := os.Getenv("MUNDO_DRIVE_REMOTE"); rcloneRemote.MatchString(remote) {
		d.remote = remote
	}
	if runtime.GOOS != "android" && d.remote == "" {
		_ = d.configureClient(os.Getenv("MUNDO_GOOGLE_CLIENT_ID"), os.Getenv("MUNDO_GOOGLE_CLIENT_SECRET"), false)
		if d.config.ClientID == "" {
			var saved struct {
				ClientID     string `json:"clientID"`
				ClientSecret string `json:"clientSecret"`
			}
			if data, err := os.ReadFile(d.oauthPath()); err == nil && json.Unmarshal(data, &saved) == nil {
				_ = d.configureClient(saved.ClientID, saved.ClientSecret, false)
			}
		}
	}
	if _, err := os.Stat(filepath.Join(root, "rclone-drive-marker")); err == nil && d.remote != "" {
		d.account.Email = "rclone:" + d.remote
	}
	if data, err := os.ReadFile(d.accountPath()); err == nil && d.config.ClientID != "" {
		json.Unmarshal(data, &d.account)
	}
	return d
}
func (d *driveBackup) accountPath() string { return filepath.Join(d.root, "google-account.json") }
func (d *driveBackup) oauthPath() string   { return filepath.Join(d.root, "google-oauth.json") }
func (d *driveBackup) configureClient(clientID, clientSecret string, persist bool) error {
	clientID, clientSecret = strings.TrimSpace(clientID), strings.TrimSpace(clientSecret)
	if clientID == "" || clientSecret == "" || !strings.HasSuffix(clientID, ".apps.googleusercontent.com") {
		return errors.New("use o ID e o segredo de um cliente OAuth do tipo Aplicativo para computador")
	}
	d.config = oauth2.Config{ClientID: clientID, ClientSecret: clientSecret, RedirectURL: d.redirect,
		Scopes:   []string{"openid", "profile", "email", "https://www.googleapis.com/auth/drive.file"},
		Endpoint: oauth2.Endpoint{AuthURL: "https://accounts.google.com/o/oauth2/v2/auth", TokenURL: "https://oauth2.googleapis.com/token"}}
	if !persist {
		return nil
	}
	data, _ := json.Marshal(map[string]string{"clientID": clientID, "clientSecret": clientSecret})
	return os.WriteFile(d.oauthPath(), data, 0600)
}
func (d *driveBackup) ctx(parent context.Context) context.Context {
	return context.WithValue(parent, oauth2.HTTPClient, d.client)
}
func (d *driveBackup) handle(w http.ResponseWriter, r *http.Request) {
	d.mu.Lock()
	defer d.mu.Unlock()
	switch {
	case r.URL.Path == "/api/account/config" && r.Method == "POST":
		if d.remote != "" || runtime.GOOS == "android" {
			fail(w, errors.New("a configuração OAuth é exclusiva do aplicativo desktop"))
			return
		}
		var input struct {
			ClientID     string `json:"clientID"`
			ClientSecret string `json:"clientSecret"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8192)).Decode(&input); err != nil {
			fail(w, err)
			return
		}
		if err := d.configureClient(input.ClientID, input.ClientSecret, true); err != nil {
			fail(w, err)
			return
		}
		writeJSON(w, map[string]bool{"ok": true})
	case r.URL.Path == "/api/account/pair" && r.Method == "POST":
		if d.remote == "" {
			fail(w, errors.New("rclone não está configurado"))
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()
		if err := d.pairRemote(ctx); err != nil {
			fail(w, err)
			return
		}
		d.account.Email = "rclone:" + d.remote
		select {
		case d.queued <- struct{}{}:
		default:
		}
		writeJSON(w, map[string]bool{"ok": true})
	case r.URL.Path == "/api/account" && r.Method == "GET":
		writeJSON(w, map[string]any{"available": d.config.ClientID != "" || d.remote != "", "needsSetup": d.config.ClientID == "" && d.remote == "" && runtime.GOOS != "android", "bridge": d.remote != "", "name": d.account.Name, "email": d.account.Email, "status": d.status, "connected": d.cloudReady() && (d.remote == "" || d.account.Email != "")})
	case r.URL.Path == "/api/account/connect" && r.Method == "POST":
		if d.config.ClientID == "" {
			fail(w, errors.New("Google Drive indisponível nesta versão do aplicativo"))
			return
		}
		if d.account.Token != nil {
			fail(w, errors.New("desconecte a conta atual primeiro"))
			return
		}
		d.pendingState = randomHex(24)
		d.verifier = oauth2.GenerateVerifier()
		d.expires = time.Now().Add(10 * time.Minute)
		writeJSON(w, map[string]string{"url": d.config.AuthCodeURL(d.pendingState, oauth2.AccessTypeOffline, oauth2.S256ChallengeOption(d.verifier), oauth2.SetAuthURLParam("prompt", "consent select_account"))})
	case r.URL.Path == "/api/account/disconnect" && r.Method == "POST":
		if d.remote != "" {
			if err := os.Remove(filepath.Join(d.root, "rclone-drive-marker")); err != nil && !errors.Is(err, os.ErrNotExist) {
				fail(w, err)
				return
			}
		}
		if err := os.Remove(d.accountPath()); err != nil && !errors.Is(err, os.ErrNotExist) {
			fail(w, err)
			return
		}
		d.account = driveAccount{}
		d.pendingState = ""
		d.status = ""
		writeJSON(w, map[string]bool{"ok": true})
	default:
		http.NotFound(w, r)
	}
}
func (d *driveBackup) callback(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(405)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.pendingState == "" || r.URL.Query().Get("state") != d.pendingState || time.Now().After(d.expires) {
		http.Error(w, "Invalid or expired session / Sessão inválida ou expirada", 400)
		return
	}
	d.pendingState = ""
	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()
	token, err := d.config.Exchange(d.ctx(ctx), r.URL.Query().Get("code"), oauth2.VerifierOption(d.verifier))
	d.verifier = ""
	if err != nil || token.RefreshToken == "" {
		http.Error(w, "Google authorization failed / Falha ao autorizar o Google", 400)
		return
	}
	client := d.config.Client(d.ctx(ctx), token)
	profile, err := client.Get("https://openidconnect.googleapis.com/v1/userinfo")
	if err != nil {
		http.Error(w, "Account unavailable / Conta indisponível", 502)
		return
	}
	defer profile.Body.Close()
	account := driveAccount{Token: token}
	if profile.StatusCode != 200 || json.NewDecoder(io.LimitReader(profile.Body, 8192)).Decode(&account) != nil || account.Email == "" {
		http.Error(w, "Account unavailable / Conta indisponível", 502)
		return
	}
	data, err := json.Marshal(account)
	if err != nil {
		http.Error(w, "Error saving account / Erro ao salvar a conta", 500)
		return
	}
	if err = os.WriteFile(d.accountPath(), data, 0600); err != nil {
		http.Error(w, "Error saving account / Erro ao salvar a conta", 500)
		return
	}
	d.account = account
	fmt.Fprintln(w, "Google Drive connected. Return to the app. / Google Drive conectado. Volte ao aplicativo.")
}

func (d *driveBackup) capture(p Profile, game string) {
	d.mu.Lock()
	if !d.cloudReady() || d.account.Email == "" {
		d.mu.Unlock()
		return
	}
	queue := filepath.Join(d.root, "drive-queue", d.accountKey())
	d.mu.Unlock()
	if err := d.snapshot(queue, p, game); err != nil {
		d.mu.Lock()
		d.status = "snapshot_error"
		d.mu.Unlock()
		return
	}
	d.mu.Lock()
	d.status = "pending"
	d.mu.Unlock()
	select {
	case d.queued <- struct{}{}:
	default:
	}
}
func (d *driveBackup) accountKey() string {
	// An account marker keeps offline snapshots from being uploaded to a different Google account.
	identity := d.account.Email
	if d.remote != "" {
		marker, _ := os.ReadFile(filepath.Join(d.root, "rclone-drive-marker"))
		identity += ":" + string(marker)
	}
	return fmt.Sprintf("%x", sha256.Sum256([]byte(strings.ToLower(identity))))
}
func (d *driveBackup) snapshot(queue string, p Profile, game string) error {
	if err := os.MkdirAll(queue, 0700); err != nil {
		return err
	}
	entries, err := os.ReadDir(queue)
	if err != nil {
		return err
	}
	var total int64
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".zip") {
			continue
		}
		info, statErr := entry.Info()
		if statErr == nil {
			total += info.Size()
		}
	}
	if total >= 4<<30 {
		return errors.New("a fila de backups locais está cheia")
	}
	tmp, err := os.CreateTemp(queue, "snapshot-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if err = errors.Join(zipGame(tmp, game), tmp.Close()); err != nil {
		return err
	}
	name := "MundoNaRede-" + p.Engine + "-" + p.ID + "-" + time.Now().UTC().Format("20060102-150405") + "-" + randomHex(3) + ".zip"
	return os.Rename(tmp.Name(), filepath.Join(queue, name))
}
func (d *driveBackup) loop(ctx context.Context) {
	tick := time.NewTicker(time.Minute)
	defer tick.Stop()
	for {
		d.flush(ctx)
		select {
		case <-ctx.Done():
			return
		case <-d.queued:
		case <-tick.C:
		}
	}
}
func (d *driveBackup) flush(ctx context.Context) {
	d.mu.Lock()
	if !d.cloudReady() || d.account.Email == "" {
		d.mu.Unlock()
		return
	}
	token, email := d.account.Token, d.account.Email
	queue := filepath.Join(d.root, "drive-queue", d.accountKey())
	d.mu.Unlock()
	files, _ := os.ReadDir(queue)
	var client *http.Client
	if token != nil {
		client = d.config.Client(d.ctx(ctx), token)
	}
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".zip") {
			continue
		}
		d.mu.Lock()
		same := d.account.Token == token && d.account.Email == email && queue == filepath.Join(d.root, "drive-queue", d.accountKey())
		d.mu.Unlock()
		if !same || ctx.Err() != nil {
			return
		}
		var uploadErr error
		if d.remote != "" {
			transfer, cancel := context.WithTimeout(ctx, 6*time.Minute)
			uploadErr = d.checkRemote(transfer)
			if uploadErr == nil {
				uploadErr = d.runRclone(transfer, "copyto", filepath.Join(queue, file.Name()), d.remotePath(file.Name()), "--immutable").Run()
			}
			cancel()
		} else {
			uploadErr = sendDriveSnapshot(ctx, client, filepath.Join(queue, file.Name()), file.Name())
		}
		d.mu.Lock()
		if d.account.Token == token && d.account.Email == email {
			d.status = "uploaded"
			if uploadErr != nil {
				d.status = "upload_error"
			}
		}
		d.mu.Unlock()
		if uploadErr != nil {
			return
		}
		if d.remote == "" {
			if parts := snapshotName.FindStringSubmatch(file.Name()); parts != nil {
				if err := d.retainCloudCopies(ctx, parts[2]); err != nil {
					d.mu.Lock()
					d.status = "retention_error"
					d.mu.Unlock()
					return
				}
			}
		}
		os.Remove(filepath.Join(queue, file.Name()))
	}
}
func sendDriveSnapshot(ctx context.Context, client *http.Client, path, name string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return err
	}
	// Look up the unique name first so retries after a lost upload response do not duplicate backups.
	check, err := http.NewRequestWithContext(ctx, "GET", googleDrive+"/drive/v3/files", nil)
	if err != nil {
		return err
	}
	query := check.URL.Query()
	query.Set("q", fmt.Sprintf("name = '%s' and trashed = false", name))
	query.Set("fields", "files(id)")
	check.URL.RawQuery = query.Encode()
	found, err := client.Do(check)
	if err != nil {
		return err
	}
	var listing struct{ Files []struct{ ID string } }
	err = json.NewDecoder(io.LimitReader(found.Body, 8192)).Decode(&listing)
	found.Body.Close()
	if err != nil || found.StatusCode != 200 {
		return errors.New("não foi possível localizar o backup no Drive")
	}
	if len(listing.Files) > 0 {
		return nil
	}
	metadata, _ := json.Marshal(map[string]string{"name": name, "mimeType": "application/zip"})
	request, err := http.NewRequestWithContext(ctx, "POST", googleDrive+"/upload/drive/v3/files?uploadType=resumable", bytes.NewReader(metadata))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Upload-Content-Type", "application/zip")
	request.Header.Set("X-Upload-Content-Length", fmt.Sprint(info.Size()))
	session, err := client.Do(request)
	if err != nil {
		return err
	}
	session.Body.Close()
	location, err := session.Location()
	if err != nil || session.StatusCode != 200 || location.Scheme != "https" || location.Host != "www.googleapis.com" {
		return errors.New("não foi possível iniciar o envio para o Drive")
	}
	upload, err := http.NewRequestWithContext(ctx, "PUT", location.String(), file)
	if err != nil {
		return err
	}
	upload.ContentLength = info.Size()
	upload.Header.Set("Content-Type", "application/zip")
	response, err := client.Do(upload)
	if err != nil {
		return err
	}
	response.Body.Close()
	if response.StatusCode != 200 && response.StatusCode != 201 {
		return fmt.Errorf("falha ao enviar para o Drive: %d", response.StatusCode)
	}
	return nil
}
