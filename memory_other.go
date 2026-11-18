//go:build !windows && !linux && !android

package main

func memoryMiB() (int, int) { return 0, 0 }
