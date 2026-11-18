package main

import (
	"syscall"
	"unsafe"
)

func memoryMiB() (int, int) {
	status := struct {
		Length                   uint32
		Load                     uint32
		TotalPhysical            uint64
		AvailablePhysical        uint64
		TotalPageFile            uint64
		AvailablePageFile        uint64
		TotalVirtual             uint64
		AvailableVirtual         uint64
		AvailableExtendedVirtual uint64
	}{}
	status.Length = uint32(unsafe.Sizeof(status))
	call := syscall.NewLazyDLL("kernel32.dll").NewProc("GlobalMemoryStatusEx")
	result, _, _ := call.Call(uintptr(unsafe.Pointer(&status)))
	if result == 0 {
		return 0, 0
	}
	return int(status.TotalPhysical >> 20), int(status.AvailablePhysical >> 20)
}
