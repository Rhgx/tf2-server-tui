//go:build windows

package console

import (
	"syscall"
	"unsafe"
)

var (
	kernel32            = syscall.NewLazyDLL("kernel32.dll")
	setConsoleTitleProc = kernel32.NewProc("SetConsoleTitleW")
)

func SetTitle(title string) error {
	ptr, err := syscall.UTF16PtrFromString(title)
	if err != nil {
		return err
	}

	ret, _, callErr := setConsoleTitleProc.Call(uintptr(unsafe.Pointer(ptr)))
	if ret == 0 {
		return callErr
	}

	return nil
}
