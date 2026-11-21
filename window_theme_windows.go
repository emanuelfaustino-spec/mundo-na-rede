//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

const (
	dwmDarkMode     = 20
	dwmBorderColor  = 34
	dwmCaptionColor = 35
	dwmTextColor    = 36
)

var nativeDWM = syscall.NewLazyDLL("dwmapi.dll").NewProc("DwmSetWindowAttribute")
var nativeUser32 = syscall.NewLazyDLL("user32.dll")
var nativeSendMessage = nativeUser32.NewProc("SendMessageW")
var nativeCreateIcon = nativeUser32.NewProc("CreateIcon")

const (
	wmSetIcon = 0x0080
	iconSmall = 0
	iconBig   = 1
)

// removeNativeWindowIcon replaces only the automatic caption icon with a
// transparent native icon. Sending a null icon makes Windows fall back to the
// executable resource, so it would not actually remove anything. The system
// menu and the minimize/maximize/close buttons remain native and accessible.
func removeNativeWindowIcon(hwnd unsafe.Pointer) {
	if hwnd == nil || nativeSendMessage.Find() != nil || nativeCreateIcon.Find() != nil {
		return
	}
	// A monochrome 16×16 icon has 32 bytes per mask. An all-one AND mask and
	// all-zero XOR mask draw every pixel transparently.
	andMask := [32]byte{}
	for i := range andMask {
		andMask[i] = 0xff
	}
	xorMask := [32]byte{}
	transparent, _, _ := nativeCreateIcon.Call(0, 16, 16, 1, 1,
		uintptr(unsafe.Pointer(&andMask[0])), uintptr(unsafe.Pointer(&xorMask[0])))
	if transparent == 0 {
		return
	}
	for _, size := range []uintptr{iconSmall, iconBig} {
		_, _, _ = nativeSendMessage.Call(uintptr(hwnd), wmSetIcon, size, transparent)
	}
}

// Windows owns the minimize/maximize/close buttons. Color only its native
// caption and border, retaining accessibility and standard window behavior.
func setNativeTitlebar(hwnd unsafe.Pointer, theme string) {
	if hwnd == nil || nativeDWM.Find() != nil {
		return
	}
	dark, caption, titleText, border := uint32(1), nativeRGB(9, 13, 29), nativeRGB(235, 241, 255), nativeRGB(42, 53, 88)
	if theme == "light" {
		dark, caption, titleText, border = 0, nativeRGB(245, 247, 252), nativeRGB(23, 33, 55), nativeRGB(180, 191, 214)
	}
	for _, item := range []struct {
		attr  uintptr
		value uint32
	}{
		{dwmDarkMode, dark},
		{dwmCaptionColor, caption},
		{dwmTextColor, titleText},
		{dwmBorderColor, border},
	} {
		value := item.value
		_, _, _ = nativeDWM.Call(uintptr(hwnd), item.attr, uintptr(unsafe.Pointer(&value)), unsafe.Sizeof(value))
	}
}

// Win32 COLORREF stores the red channel in the least significant byte.
func nativeRGB(r, g, b byte) uint32 {
	return uint32(r) | uint32(g)<<8 | uint32(b)<<16
}
