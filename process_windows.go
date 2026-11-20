//go:build windows

package main

import (
	"os/exec"
	"syscall"
)

// Server engines are console programs. The panel already captures their output
// and stdin, so exposing a second console only creates a fragile stop button.
func hideServerWindow(cmd *exec.Cmd) {
	// HideWindow only hides a console after Windows creates it. CREATE_NO_WINDOW
	// prevents that console from being allocated at all, so closing a visible
	// terminal can never accidentally stop the hosted world.
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000} // CREATE_NO_WINDOW
}
