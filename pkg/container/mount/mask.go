package mount

import (
	"fmt"
	"os"
	"syscall"
)

func MaskPath(path string) error {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil
		// return syscall.Mount("/dev/null", path, "", syscall.MS_BIND, "")
	}

	if fileInfo.IsDir() {
		return syscall.Mount("tmpfs", path, "tmpfs", syscall.MS_RDONLY, "")
	}

	if err := syscall.Mount("/dev/null", path, "", syscall.MS_BIND, ""); err != nil {
		return fmt.Errorf("mount failed: %s(%s), %s", path, fileInfo, err)
	}

	return nil
}
