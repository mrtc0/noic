package container

import (
	"errors"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func Start(context *cli.Context) error {
	id := context.Args().First()
	if id == "" {
		return errors.New("container id cannnot be empty")
	}

	stateRootDirectory := context.GlobalString("root")
	c, err := FindByID(id, stateRootDirectory)
	if err != nil {
		return fmt.Errorf("container %s is not found: %s", id, err)
	}

	_, err = os.OpenFile(c.ExecFifoPath, os.O_RDONLY, 0)
	if err != nil {
		return fmt.Errorf("failed open execFifo %s: %v", c.ExecFifoPath, err)
	}

	if err = os.Remove(c.ExecFifoPath); err != nil {
		return fmt.Errorf("failed remove execFifo %s: %v", c.ExecFifoPath, err)
	}

	logrus.Debugf("started container %s", id)

	return nil
}
