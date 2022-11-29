package container

import (
	"fmt"
	"os"
	"path/filepath"

	specsgo "github.com/opencontainers/runtime-spec/specs-go"
)

type ContainerFactory struct {
	ContainerID        string
	StateRootDirectory string
	BundlePath         string
	UseSystemdCgroups  bool
	ConsoleSocket      string
}

func (f *ContainerFactory) Create() (*Container, error) {
	if f.BundlePath != "" {
		if err := os.Chdir(f.BundlePath); err != nil {
			return nil, fmt.Errorf("faild chdir %s: %v", f.BundlePath, err)
		}
	}

	spec, err := LoadSpec(DefaultSpecConfigFilename)
	if err != nil {
		return nil, err
	}

	/*
		_, err = os.Stat(containerRoot)
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%s container root directory does not exists: %v", spec.Root.Path, err)
		}
	*/

	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	p := filepath.Join(f.StateRootDirectory, f.ContainerID)
	if err := os.MkdirAll(p, 0o700); err != nil {
		return nil, err
	}

	rootfsPath := spec.Root.Path
	if !filepath.IsAbs(rootfsPath) {
		rootfsPath = filepath.Join(cwd, rootfsPath)
	}

	execFifoPath := filepath.Join(f.StateRootDirectory, f.ContainerID, execFifoFilename)
	c := &Container{
		ID:           f.ContainerID,
		Root:         rootfsPath,
		ExecFifoPath: execFifoPath,
		Spec:         spec,
		State: specsgo.State{
			Version:     spec.Version,
			ID:          f.ContainerID,
			Status:      "creating",
			Bundle:      f.BundlePath,
			Annotations: map[string]string{},
		},
		StateRootDirectory: f.StateRootDirectory,
		UseSystemdCgroups:  f.UseSystemdCgroups,
		ConsoleSocket:      f.ConsoleSocket,
	}

	return c, nil
}
