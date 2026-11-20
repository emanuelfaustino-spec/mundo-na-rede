//go:build !windows

package main

import (
	"context"
	"errors"
	"os/exec"
)

func bundledJava(root string) string { return "" }
func javaExecutable(root string) (string, error) {
	if java, err := exec.LookPath("java"); err == nil {
		return java, nil
	}
	return "", errors.New("Java não está instalado neste computador")
}
func (m *manager) installPortableJava(ctx context.Context) error {
	return errors.New("o pacote Java portátil está disponível somente para Windows")
}
