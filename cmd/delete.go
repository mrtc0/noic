package cmd

import (
	"errors"

	"github.com/mrtc0/noic/pkg/container"
	"github.com/urfave/cli"
)

var DeleteCommand = cli.Command{
	Name:  "delete",
	Usage: "delete a container",
	ArgsUsage: `<container-id>

Where "<container-id>" is your name for instance of the container.
	`,
	Action: func(context *cli.Context) error {
		id := context.Args().First()
		if id == "" {
			return errors.New("container id cannnot be empty")
		}

		c, err := container.FindByID(id)
		if err != nil {
			return err
		}

		if c.CurrentStatus() == container.Stopped {
			return c.Destroy()
		}

		if err = c.Kill(); err != nil {
			return err
		}

		return c.Destroy()
	},
}
