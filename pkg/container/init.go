package container

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/mrtc0/noic/pkg/container/apparmor"
	"github.com/mrtc0/noic/pkg/container/capabilities"
	"github.com/mrtc0/noic/pkg/container/mount"
	"github.com/mrtc0/noic/pkg/container/processes"
	"github.com/mrtc0/noic/pkg/container/seccomp"
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

	if err := apparmor.ApplyProfile(container.Spec.Process.ApparmorProfile); err != nil {
		return err
	}

	pid := os.Getpid()
	if container.Spec.Process.Rlimits != nil {
		if err := processes.SetupRlimits(pid, *container.Spec.Process); err != nil {
			return err
		}
	}

	/*
		TODO: Support cgroups

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
	*/

	if err := mount.MountRootFs(container.Root, container.Spec); err != nil {
		return err
	}

	if err := readonlyPathMount(container.Spec.Linux.ReadonlyPaths); err != nil {
		return err
	}

	if container.Spec.Process.NoNewPrivileges {
		if err := processes.SetupNowNewPrivileges(); err != nil {
			return err
		}
	}

	if container.Spec.Linux.Seccomp != nil {
		if err := seccomp.LoadSeccompProfile(*container.Spec.Linux.Seccomp); err != nil {
			return err
		}
	}

	if container.Spec.Process.OOMScoreAdj != nil {
		if err := processes.ApplyOOMScoreAdj(*container.Spec.Process.OOMScoreAdj); err != nil {
			return err
		}
	}

	if container.Spec.Process.Capabilities != nil {
		cap := capabilities.New(*container.Spec.Process.Capabilities)
		if err := cap.Apply(); err != nil {
			return fmt.Errorf("failed apply capabilities: %v", err)
		}
	}

	path, err := exec.LookPath(command[0])
	if err != nil {
		return fmt.Errorf("%s not found: %v", command[0], err)
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

func readonlyPathMount(paths []string) error {
	for _, path := range paths {
		if err := syscall.Mount(path, path, "", unix.MS_BIND|unix.MS_REC, ""); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return err
		}

		var s unix.Statfs_t
		if err := unix.Statfs(path, &s); err != nil {
			return &os.PathError{Op: "statfs", Path: path, Err: err}
		}
		flags := uintptr(s.Flags) & (unix.MS_NOSUID | unix.MS_NODEV | unix.MS_NOEXEC)

		if err := syscall.Mount(path, path, "", flags|unix.MS_BIND|unix.MS_REMOUNT|unix.MS_RDONLY, ""); err != nil {
			return err
		}
	}

	return nil
}

/*
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

	return nil
}
*/
