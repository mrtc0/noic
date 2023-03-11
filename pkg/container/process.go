package container

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"syscall"

	"github.com/opencontainers/runtime-spec/specs-go"
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

	if _, err := exec.LookPath("/proc/self/exe"); err != nil {
		return nil, nil, err
	}

	args := []string{os.Args[0], "init"}
	cmd := exec.Command("/proc/self/exe", args[1:]...)
	cmd.Args[0] = args[0]

	attr, err := sysProcAttr(c.Spec.Linux.Namespaces)
	if err != nil {
		return nil, nil, err
	}

	cmd.SysProcAttr = attr

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.ExtraFiles = []*os.File{readPipe}
	if c.Spec.Process.Terminal {
		conn, err := net.Dial("unix", c.ConsoleSocket)
		if err != nil {
			return nil, nil, err
		}

		uc, ok := conn.(*net.UnixConn)
		if !ok {
			return nil, nil, fmt.Errorf("casting to UnixConn failed")
		}

		socket, err := uc.File()
		if err != nil {
			return nil, nil, err
		}

		cmd.ExtraFiles = append(cmd.ExtraFiles, socket)
		cmd.Env = append(cmd.Env, "_NOIC_CONSOLE_FD=4")
	}

	cmd.Dir = c.Root
	cmd.Env = append(cmd.Env, c.Spec.Process.Env...)

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
