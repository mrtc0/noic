package cmd

import (
	"github.com/mrtc0/noic/pkg/container"
	"github.com/urfave/cli"
)

var RunCommand = cli.Command{
	Name:  "run",
	Usage: "create and run container",
	ArgsUsage: `<rootfs>

Where "<rootfs>" is your container rootfs directory.
	`,
	Description: "The run command creates an instance of a container.",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "rootfs, r",
			Value: ".",
			Usage: `path to the rootfs directory, defaults to the current directory`,
		},
	},
	Action: func(context *cli.Context) error {
		err := container.StartContainer(context)
		if err != nil {
			return err
		}
		return nil
	},
}
