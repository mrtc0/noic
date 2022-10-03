package container

import (
	"errors"
	"os"

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

	_, err = os.OpenFile(c.ExecFifoPath, os.O_RDONLY, 0)
	if err != nil {
		return err
	}

	if err = os.Remove(c.ExecFifoPath); err != nil {
		return err
	}

	return nil
}
