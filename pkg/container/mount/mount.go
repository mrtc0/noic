package mount

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/opencontainers/runtime-spec/specs-go"
)

var (
	mountFlags = map[string]int{
		"nosuid":      syscall.MS_NOSUID,
		"nodev":       syscall.MS_NODEV,
		"noexec":      syscall.MS_NOEXEC,
		"strictatime": syscall.MS_STRICTATIME,
		"relatime":    syscall.MS_RELATIME,
		"ro":          syscall.MS_RDONLY,
		"bind":        syscall.MS_BIND,
		"rbind":       syscall.MS_BIND | syscall.MS_REC,
		"rprivate":    syscall.MS_PRIVATE | syscall.MS_REC,
	}
)

func MountFilesystems(rootfs string, mounts []specs.Mount) error {
	for _, mnt := range mounts {
		if mnt.Destination == "" {
			return fmt.Errorf("invalid destination of mount point")
		}

		var flags int
		var labels []string
		for _, option := range mnt.Options {
			if flag, exists := mountFlags[option]; exists {
				flags |= flag
			} else {
				labels = append(labels, option)
			}
		}

		dest := path.Join(rootfs, mnt.Destination)
		switch mnt.Type {
		case "cgroup":
			continue
		case "bind":
			if err := bindMount(mnt.Source, dest, uintptr(flags), strings.Join(labels, ",")); err != nil {
				return fmt.Errorf("failed bind mount %s: %s", mnt.Source, err)
			}
		default:
			if err := os.MkdirAll(dest, 0o755); err != nil {
				return fmt.Errorf("failed create directory: %s", mnt.Destination)
			}

			if err := syscall.Mount(mnt.Source, dest, mnt.Type, uintptr(flags), strings.Join(labels, ",")); err != nil {
				return fmt.Errorf("failed mount. source: %s, destination: %s, type: %s, %v", mnt.Source, mnt.Destination, mnt.Type, err)
			}
		}
	}

	return nil
}

func bindMount(source, destination string, flags uintptr, data string) error {
	stat, err := os.Stat(source)
	if err != nil {
		return err
	}

	if err := createFileOrDirectory(destination, stat.IsDir()); err != nil {
		return err
	}

	if err := syscall.Mount(source, destination, "bind", flags, data); err != nil {
		return fmt.Errorf("failed mount. source: %s, destination: %s, type: bind, %v", source, destination, err)
	}

	return nil
}

func createFileOrDirectory(path string, isDir bool) error {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			if isDir {
				return os.MkdirAll(path, 0o755)
			}
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				return err
			}

			f, err := os.OpenFile(path, os.O_CREATE, 0o755)
			if err != nil {
				return err
			}

			f.Close()
		}
	}

	return nil
}
