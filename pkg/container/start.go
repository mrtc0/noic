package container

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/urfave/cli"
)

func Start(context *cli.Context) error {
	id := context.Args().First()
	if id == "" {
		return errors.New("container id cannnot be empty")
	}

	path, err := StateFilePath(id)
	if err != nil {
		return err
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var container *Container
	json.Unmarshal(raw, &container)
	c, err := Create(context)
	if err != nil {
		return err
	}

	c.Run()

	return nil
}
