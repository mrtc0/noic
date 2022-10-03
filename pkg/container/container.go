package container

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/mrtc0/noic/pkg/process"
	"github.com/mrtc0/noic/pkg/specs"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

const execFifoFilename = "exec.fifo"

type Container struct {
	ID           string
	Root         string
	ExecFifoPath string
	State        Status
	Spec         *specs.Spec
}

func FindByID(id string) (*Container, error) {
	path, err := StateFilePath(id)
	if err != nil {
		logrus.Debug(err)
		return nil, fmt.Errorf("container %s does not exists", id)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var container *Container
	if err = json.Unmarshal(raw, &container); err != nil {
		return nil, err
	}

	return container, nil
}

func newContainer(context *cli.Context, id string, spec *specs.Spec) (*Container, error) {
	factory, err := loadFactory(context)
	if err != nil {
		return nil, err
	}

	return factory.Create(id, spec)
}

func (c *Container) Run() {
	parent, writePipe, err := process.NewParentProcess(c.Root, c.Spec.Process.Env)
	if err != nil {
		logrus.Error("Failed")
	}

	if err := parent.Start(); err != nil {
		logrus.Error(err)
	}

	b, err := json.Marshal(c)
	if err != nil {
		logrus.Error(err)
	}
	writePipe.Write(b)
	writePipe.Close()
}

// Container Status
type Status int

const (
	Created Status = iota
	Running
	Paused
	Stopped
)

func (s Status) String() string {
	switch s {
	case Created:
		return "created"
	case Running:
		return "running"
	case Paused:
		return "paused"
	case Stopped:
		return "stopped"
	default:
		return "unknown"
	}
}
