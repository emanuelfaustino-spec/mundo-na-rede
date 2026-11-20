//go:build !windows

package main

// Non-Windows systems apply the conservative budget before starting the engine.
// Hard isolation depends on platform-specific sandbox or cgroup integration.
func limitServerProcess(pid int, limitMiB int) (func(), error) { return func() {}, nil }
