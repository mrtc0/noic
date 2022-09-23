package container

import (
	"os"
	"path/filepath"

	"github.com/mrtc0/noic/pkg/specs"
	"github.com/urfave/cli"
)

type ContainerFactory struct {
}

func loadFactory(context *cli.Context) (*ContainerFactory, error) {
	return &ContainerFactory{}, nil
}

func (f *ContainerFactory) Create(id string, spec *specs.Spec) (*Container, error) {
	p, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	containerRoot := filepath.Join(p, spec.Root.Path)
	_, err = os.Stat(containerRoot)
	if os.IsNotExist(err) {
		return nil, err
	}

	c := &Container{ID: id, Root: containerRoot, Spec: spec}
	c.State = Created

	return c, nil
}
