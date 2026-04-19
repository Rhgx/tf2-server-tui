//go:build !windows && !linux

package console

func SetTitle(string) error {
	return nil
}
