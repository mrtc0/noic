package container

import (
	"os"
	"os/exec"
	"syscall"

	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/sirupsen/logrus"
)

func newPipe() (*os.File, *os.File, error) {
	r, w, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}

	return r, w, nil
}

func (c Container) NewParentProcess() (*exec.Cmd, *os.File, error) {
	readPipe, writePipe, err := newPipe()
	if err != nil {
		return nil, nil, err
	}

	initCmd, err := os.Readlink("/proc/self/exe")
	if err != nil {
		logrus.Error("Failed readlink /proc/self/exe")
		return nil, nil, err
	}

	cmd := exec.Command(initCmd, "init")
	/*
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
		}
	*/

	attr, err := sysProcAttr(c.Spec.Linux.Namespaces)
	if err != nil {
		return nil, nil, err
	}

	cmd.SysProcAttr = attr

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.ExtraFiles = []*os.File{readPipe}

	cmd.Dir = c.Root
	cmd.Env = append(os.Environ(), c.Spec.Process.Env...)

	return cmd, writePipe, nil
}

func sysProcAttr(namespaces []specs.LinuxNamespace) (*syscall.SysProcAttr, error) {
	var flags uintptr
	flags = 0

	for _, namespace := range namespaces {
		switch namespace.Type {
		case "pid":
			flags = flags | syscall.CLONE_NEWPID
		case "network":
			flags = flags | syscall.CLONE_NEWNET
		case "mount":
			flags = flags | syscall.CLONE_NEWNS
		case "ipc":
			flags = flags | syscall.CLONE_NEWIPC
		case "uts":
			flags = flags | syscall.CLONE_NEWUTS
			/*
				case "user":
					flags = flags | syscall.CLONE_NEWUSER
				case "cgroup":
					flags = flags | 0x2000000
			*/
		}
	}

	return &syscall.SysProcAttr{Cloneflags: flags}, nil
}
