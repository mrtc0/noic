package container

import (
	"errors"

	"github.com/urfave/cli"
)

func Start(context *cli.Context) error {
	id := context.Args().First()
	if id == "" {
		return errors.New("container id cannnot be empty")
	}

	c, err := FindByID(id)
	if err != nil {
		return err
	}

	c.Run()

	return nil
}
