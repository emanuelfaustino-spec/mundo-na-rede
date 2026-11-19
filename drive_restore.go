package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var snapshotName = regexp.MustCompile(`^MundoNaRede-(bedrock|pnx|java|both)-([a-f0-9]{16})-[0-9]{8}-[0-9]{6}-[a-f0-9]{6}\.zip$`)
var driveFileID = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,160}$`)
var rcloneRemote = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,48}$`)

type cloudCopy struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Engine   string `json:"engine"`
	Modified string `json:"modified"`
	Size     int64  `json:"size"`
}

func (d *driveBackup) cloudReady() bool {
	return d.account.Token != nil || d.remote != ""
}
func (d *driveBackup) runRclone(ctx context.Context, args ...string) *exec.Cmd {
	return exec.CommandContext(ctx, "rclone", args...)
}
func (d *driveBackup) remotePath(name string) string {
	return d.remote + ":MundoNaRede/" + name
}
func (d *driveBackup) checkRemote(ctx context.Context) error {
	marker, err := os.ReadFile(filepath.Join(d.root, "rclone-drive-marker"))
	if err != nil {
		return errors.New("Google Drive requires a paired backup folder")
	}
	name := strings.TrimSpace(string(marker))
	if !regexp.MustCompile(`^device-[a-f0-9]{32}\.txt$`).MatchString(name) {
		return errors.New("Invalid Drive pairing")
	}
	if _, err = d.runRclone(ctx, "cat", d.remotePath(name), "--count", "1").Output(); err != nil {
		return errors.New("Google Drive backup folder could not be verified")
	}
	return nil
}
func (d *driveBackup) pairRemote(ctx context.Context) error {
	path := filepath.Join(d.root, "rclone-drive-marker")
	if _, err := os.Stat(path); err == nil {
		return d.checkRemote(ctx)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := d.runRclone(ctx, "mkdir", d.remotePath("")).Run(); err != nil {
		return errors.New("Could not open the backup folder")
	}
	name := "device-" + randomHex(16) + ".txt"
	cmd := d.runRclone(ctx, "rcat", d.remotePath(name))
	cmd.Stdin = strings.NewReader("Mundo na Rede")
	if err := cmd.Run(); err != nil {
		return errors.New("Could not pair Google Drive through rclone")
	}
	return os.WriteFile(path, []byte(name), 0600)
}
func (d *driveBackup) listCopies(ctx context.Context) ([]cloudCopy, error) {
	d.mu.Lock()
	token := d.account.Token
	remote := d.remote
	d.mu.Unlock()
	out := []cloudCopy{}
	if remote != "" {
		if err := d.checkRemote(ctx); err != nil {
			return nil, err
		}
		data, err := d.runRclone(ctx, "lsjson", d.remotePath(""), "--files-only", "--include", "MundoNaRede-*.zip").Output()
		if err != nil {
			return nil, errors.New("Could not list backups through rclone")
		}
		var files []struct {
			Name, ModTime string
			Size          int64
		}
		if err = json.Unmarshal(data, &files); err != nil {
			return nil, err
		}
		for _, file := range files {
			parts := snapshotName.FindStringSubmatch(file.Name)
			if parts != nil {
				out = append(out, cloudCopy{file.Name, file.Name, parts[1], file.ModTime, file.Size})
			}
		}
		return out, nil
	}
	if token == nil {
		return nil, errors.New("Connect Google Drive first")
	}
	client := d.config.Client(d.ctx(ctx), token)
	page := ""
	for range 20 {
		query := url.Values{"q": {"name contains 'MundoNaRede-' and trashed = false and 'me' in owners"}, "fields": {"nextPageToken,files(id,name,modifiedTime,size)"}, "orderBy": {"modifiedTime desc"}, "pageSize": {"100"}}
		if page != "" {
			query.Set("pageToken", page)
		}
		res, err := downloadRequest(client, ctx, googleDrive+"/drive/v3/files?"+query.Encode())
		if err != nil {
			return nil, err
		}
		var data struct {
			NextPageToken string
			Files         []struct{ ID, Name, ModifiedTime, Size string }
		}
		err = json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(&data)
		res.Body.Close()
		if err != nil {
			return nil, err
		}
		for _, file := range data.Files {
			parts := snapshotName.FindStringSubmatch(file.Name)
			size, _ := strconv.ParseInt(file.Size, 10, 64)
			if parts != nil {
				out = append(out, cloudCopy{file.ID, file.Name, parts[1], file.ModifiedTime, size})
			}
		}
		page = data.NextPageToken
		if page == "" {
			break
		}
	}
	return out, nil
}

func (d *driveBackup) downloadCopy(ctx context.Context, id, engine string, dest *os.File) error {
	copies, err := d.listCopies(ctx)
	if err != nil {
		return err
	}
	var selected *cloudCopy
	for i := range copies {
		if copies[i].ID == id {
			selected = &copies[i]
			break
		}
	}
	if selected == nil || selected.Engine != engine || selected.Size <= 0 || selected.Size > maxArchive {
		return errors.New("Backup unavailable or incompatible with this server")
	}
	if d.remote != "" {
		if err = d.checkRemote(ctx); err != nil {
			return err
		}
		cmd := d.runRclone(ctx, "cat", d.remotePath(selected.Name), "--count", strconv.FormatInt(maxArchive+1, 10))
		cmd.Stdout = dest
		if err = cmd.Run(); err != nil {
			return errors.New("Could not download backup through rclone")
		}
	} else {
		if !driveFileID.MatchString(id) {
			return errors.New("Invalid Drive file")
		}
		d.mu.Lock()
		token := d.account.Token
		d.mu.Unlock()
		if token == nil {
			return errors.New("Connect Google Drive first")
		}
		client := d.config.Client(d.ctx(ctx), token)
		res, err := downloadRequest(client, ctx, googleDrive+"/drive/v3/files/"+id+"?alt=media")
		if err != nil {
			return err
		}
		defer res.Body.Close()
		if _, err = io.Copy(dest, io.LimitReader(res.Body, maxArchive+1)); err != nil {
			return err
		}
	}
	info, err := dest.Stat()
	if err != nil {
		return err
	}
	if info.Size() != selected.Size {
		return fmt.Errorf("backup download size mismatch")
	}
	return nil
}

func (d *driveBackup) deleteCopy(ctx context.Context, id string) error {
	if !driveFileID.MatchString(id) {
		return errors.New("backup inválido")
	}
	d.mu.Lock()
	token, remote := d.account.Token, d.remote
	d.mu.Unlock()
	if remote != "" {
		return errors.New("exclusão de backup requer o Google Drive conectado diretamente")
	}
	if token == nil {
		return errors.New("conecte o Google Drive primeiro")
	}
	client := d.config.Client(d.ctx(ctx), token)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, googleDrive+"/drive/v3/files/"+id, nil)
	if err != nil {
		return err
	}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNoContent {
		return fmt.Errorf("não foi possível excluir o backup (%d)", res.StatusCode)
	}
	return nil
}

