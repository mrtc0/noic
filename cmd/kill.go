package cmd

import (
	"errors"
	"fmt"

	"github.com/mrtc0/noic/pkg/container"
	"github.com/urfave/cli"
)

var KillCommand = cli.Command{
	Name:  "kill",
	Usage: "kill a container",
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

		if c.CurrentStatus() == container.Stopped {
			return fmt.Errorf("container is not running")
		}

		return c.Kill()
	},
}
