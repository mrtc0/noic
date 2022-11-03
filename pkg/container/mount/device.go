package mount

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"

	specsgo "github.com/opencontainers/runtime-spec/specs-go"
	"golang.org/x/sys/unix"
)

// https://github.com/opencontainers/runtime-spec/blob/494a5a6aca782455c0fbfc35af8e12f04e98a55e/config-linux.md#default-devices
func createDefaultDevices(rootfsPath string) error {
	uid := uint32(0)
	gid := uint32(0)
	mode := fs.FileMode(0o666)
	defaultDevices := []specsgo.LinuxDevice{
		{
			Path:     "/dev/null",
			Type:     "c",
			Major:    1,
			Minor:    3,
			FileMode: &mode,
			UID:      &uid,
			GID:      &gid,
		},
		{
			Path:     "/dev/random",
			Type:     "c",
			Major:    1,
			Minor:    8,
			FileMode: &mode,
			UID:      &uid,
			GID:      &gid,
		},
		{
			Path:     "/dev/full",
			Type:     "c",
			Major:    1,
			Minor:    7,
			FileMode: &mode,
			UID:      &uid,
			GID:      &gid,
		},
		{
			Path:     "/dev/tty",
			Type:     "c",
			Major:    5,
			Minor:    0,
			FileMode: &mode,
			UID:      &uid,
			GID:      &gid,
		},
		{
			Path:     "/dev/zero",
			Type:     "c",
			Major:    1,
			Minor:    5,
			FileMode: &mode,
			UID:      &uid,
			GID:      &gid,
		},
		{
			Path:     "/dev/urandom",
			Type:     "c",
			Major:    1,
			Minor:    9,
			FileMode: &mode,
			UID:      &uid,
			GID:      &gid,
		},
	}

	if err := setupPtmx(rootfsPath); err != nil {
		return err
	}

	if err := createDevSymlinks(rootfsPath); err != nil {
		return err
	}

	return createDevices(defaultDevices, rootfsPath)
}

func createDevices(devices []specsgo.LinuxDevice, path string) error {
	for _, device := range devices {
		dest := filepath.Join(path, device.Path)
		if err := mknodDevice(dest, device); err != nil {
			return fmt.Errorf("faild mknod: %s", err)
		}
	}

	return nil
}

func mknodDevice(dest string, device specsgo.LinuxDevice) error {
	fileMode := *device.FileMode
	deviceType := device.Type
	switch deviceType {
	case "c":
		fileMode |= unix.S_IFCHR
	case "b":
		fileMode |= unix.S_IFBLK
	case "p":
		fileMode |= unix.S_IFIFO
	default:
		return fmt.Errorf("invalid device type: %s", deviceType)
	}

	d := unix.Mkdev(uint32(device.Major), uint32(device.Minor))
	if err := syscall.Mknod(dest, uint32(fileMode), int(d)); err != nil {
		return err
	}

	return os.Chown(dest, int(*device.UID), int(*device.GID))
}

func createDevSymlinks(path string) error {
	links := [][2]string{
		{"/proc/self/fd", "/dev/fd"},
		{"/proc/self/fd/0", "/dev/stdin"},
		{"/proc/self/fd/1", "/dev/stdout"},
		{"/proc/self/fd/2", "/dev/stderr"},
	}

	for _, link := range links {
		if err := os.Symlink(link[0], filepath.Join(path, link[1])); err != nil {
			return err
		}
	}

	return nil
}

func setupPtmx(path string) error {
	ptmx := filepath.Join(path, "dev/ptmx")
	if err := os.Remove(ptmx); err != nil && !os.IsNotExist(err) {
		return err
	}
	if err := os.Symlink("pts/ptmx", ptmx); err != nil {
		return err
	}
	return nil
}