func (d *driveBackup) retainCloudCopies(ctx context.Context, profileID string) error {
	copies, err := d.listCopies(ctx)
	if err != nil {
		return err
	}
	filtered := make([]cloudCopy, 0, len(copies))
	needle := "-" + profileID + "-"
	for _, copy := range copies {
		if strings.Contains(copy.Name, needle) {
			filtered = append(filtered, copy)
		}
	}
	sort.Slice(filtered, func(i, j int) bool { return filtered[i].Modified > filtered[j].Modified })
	for _, copy := range filtered[backupRetention:] {
		if err := d.deleteCopy(ctx, copy.ID); err != nil {
			return err
		}
	}
	return nil
}

func (m *manager) restoreCloud(r *http.Request, p Profile, game string) error {
	var input struct{ ID string }
	if err := json.NewDecoder(io.LimitReader(r.Body, 4096)).Decode(&input); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(r.Context(), 6*time.Minute)
	defer cancel()
	temp, err := os.MkdirTemp(m.dir(p.ID), "cloud-restore-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(temp)
	archive, err := os.Create(filepath.Join(temp, "backup.zip"))
	if err != nil {
		return err
	}
	if err = errors.Join(m.drive.downloadCopy(ctx, input.ID, p.Engine, archive), archive.Close()); err != nil {
		return err
	}
	extracted := filepath.Join(temp, "game")
	if err = unpack(archive.Name(), extracted); err != nil {
		return err
	}
	return restoreGame(p, extracted, game)
}
