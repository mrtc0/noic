package cmd

import (
	"errors"
	"os"
	"path/filepath"

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

		path := filepath.Join(container.StateDir, id)
		if err := os.RemoveAll(path); err != nil {
			return err
		}

		return nil
	},
}
