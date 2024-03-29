package cmd

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/mrtc0/noic/pkg/container"
	"github.com/opencontainers/runtime-spec/specs-go"
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

		stateRootDirectory := context.GlobalString("root")
		c, err := container.FindByID(id, stateRootDirectory)
		if err != nil {
			return err
		}

		c.State.Status = specs.ContainerState(c.CurrentStatus().String())
		j, err := json.Marshal(c.State)
		if err != nil {
			return err
		}
		fmt.Println(string(j))

		return nil
	},
}
