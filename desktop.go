package main

import (
	"os"
	"path/filepath"
	"runtime"

	webview2 "github.com/jchv/go-webview2"
)

func openWindowsPanel(url string) {
	// WebView2 uses COM and its window message loop must stay on one OS thread.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	data, _ := os.UserCacheDir()
	w := webview2.NewWithOptions(webview2.WebViewOptions{
		DataPath: filepath.Join(data, "MundoNexus", "webview"), AutoFocus: true,
		WindowOptions: webview2.WindowOptions{Title: "Mundo Nexus", Width: 1120, Height: 760, Center: true},
	})
	defer w.Destroy()
	w.Init(`window.addEventListener('keydown', e => { if (e.key === 'F5') e.preventDefault(); }, true);`)
	w.Navigate(url)
	w.Run()
}
