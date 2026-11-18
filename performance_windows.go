package main

import (
	"syscall"
	"unsafe"
)

var processTimes = syscall.NewLazyDLL("kernel32.dll").NewProc("GetProcessTimes")
var systemTimes = syscall.NewLazyDLL("kernel32.dll").NewProc("GetSystemTimes")
var processMemory = syscall.NewLazyDLL("psapi.dll").NewProc("GetProcessMemoryInfo")

func readCPU(pid int) cpuSample {
	var sample cpuSample
	var idle, kernel, user uint64
	if success, _, _ := systemTimes.Call(uintptr(unsafe.Pointer(&idle)), uintptr(unsafe.Pointer(&kernel)), uintptr(unsafe.Pointer(&user))); success != 0 {
		sample.total = kernel + user
		sample.idle = idle
	}
	handle, err := syscall.OpenProcess(0x0400|0x0010, false, uint32(pid))
	if err != nil {
		return sample
	}
	defer syscall.CloseHandle(handle)
	var created, exited uint64
	kernel, user = 0, 0
	if success, _, _ := processTimes.Call(uintptr(handle), uintptr(unsafe.Pointer(&created)), uintptr(unsafe.Pointer(&exited)), uintptr(unsafe.Pointer(&kernel)), uintptr(unsafe.Pointer(&user))); success != 0 {
		sample.process = kernel + user
	}
	var stats struct {
		Length, PageFaultCount                                                               uint32
		PeakWorkingSetSize, WorkingSetSize, QuotaPeakPagedPoolUsage, QuotaPagedPoolUsage     uintptr
		QuotaPeakNonPagedPoolUsage, QuotaNonPagedPoolUsage, PagefileUsage, PeakPagefileUsage uintptr
	}
	stats.Length = uint32(unsafe.Sizeof(stats))
	if success, _, _ := processMemory.Call(uintptr(handle), uintptr(unsafe.Pointer(&stats)), uintptr(stats.Length)); success != 0 {
		sample.memory = uint64(stats.WorkingSetSize)
	}
	return sample
}
