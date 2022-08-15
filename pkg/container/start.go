package container

import (
	"github.com/urfave/cli"
)

func StartContainer(context *cli.Context) error {
	c, err := CreateContainer(context)
	if err != nil {
		return err
	}

	c.Run()

	return nil
}
