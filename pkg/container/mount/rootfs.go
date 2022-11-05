package mount

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	specsgo "github.com/opencontainers/runtime-spec/specs-go"
)

var (
	mountPropergationFlags = map[string]int{
		"shared":     syscall.MS_SHARED,
		"slave":      syscall.MS_SLAVE | syscall.MS_REC,
		"private":    syscall.MS_PRIVATE,
		"unbindable": syscall.MS_UNBINDABLE,
		"":           syscall.MS_PRIVATE | syscall.MS_REC,
	}
)

func MountRootFs(rootfs string, spec *specsgo.Spec) error {
	flags, exists := mountPropergationFlags[spec.Linux.RootfsPropagation]
	if !exists {
		return fmt.Errorf("invalid rootfsPropagation: %s", spec.Linux.RootfsPropagation)
	}

	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	if err := syscall.Mount("", "/", "", uintptr(flags), ""); err != nil {
		return fmt.Errorf("failed to mount rootfs: %s", err)
	}

	// bind mount for pivot_root
	if err := syscall.Mount(rootfs, rootfs, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("failed to bind mount for pivot_root: src=%s dest=%s, %s", rootfs, rootfs, err)
	}

	if err := MountFilesystems(rootfs, spec.Mounts); err != nil {
		return err
	}

	if err := createDefaultDevices(rootfs); err != nil {
		return err
	}

	if err := createDevices(spec.Linux.Devices, rootfs); err != nil {
		return err
	}

	oldDir := filepath.Join(pwd, ".old")
	if err = os.Mkdir(oldDir, 0777); err != nil {
		return fmt.Errorf("failed create .old directory: %s", err)
	}

	if err := syscall.PivotRoot(pwd, oldDir); err != nil {
		return fmt.Errorf("failed pivot_root: %s", err)
	}

	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf("failed change directory after pivot_root: %s", err)
	}

	if err := syscall.Unmount("/.old", syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("failed unmount .old after pivot_root: %s", err)
	}

	if err := os.Remove("/.old"); err != nil {
		return fmt.Errorf("failed remove .old after pivot_root: %s", err)
	}

	if err := syscall.Mount("", "/", "", uintptr(flags), ""); err != nil {
		return fmt.Errorf("failed to mount rootfs: %s", err)
	}

	for _, path := range spec.Linux.MaskedPaths {
		if err := MaskPath(path); err != nil {
			return err
		}
	}

	return nil
}
