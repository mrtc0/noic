package container

import (
	"encoding/json"

	"github.com/mrtc0/noic/pkg/process"
	"github.com/mrtc0/noic/pkg/specs"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

type Container struct {
	ID    string
	Root  string
	State Status
	Spec  *specs.Spec
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

	parent.Wait()
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
