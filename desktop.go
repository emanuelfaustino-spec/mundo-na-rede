package main

import (
	"os"
	"os/exec"
	"path/filepath"
)

func openWindowsPanel(url string) {
	for _, variable := range []string{"ProgramFiles(x86)", "ProgramFiles", "LOCALAPPDATA"} {
		root := os.Getenv(variable)
		if root == "" {
			continue
		}
		edge := filepath.Join(root, "Microsoft", "Edge", "Application", "msedge.exe")
		if _, err := os.Stat(edge); err == nil {
			// App mode removes the browser chrome. Release builds use the Windows
			// GUI subsystem, so launching Mundo Nexo never needs a console.
			if err = exec.Command(edge, "--app="+url, "--no-first-run", "--disable-features=TranslateUI").Start(); err == nil {
				return
			}
		}
	}
	exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Run()
}
