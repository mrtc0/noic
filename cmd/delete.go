package cmd

import (
	"errors"
	"fmt"

	"github.com/mrtc0/noic/pkg/container"
	"github.com/urfave/cli"
)

var DeleteCommand = cli.Command{
	Name:  "delete",
	Usage: "delete a container",
	ArgsUsage: `<container-id>

Where "<container-id>" is your name for instance of the container.
	`,
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "force, f",
			Usage: "Forcibly deletes the container if it is still running (uses SIGKILL)",
		},
	},
	Action: func(context *cli.Context) error {
		id := context.Args().First()
		force := context.Bool("force")

		// Not supported
		if force {
			return nil
		}

		if id == "" {
			return errors.New("container id cannnot be empty")
		}

		stateRootDirectory := context.GlobalString("root")
		c, err := container.FindByID(id, stateRootDirectory)
		if err != nil {
			return err
		}

		if c.CurrentStatus() != container.Stopped {
			return fmt.Errorf("container is not stoppped")
		}

		return c.Destroy()
	},
}
