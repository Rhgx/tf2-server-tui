//go:build linux

package steam

import (
	"fmt"
	"os/exec"
)

func Launch(ip string, port int) error {
	url := fmt.Sprintf("steam://connect/%s:%d", ip, port)
	return startWithAvailableCommand(
		[]string{"xdg-open", url},
		[]string{"gio", "open", url},
		[]string{"steam", url},
	)
}

func OpenPath(path string) error {
	return startWithAvailableCommand(
		[]string{"xdg-open", path},
		[]string{"gio", "open", path},
	)
}

func startWithAvailableCommand(commands ...[]string) error {
	for _, command := range commands {
		if len(command) == 0 {
			continue
		}

		if _, err := exec.LookPath(command[0]); err != nil {
			continue
		}

		return exec.Command(command[0], command[1:]...).Start()
	}

	return fmt.Errorf("no supported opener found; install xdg-open, gio, or steam")
}
