//go:build windows

package main

import (
	"os/exec"
	"syscall"
)

// Server engines are console programs. The panel already captures their output
// and stdin, so exposing a second console only creates a fragile stop button.
func hideServerWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
}
