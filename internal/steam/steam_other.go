//go:build !windows && !linux

package steam

import "fmt"

func Launch(ip string, port int) error {
	return fmt.Errorf("steam launching is currently implemented for Windows and Linux only")
}

func OpenPath(path string) error {
	return fmt.Errorf("opening files is currently implemented for Windows and Linux only")
}
