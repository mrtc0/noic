package process

import (
	"os"
	"os/exec"
	"syscall"

	"github.com/sirupsen/logrus"
)

func newPipe() (*os.File, *os.File, error) {
	r, w, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}

	return r, w, nil
}

func NewParentProcess(rootfs string, env []string) (*exec.Cmd, *os.File, error) {
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
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.ExtraFiles = []*os.File{readPipe}

	cmd.Dir = rootfs
	cmd.Env = append(os.Environ(), env...)

	return cmd, writePipe, nil
}
