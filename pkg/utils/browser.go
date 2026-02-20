// pkg/utils/browser.go
package utils

import (
	"fmt"
	"os/exec"
	"runtime"
)

// OpenBrowser abre el navegador predeterminado en la URL especificada
func OpenBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	case "linux":
		cmd = "xdg-open"
		args = []string{url}
	default:
		return fmt.Errorf("sistema operativo no soportado: %v", runtime.GOOS)
	}

	return exec.Command(cmd, args...).Start()
}
