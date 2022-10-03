package cmd

import (
	"github.com/mrtc0/noic/pkg/container"
	"github.com/urfave/cli"
)

var CreateCommand = cli.Command{
	Name:  "create",
	Usage: "create a container",
	ArgsUsage: `<container-id>

Where "<container-id>" is your name for instance of the container.
	`,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "bundle, b",
			Value: ".",
			Usage: `path to the bundle directory, defaults to the current directory`,
		},
	},
	Action: func(context *cli.Context) error {
		c, err := container.Create(context)
		if err != nil {
			return err
		}

		c.Run()

		return nil
	},
}
