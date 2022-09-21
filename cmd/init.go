package cmd

import (
	"github.com/mrtc0/noic/pkg/container"
	"github.com/urfave/cli"
)

var InitCommand = cli.Command{
	Name:  "init",
	Usage: "init container process",
	Action: func(context *cli.Context) error {
		err := container.Init(context)
		if err != nil {
			return err
		}
		return nil
	},
}
