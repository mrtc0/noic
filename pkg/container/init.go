package container

import (
	"os"
	"path/filepath"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func Init(ctx *cli.Context) error {
	command := []string{"/bin/sleep", "30"}
	hostname := "sandbox"

	if err := syscall.Sethostname([]byte(hostname)); err != nil {
		return err
	}

	if err := setupMount(); err != nil {
		logrus.Error("Failed setupMount")
		return err
	}

	if err := syscall.Exec(command[0], command[0:], []string{}); err != nil {
		return err
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
