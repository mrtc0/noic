package processes

import (
	"fmt"
	"os"
	"strconv"
	"syscall"
	"unsafe"

	"github.com/mrtc0/noic/pkg/container/mount"
)

func SetupConsole(socket *os.File) error {
	defer socket.Close()

	pty, err := openPty()
	if err != nil {
		return fmt.Errorf("failed openPty: %s", err)
	}

	defer pty.Master.Close()

	if err := mount.MountConsole(pty.SlavePath); err != nil {
		return fmt.Errorf("failed console mount %s: %s", pty.SlavePath, err)
	}

	oob := syscall.UnixRights(int(pty.Master.Fd()))
	if err := syscall.Sendmsg(int(socket.Fd()), []byte(pty.Master.Name()), oob, nil, 0); err != nil {
		return fmt.Errorf("failed sendmsg: %s", err)
	}

	return dupStdio(pty.SlavePath)
}

type PTY struct {
	Master    *os.File
	SlavePath string
}

func (p *PTY) PTSName() (string, error) {
	n, err := p.PTSNumber()
	if err != nil {
		return "", err
	}
	return "/dev/pts/" + strconv.Itoa(int(n)), nil
}

func (p *PTY) PTSNumber() (uint, error) {
	var ptyno uint
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(p.Master.Fd()), uintptr(syscall.TIOCGPTN), uintptr(unsafe.Pointer(&ptyno)))
	if errno != 0 {
		return 0, errno
	}
	return ptyno, nil
}

func openPty() (*PTY, error) {
	master, err := os.OpenFile("/dev/ptmx", os.O_RDWR, 0)
	if err != nil {
		return nil, err
	}

	pty := &PTY{Master: master}
	slaveName, err := pty.PTSName()
	if err != nil {
		master.Close()
		return nil, err
	}

	var u uint
	if _, _, err := syscall.Syscall(syscall.SYS_IOCTL, uintptr(master.Fd()), uintptr(syscall.TIOCSPTLCK), uintptr(unsafe.Pointer(&u))); err != 0 {
		return nil, err
	}

	pty.SlavePath = slaveName
	return pty, nil
}

func dupStdio(slavePath string) error {
	fd, err := syscall.Open(slavePath, syscall.O_RDWR, 0)
	if err != nil {
		return &os.PathError{
			Op:   "open",
			Path: slavePath,
			Err:  err,
		}
	}
	for _, i := range []int{0, 1, 2} {
		if err := syscall.Dup3(fd, i, 0); err != nil {
			return err
		}
	}
	return nil
}
