//go:build linux

package console

import (
	"fmt"
	"os"
)

func SetTitle(title string) error {
	_, err := fmt.Fprintf(os.Stdout, "\033]0;%s\007", title)
	return err
}
