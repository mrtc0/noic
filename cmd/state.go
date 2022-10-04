package cmd

import (
	"errors"
	"fmt"

	"github.com/mrtc0/noic/pkg/container"
	"github.com/urfave/cli"
)

var StateCommand = cli.Command{
	Name:  "state",
	Usage: "state of contaienr",
	ArgsUsage: `<container-id>

Where "<container-id>" is your name for instance of the container.
	`,
	Action: func(context *cli.Context) error {
		id := context.Args().First()
		if id == "" {
			return errors.New("container id cannnot be empty")
		}

		container, err := container.FindByID(id)
		if err != nil {
			return err
		}

		fmt.Println(container.CurrentStatus())

		return nil
	},
}
