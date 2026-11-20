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
