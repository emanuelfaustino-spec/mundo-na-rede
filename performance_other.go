//go:build !windows && !linux && !android

package main

func readCPU(_ int) cpuSample { return cpuSample{} }
