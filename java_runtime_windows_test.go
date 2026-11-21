//go:build windows

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestBundledJavaLookup(t *testing.T) {
	root := t.TempDir()
	if bundledJava(root) != "" {
		t.Fatal("biblioteca vazia não pode conter Java")
	}
	bin := filepath.Join(root, "engines", "java", "25", "bin")
	if err := os.MkdirAll(bin, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bin, "java.exe"), []byte("fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	if got := bundledJava(root); got != filepath.Join(bin, "java.exe") {
		t.Fatalf("Java não encontrado: %s", got)
	}
}
func TestWindowsJobAssignment(t *testing.T) {
	// A harmless short-lived child verifies that Windows accepts the job object.
	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", "Start-Sleep -Seconds 4")
	if err := cmd.Start(); err != nil {
		t.Skipf("processo auxiliar não disponível: %v", err)
	}
	defer func() { _ = cmd.Process.Kill(); _ = cmd.Wait() }()
	release, err := limitServerProcess(cmd.Process.Pid, 1024)
	if err != nil {
		t.Fatalf("não foi possível aplicar o limite ao processo de teste: %v", err)
	}
	release()
}
