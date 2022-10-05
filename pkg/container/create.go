package container

import (
	"errors"
	"fmt"
	"os"

	"github.com/mrtc0/noic/pkg/specs"
	"github.com/urfave/cli"
)

func Create(context *cli.Context) (*Container, error) {
	id := context.Args().First()
	if id == "" {
		return nil, errors.New("container id cannnot be empty")
	}

	_, err := FindByID(id)
	if err == nil {
		return nil, fmt.Errorf("container %s is exists", id)
	}

	bundle := context.String("bundle")
	if bundle != "" {
		if err := os.Chdir(bundle); err != nil {
			return nil, err
		}
	}

	spec, err := specs.LoadSpec(specs.DefaultSpecConfigFilename)
	if err != nil {
		return nil, err
	}

	c, err := newContainer(context, id, spec)
	if err != nil {
		return nil, fmt.Errorf("failed create container: %v", err)
	}

	return c, err
}
