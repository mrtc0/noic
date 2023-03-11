package container

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"syscall"

	"github.com/mrtc0/noic/pkg/container/apparmor"
	"github.com/mrtc0/noic/pkg/container/capabilities"
	"github.com/mrtc0/noic/pkg/container/cgroups"
	"github.com/mrtc0/noic/pkg/container/mount"
	"github.com/mrtc0/noic/pkg/container/processes"
	"github.com/mrtc0/noic/pkg/container/seccomp"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"golang.org/x/sys/unix"
)

func Init(ctx *cli.Context, pipe *os.File) error {
	logrus.Debug("init start")
	var container *Container
	if err := json.NewDecoder(pipe).Decode(&container); err != nil {
		return err
	}

	if err := awaitStart(container.ExecFifoPath); err != nil {
		return err
	}

	pid := os.Getpid()

	command := container.Spec.Process.Args
	hostname := container.Spec.Hostname

	if err := syscall.Sethostname([]byte(hostname)); err != nil {
		return err
	}

	if err := apparmor.ApplyProfile(container.Spec.Process.ApparmorProfile); err != nil {
		return err
	}

	if container.Spec.Linux.Resources != nil && container.Spec.Linux.CgroupsPath != "" {
		c := &cgroups.CgroupConfig{
			UseSystemd: container.UseSystemdCgroups,
			CgroupPath: container.Spec.Linux.CgroupsPath,
			Resources:  container.Spec.Linux.Resources,
			Name:       container.ID,
			Pid:        container.InitProcess.Pid,
		}

		_, err := cgroups.New(c)
		if err != nil {
			return fmt.Errorf("failed create cgroup: %s", err)
		}
	}

	if container.Spec.Process.Rlimits != nil {
		if err := processes.SetupRlimits(pid, *container.Spec.Process); err != nil {
			return err
		}
	}

	if err := mount.MountRootFs(container.Root, container.Spec); err != nil {
		return err
	}

	if container.Spec.Process.Terminal {
		if envConsole := os.Getenv("_NOIC_CONSOLE_FD"); envConsole != "" {
			console, err := strconv.Atoi(envConsole)
			if err != nil {
				return fmt.Errorf("unable to convert _LIBCONTAINER_CONSOLE: %w", err)
			}
			consoleSocket := os.NewFile(uintptr(console), "console-socket")
			defer consoleSocket.Close()

			if err := processes.SetupConsole(consoleSocket); err != nil {
				return err
			}
		}
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

	if err := os.Chdir(container.Spec.Process.Cwd); err != nil {
		return err
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
