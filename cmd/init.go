package cmd

import (
	"fmt"
	"os"

	"github.com/mrtc0/noic/pkg/container"
	"github.com/urfave/cli"
)

var InitCommand = cli.Command{
	Name:  "init",
	Usage: "init container process",
	Action: func(context *cli.Context) error {
		pipe := os.NewFile(uintptr(3), "pipe")
		defer pipe.Close()
		err := container.Init(context, pipe)
		if err != nil {
			fmt.Printf("init error: %s", err)
			return err
		}
		return nil
	},
}
