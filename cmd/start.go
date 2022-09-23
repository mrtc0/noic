package cmd

import (
	"github.com/mrtc0/noic/pkg/container"
	"github.com/urfave/cli"
)

var StartCommand = cli.Command{
	Name:      "start",
	Usage:     "start container",
	ArgsUsage: `<container-id>`,
	Action: func(context *cli.Context) error {
		err := container.Start(context)
		if err != nil {
			return err
		}
		return nil
	},
}
