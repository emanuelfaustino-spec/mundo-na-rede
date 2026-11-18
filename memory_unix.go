//go:build linux || android

package main

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

func memoryMiB() (int, int) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, 0
	}
	defer file.Close()
	var total, available int
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 {
			continue
		}
		switch fields[0] {
		case "MemTotal:":
			total, _ = strconv.Atoi(fields[1])
		case "MemAvailable:":
			available, _ = strconv.Atoi(fields[1])
		}
	}
	return total / 1024, available / 1024
}
