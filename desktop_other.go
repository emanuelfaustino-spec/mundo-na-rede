//go:build !windows

package main

// The native WebView2 window is Windows-only. Other platforms open the local
// dashboard via the platform-specific browser launcher in main.
func openWindowsPanel(url string) {}
