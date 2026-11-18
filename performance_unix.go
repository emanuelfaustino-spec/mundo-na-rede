//go:build linux || android

package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func readCPU(pid int) cpuSample {
	var sample cpuSample
	source, _ := os.ReadFile("/proc/stat")
	lines := strings.SplitN(string(source), "\n", 2)
	fields := strings.Fields(lines[0])
	if len(fields) >= 9 {
		for index, field := range fields[1:9] {
			value, _ := strconv.ParseUint(field, 10, 64)
			sample.total += value
			if index == 3 || index == 4 {
				sample.idle += value
			}
		}
	}
	data, _ := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	text := string(data)
	if end := strings.LastIndex(text, ")"); end >= 0 {
		fields = strings.Fields(text[end+1:])
		if len(fields) >= 22 {
			user, _ := strconv.ParseUint(fields[11], 10, 64)
			sys, _ := strconv.ParseUint(fields[12], 10, 64)
			pages, _ := strconv.ParseUint(fields[21], 10, 64)
			sample.process = user + sys
			sample.memory = pages * uint64(os.Getpagesize())
		}
	}
	return sample
}
