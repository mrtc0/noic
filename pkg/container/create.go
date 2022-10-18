package container

import (
	"errors"
	"fmt"
	"os"

	specsgo "github.com/opencontainers/runtime-spec/specs-go"
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

	spec, err := LoadSpec(DefaultSpecConfigFilename)
	if err != nil {
		return nil, err
	}

	c, err := newContainer(context, id, spec)
	if err != nil {
		return nil, fmt.Errorf("failed create container: %v", err)
	}

	c.State = specsgo.State{
		Version:     spec.Version,
		ID:          id,
		Status:      "creating",
		Bundle:      bundle,
		Annotations: map[string]string{},
	}

	return c, err
}
