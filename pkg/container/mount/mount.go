package mount

import (
	"fmt"
	"os"
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

func MountFilesystems(mounts []specs.Mount) error {
	for _, mnt := range mounts {
		if mnt.Destination == "" {
			return fmt.Errorf("invalid destination of mount point")
		}

		// TODO: cgroup mount
		if mnt.Type == "cgroup" {
			continue
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

		if mnt.Type == "bind" {
			// TODO: 動かない...
			continue
		}

		if err := os.MkdirAll(mnt.Destination, 0o755); err != nil {
			return fmt.Errorf("failed create directory: %s", mnt.Destination)
		}

		if err := syscall.Mount(mnt.Source, mnt.Destination, mnt.Type, uintptr(flags), strings.Join(labels, ",")); err != nil {
			return fmt.Errorf("failed mount. source: %s, destination: %s, type: %s, %v", mnt.Source, mnt.Destination, mnt.Type, err)
		}
	}

	return nil
}
