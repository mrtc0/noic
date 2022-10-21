package container

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/mrtc0/noic/pkg/container/cgroups"
	"github.com/mrtc0/noic/pkg/container/seccomp"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/sirupsen/logrus"
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

		pid := os.Getpid()
		// pid := container.InitProcess.Pid
		if err := mgr.Add(uint64(pid)); err != nil {
			fmt.Println(err)
			return fmt.Errorf("failed add process to cgroup: %s", err)
		}
	}

	if err := setupMount(); err != nil {
		logrus.Error("Failed setupMount")
		return err
	}

	if err := readonlyPathMount(container.Spec.Linux.ReadonlyPaths); err != nil {
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

func createDevices(devices []specs.LinuxDevice, path string) error {
	for _, device := range devices {
		dest := filepath.Join(path, device.Path)
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}

	}
}

func mknodDevice(dest string, device *specs.LinuxDevice) {
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

	mountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	if err := syscall.Mount("proc", "/proc", "proc", uintptr(mountFlags), ""); err != nil {
		return err
	}
	if err := syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=755"); err != nil {
		return err
	}

	return nil
}
