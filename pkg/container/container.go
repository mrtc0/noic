package container

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mrtc0/noic/pkg/process"
	"github.com/mrtc0/noic/pkg/specs"
	gopsutil "github.com/shirou/gopsutil/process"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

const execFifoFilename = "exec.fifo"

type Container struct {
	ID           string
	Root         string
	ExecFifoPath string
	Spec         *specs.Spec
	InitProcess  *process.InitProcess
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

	c.InitProcess = &process.InitProcess{Pid: parent.Process.Pid}

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

// CurrentStatus is return current container status
func (c *Container) CurrentStatus() Status {
	if c.InitProcess == nil {
		return Stopped
	}

	ps, err := gopsutil.NewProcess(int32(c.InitProcess.Pid))
	if err != nil {
		return Stopped
	}

	stat, err := ps.Status()
	if err != nil {
		return Stopped
	}

	if stat == "Z" {
		return Stopped
	}

	if _, err := os.Stat(filepath.Join(StateDir, c.ID, execFifoFilename)); err == nil {
		return Created
	}

	return Running
}
