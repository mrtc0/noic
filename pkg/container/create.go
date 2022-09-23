package container

import (
	"errors"
	"os"

	"github.com/mrtc0/noic/pkg/specs"
	"github.com/urfave/cli"
)

func Create(context *cli.Context) (*Container, error) {
	id := context.Args().First()
	if id == "" {
		return nil, errors.New("container id cannnot be empty")
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

	container, err := newContainer(context, id, spec)
	if err != nil {
		return nil, err
	}

	if err = saveState(container); err != nil {
		return nil, err
	}

	return container, err
}
