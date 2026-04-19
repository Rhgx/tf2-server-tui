package lockfile

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

type Lock struct {
	path string
	file *os.File
}

// Acquire creates an exclusive single-instance lock file. The lock lives next
// to the executable so accidental double-launches fail fast with a readable
// error in portable installs.
func Acquire(name string) (*Lock, error) {
	if name == "" {
		name = "tf2-server-tui.lock"
	}

	lockPath, err := resolveLockPath(name)
	if err != nil {
		return nil, err
	}

	file, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		if os.IsExist(err) {
			return nil, fmt.Errorf("another copy is already running (%s)", lockPath)
		}
		return nil, err
	}

	if _, err := file.WriteString(strconv.Itoa(os.Getpid())); err != nil {
		file.Close()
		_ = os.Remove(lockPath)
		return nil, err
	}

	return &Lock{
		path: lockPath,
		file: file,
	}, nil
}

// Release closes and removes the lock file created by Acquire.
func (l *Lock) Release() error {
	if l == nil {
		return nil
	}

	if l.file != nil {
		if err := l.file.Close(); err != nil {
			return err
		}
		l.file = nil
	}

	if l.path != "" {
		if err := os.Remove(l.path); err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	return nil
}

// resolveLockPath chooses where the lock file should live. It keeps the lock
// file alongside the shipped binary when possible, which makes portable folder
// copies behave predictably.
func resolveLockPath(name string) (string, error) {
	exePath, err := os.Executable()
	if err == nil {
		return filepath.Join(filepath.Dir(exePath), name), nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return filepath.Join(cwd, name), nil
}
