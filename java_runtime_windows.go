//go:build windows

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
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const adoptiumJRE25 = "https://api.adoptium.net/v3/assets/latest/25/hotspot?architecture=x64&heap_size=normal&image_type=jre&jvm_impl=hotspot&os=windows&vendor=eclipse"

func bundledJava(root string) string {
	path := filepath.Join(root, "engines", "java", "25", "bin", "java.exe")
	if info, err := os.Stat(path); err == nil && info.Mode().IsRegular() {
		return path
	}
	return ""
}

func javaExecutable(root string) (string, error) {
	if java := bundledJava(root); java != "" {
		return java, nil
	}
	if java, err := exec.LookPath("java"); err == nil {
		return java, nil
	}
	return "", errors.New("Java não está disponível. Prepare o Java 25 na Biblioteca de motores.")
}

func (m *manager) installPortableJava(ctx context.Context) error {
	if bundledJava(m.root) != "" {
		return nil
	}
	base := filepath.Join(m.root, "engines", "java")
	if err := os.MkdirAll(base, 0700); err != nil {
		return err
	}
	destination := filepath.Join(base, "25")
	if _, err := os.Stat(destination); err == nil {
		return errors.New("a instalação portátil do Java está incompleta; revise a biblioteca de motores")
	}
	client := &http.Client{Timeout: 8 * time.Minute, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if req.URL.Scheme != "https" {
			return errors.New("redirecionamento não seguro do Java")
		}
		if len(via) >= 12 {
			return errors.New("redirecionamentos demais")
		}
		return nil
	}}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, adoptiumJRE25, nil)
	if err != nil {
		return err
	}
	request.Header.Set("User-Agent", paperAgent)
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return fmt.Errorf("serviço Adoptium indisponível (%d)", response.StatusCode)
	}
	var releases []struct {
		Binary struct {
			Package struct {
				Link     string `json:"link"`
				Checksum string `json:"checksum"`
			} `json:"package"`
		} `json:"binary"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 2<<20)).Decode(&releases); err != nil {
		return err
	}
	if len(releases) == 0 {
		return errors.New("Adoptium não retornou Java 25 para Windows")
	}
	pkg := releases[0].Binary.Package
	link, err := url.Parse(pkg.Link)
	if err != nil || link.Scheme != "https" || !strings.EqualFold(link.Host, "github.com") ||
		!strings.HasPrefix(strings.ToLower(link.Path), "/adoptium/temurin25-binaries/releases/download/") || len(pkg.Checksum) != 64 {
		return errors.New("metadados da distribuição Java inesperados")
	}
	stage, err := os.MkdirTemp(base, "prepare-java-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(stage)
	archive := filepath.Join(stage, "jre.zip")
	request, err = http.NewRequestWithContext(ctx, http.MethodGet, pkg.Link, nil)
	if err != nil {
		return err
	}
	request.Header.Set("User-Agent", paperAgent)
	response, err = client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return fmt.Errorf("download Java indisponível (%d)", response.StatusCode)
	}
	target, err := os.Create(archive)
	if err != nil {
		return err
	}
	hash := sha256.New()
	bytes, copyErr := io.Copy(io.MultiWriter(target, hash), io.LimitReader(response.Body, (256<<20)+1))
	copyErr = errors.Join(copyErr, target.Close())
	if copyErr != nil {
		return copyErr
	}
	if bytes > 256<<20 {
		return errors.New("arquivo Java maior do que o permitido")
	}
	if !strings.EqualFold(hex.EncodeToString(hash.Sum(nil)), pkg.Checksum) {
		return errors.New("checksum SHA-256 do Java não confere")
	}
	extracted := filepath.Join(stage, "unpacked")
	if err := os.Mkdir(extracted, 0700); err != nil {
		return err
	}
	if err := unpack(archive, extracted); err != nil {
		return err
	}
	candidate := ""
	if _, err := os.Stat(filepath.Join(extracted, "bin", "java.exe")); err == nil {
		candidate = extracted
	}
	if candidate == "" {
		entries, err := os.ReadDir(extracted)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			possible := filepath.Join(extracted, entry.Name())
			if info, err := os.Stat(filepath.Join(possible, "bin", "java.exe")); err == nil && info.Mode().IsRegular() {
				candidate = possible
				break
			}
		}
	}
	if candidate == "" {
		return errors.New("o pacote Java não contém um executável válido")
	}
	if err := os.Rename(candidate, destination); err != nil {
		return err
	}
	return nil
}
