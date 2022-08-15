package container

import (
	"errors"

	"github.com/urfave/cli"
)

func CreateContainer(context *cli.Context) (*Container, error) {
	rootfs := context.String("rootfs")
	c, err := Create(rootfs)
	return c, err
}

func Create(rootfs string) (*Container, error) {
	if rootfs == "" {
		return nil, errors.New("rootfs not set")
	}

	c := &Container{rootfs: rootfs}
	c.state = Created
	return c, nil
}
