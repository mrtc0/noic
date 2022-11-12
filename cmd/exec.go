package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/mrtc0/noic/pkg/container"
	"github.com/urfave/cli"
)

var ExecCommand = cli.Command{
	Name:  "exec",
	Usage: "exec",
	ArgsUsage: `<container-id>

Where "<container-id>" is your name for instance of the container.
	`,
	Action: func(context *cli.Context) error {
		id := context.Args().First()
		if id == "" {
			return errors.New("container id cannnot be empty")
		}

		stateRootDirectory := context.GlobalString("root")
		c, err := container.FindByID(id, stateRootDirectory)
		if err != nil {
			return err
		}

		if c.CurrentStatus() != container.Running {
			return fmt.Errorf("container is not running")
		}

		commands := context.Args()[1:]
		if len(commands) == 0 {
			return fmt.Errorf("command cannot be empty")
		}

		nsexecOptions := []string{"-t", fmt.Sprintf("%d", c.InitProcess.Pid), "-a"}
		nsexecOptions = append(nsexecOptions, commands...)

		nsexec := exec.Command("nsenter", nsexecOptions...)
		nsexec.Stdin = os.Stdin
		nsexec.Stdout = os.Stdout
		nsexec.Stderr = os.Stderr

		return nsexec.Run()
	},
}
