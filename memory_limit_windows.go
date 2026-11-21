//go:build windows

package main

import (
	"errors"
	"fmt"
	"syscall"
	"unsafe"
)

type jobBasicLimits struct {
	ProcessUserTime    int64
	JobUserTime        int64
	Flags              uint32
	MinimumWorkingSet  uintptr
	MaximumWorkingSet  uintptr
	ActiveProcessLimit uint32
	Affinity           uintptr
	PriorityClass      uint32
	SchedulingClass    uint32
}
type jobCounters struct {
	ReadOperationCount  uint64
	WriteOperationCount uint64
	OtherOperationCount uint64
	ReadTransferCount   uint64
	WriteTransferCount  uint64
	OtherTransferCount  uint64
}
type jobExtendedLimits struct {
	Basic         jobBasicLimits
	IO            jobCounters
	ProcessMemory uintptr
	JobMemory     uintptr
	PeakProcess   uintptr
	PeakJob       uintptr
}

var kernelJob = syscall.NewLazyDLL("kernel32.dll")
var createJobObject = kernelJob.NewProc("CreateJobObjectW")
var setJobInfo = kernelJob.NewProc("SetInformationJobObject")
var assignJob = kernelJob.NewProc("AssignProcessToJobObject")

func limitServerProcess(pid int, limitMiB int) (func(), error) {
	if pid <= 0 || limitMiB < 512 {
		return nil, fmt.Errorf("limite de memória inválido")
	}
	handle, _, callErr := createJobObject.Call(0, 0)
	if handle == 0 {
		return nil, fmt.Errorf("criação do controle de RAM: %v", callErr)
	}
	job := syscall.Handle(handle)
	closeJob := func() { _ = syscall.CloseHandle(job) }
	limits := jobExtendedLimits{}
	limits.Basic.Flags = 0x200 // JOB_OBJECT_LIMIT_JOB_MEMORY
	limits.JobMemory = uintptr(limitMiB) * 1024 * 1024
	ok, _, setErr := setJobInfo.Call(handle, 9, uintptr(unsafe.Pointer(&limits)), uintptr(unsafe.Sizeof(limits)))
	if ok == 0 {
		closeJob()
		return nil, fmt.Errorf("limite de RAM do Windows: %v", setErr)
	}
	process, err := syscall.OpenProcess(0x0100|0x0001|0x0400, false, uint32(pid)) // SET_QUOTA + QUERY_INFORMATION
	if err != nil {
		closeJob()
		return nil, err
	}
	defer syscall.CloseHandle(process)
	ok, _, assignErr := assignJob.Call(handle, uintptr(process))
	if ok == 0 {
		closeJob()
		return nil, errors.New("o Windows não permitiu isolar a memória do servidor: " + assignErr.Error())
	}
	return closeJob, nil
}
