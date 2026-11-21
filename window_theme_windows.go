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
	wmSetIcon       = 0x0080
	iconSmall       = 0
	iconBig         = 1
	imageIcon       = 1
)

var nativeDWM = syscall.NewLazyDLL("dwmapi.dll").NewProc("DwmSetWindowAttribute")
var nativeUser32 = syscall.NewLazyDLL("user32.dll")
var nativeSendMessage = nativeUser32.NewProc("SendMessageW")
var nativeLoadImage = nativeUser32.NewProc("LoadImageW")
var nativeKernel32 = syscall.NewLazyDLL("kernel32.dll")
var nativeGetModuleHandle = nativeKernel32.NewProc("GetModuleHandleW")
var nativeShell32 = syscall.NewLazyDLL("shell32.dll")
var nativeSetAppID = nativeShell32.NewProc("SetCurrentProcessExplicitAppUserModelID")

// setTaskbarAppID gives Windows a stable identity for this executable. Without
// it, WebView2 windows can be grouped under a transient host identity, making
// a minimized window hard to restore from a pinned taskbar button.
func setTaskbarAppID() {
	if nativeSetAppID.Find() != nil {
		return
	}
	id, err := syscall.UTF16PtrFromString("MundoNexus.Desktop")
	if err != nil {
		return
	}
	_, _, _ = nativeSetAppID.Call(uintptr(unsafe.Pointer(id)))
}

// setNativeWindowIcon assigns the executable icon to WebView2's native host
// window. WebView2 otherwise uses its generic window-class icon in the title
// bar and taskbar preview even when the executable itself has a proper ICO.
func setNativeWindowIcon(hwnd unsafe.Pointer) {
	if hwnd == nil || nativeSendMessage.Find() != nil || nativeLoadImage.Find() != nil || nativeGetModuleHandle.Find() != nil {
		return
	}
	module, _, _ := nativeGetModuleHandle.Call(0)
	if module == 0 {
		return
	}
	for _, entry := range []struct{ kind, size uintptr }{{iconSmall, 16}, {iconBig, 32}} {
		icon, _, _ := nativeLoadImage.Call(module, 1, imageIcon, entry.size, entry.size, 0)
		if icon != 0 {
			_, _, _ = nativeSendMessage.Call(uintptr(hwnd), wmSetIcon, entry.kind, icon)
		}
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
