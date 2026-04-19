//go:build windows

package steam

import (
	"fmt"
	"os/exec"
)

func Launch(ip string, port int) error {
	url := fmt.Sprintf("steam://connect/%s:%d", ip, port)
	return exec.Command("rundll32.exe", "url.dll,FileProtocolHandler", url).Start()
}

func OpenPath(path string) error {
	return exec.Command("cmd", "/c", "start", "", path).Start()
}
