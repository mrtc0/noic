package apparmor

import (
	"errors"
	"fmt"
	"os"
)

func ApplyProfile(name string) error {
	if name == "" {
		return nil
	}

	apparmorExecPath := "/proc/self/attr/apparmor/exec"
	// apparmorExecPath := fmt.Sprintf("/proc/%d/attr/apparmor/exec", pid)
	if _, err := os.Stat(apparmorExecPath); errors.Is(err, os.ErrNotExist) {
		apparmorExecPath = "/proc/self/attr/exec"
	}

	f, err := os.OpenFile(apparmorExecPath, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(fmt.Sprintf("exec %s", name))
	return err
}
