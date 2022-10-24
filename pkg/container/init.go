package container

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/mrtc0/noic/pkg/container/capabilities"
	"github.com/mrtc0/noic/pkg/container/cgroups"
	"github.com/mrtc0/noic/pkg/container/mount"
	"github.com/mrtc0/noic/pkg/container/processes"
	"github.com/mrtc0/noic/pkg/container/seccomp"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/urfave/cli"
	"golang.org/x/sys/unix"
)

func Init(ctx *cli.Context, pipe *os.File) error {
	var container *Container
	if err := json.NewDecoder(pipe).Decode(&container); err != nil {
		return err
	}

	if err := awaitStart(container.ExecFifoPath); err != nil {
		return err
	}

	command := container.Spec.Process.Args
	hostname := container.Spec.Hostname

	if err := syscall.Sethostname([]byte(hostname)); err != nil {
		return err
	}

	pid := os.Getpid()
	if container.Spec.Process.Rlimits != nil {
		if err := processes.SetupRlimits(pid, *container.Spec.Process); err != nil {
			return err
		}
	}

	if container.Spec.Linux.Resources != nil {
		// TODO: support container.Spec.Linux.CgroupsPath
		mountpoint := ""
		if container.Spec.Linux.CgroupsPath != "" {
			mountpoint = container.Spec.Linux.CgroupsPath
		}
		mgr, err := cgroups.New(container.ID, mountpoint, *container.Spec.Linux.Resources)
		if err != nil {
			fmt.Println(err)
			return fmt.Errorf("create cgroup failed: %s", err)
		}

		if err := mgr.Add(uint64(pid)); err != nil {
			return fmt.Errorf("failed add process to cgroup: %s", err)
		}
	}

	if err := setupMount(); err != nil {
		fmt.Printf("Failed setupMount: %s", err)
		return err
	}

	if err := mount.MountFilesystems(container.Spec.Mounts); err != nil {
		return err
	}

	if err := readonlyPathMount(container.Spec.Linux.ReadonlyPaths); err != nil {
		return err
	}

	pwd, err := os.Getwd()
	if err != nil {
		return err
	}
	if err := createDefaultDevices(pwd); err != nil {
		return err
	}
	if err := createDevices(container.Spec.Linux.Devices, pwd); err != nil {
		return err
	}

	if err := processes.SetupNowNewPrivileges(*container.Spec.Process); err != nil {
		return err
	}

	path, err := exec.LookPath(command[0])
	if err != nil {
		return fmt.Errorf("%s not found: %v", command[0], err)
	}

	if container.Spec.Linux.Seccomp != nil {
		if err = seccomp.LoadSeccompProfile(*container.Spec.Linux.Seccomp); err != nil {
			return err
		}
	}

	if container.Spec.Process.Capabilities != nil {
		cap := capabilities.New(*container.Spec.Process.Capabilities)
		if err := cap.Apply(); err != nil {
			return fmt.Errorf("faild apply capabilities: %v", err)
		}
	}

	// Run a container process
	if err := syscall.Exec(path, command[0:], container.Spec.Process.Env); err != nil {
		return err
	}

	return nil
}

func awaitStart(path string) error {
	if err := unix.Mkfifo(path, 0o622); err != nil {
		return fmt.Errorf("mkfifo(%s) failed: %v", path, err)
	}

	_, err := unix.Open(path, unix.O_WRONLY|unix.O_CLOEXEC, 0)
	if err != nil {
		return fmt.Errorf("failed open exec.Fifo file(%s): %v", path, err)
	}

	return nil
}

// https://github.com/opencontainers/runtime-spec/blob/494a5a6aca782455c0fbfc35af8e12f04e98a55e/config-linux.md#default-devices
func createDefaultDevices(path string) error {
	uid := uint32(0)
	gid := uint32(0)
	mode := fs.FileMode(0o666)
	defaultDevices := []specs.LinuxDevice{
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

	if err := setupPtmx(path); err != nil {
		return err
	}

	if err := createDevSymlinks(path); err != nil {
		return err
	}

	return createDevices(defaultDevices, path)
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

func createDevices(devices []specs.LinuxDevice, path string) error {
	for _, device := range devices {
		dest := filepath.Join(path, device.Path)
		if err := mknodDevice(dest, device); err != nil {
			return err
		}
	}

	return nil
}

func mknodDevice(dest string, device specs.LinuxDevice) error {
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

func readonlyPathMount(paths []string) error {
	for _, path := range paths {
		if err := syscall.Mount(path, path, "", unix.MS_BIND|unix.MS_RDONLY, ""); err != nil {
			return err
		}
	}
	return nil
}

func setupMount() error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	if err := syscall.Mount("", "/", "", syscall.MS_REC|syscall.MS_PRIVATE, ""); err != nil {
		return err
	}

	if err := syscall.Mount(pwd, pwd, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return err
	}

	oldDir := filepath.Join(pwd, ".old")
	if err = os.Mkdir(oldDir, 0777); err != nil {
		return err
	}

	if err := syscall.PivotRoot(pwd, oldDir); err != nil {
		return err
	}

	if err := syscall.Chdir("/"); err != nil {
		return err
	}

	if err := syscall.Unmount("/.old", syscall.MNT_DETACH); err != nil {
		return err
	}

	if err := os.Remove("/.old"); err != nil {
		return err
	}

	/*
		mountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
		if err := syscall.Mount("proc", "/proc", "proc", uintptr(mountFlags), ""); err != nil {
			return err
		}
		if err := syscall.Mount("sysfs", "/sys", "sysfs", uintptr(mountFlags), ""); err != nil {
			return err
		}
		/*
			if err := syscall.Mount("tmpfs", "/dev/shm", "tmpfs", uintptr(mountFlags), "mode=1777"); err != nil {
				return err
			}
	*/

	return nil
}
