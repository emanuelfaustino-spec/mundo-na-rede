//go:build !windows

package main

import "os/exec"

func hideServerWindow(_ *exec.Cmd) {}
