package mount

import (
	"os"
	"syscall"
)

func MountConsole(slavePath string) error {
	m := syscall.Umask(0o000)
	defer syscall.Umask(m)

	f, err := os.Create("/dev/console")
	if err != nil && !os.IsExist(err) {
		return err
	}

	if f != nil {
		f.Close()
	}

	return syscall.Mount(slavePath, "/dev/console", "bind", syscall.MS_BIND, "")
}
