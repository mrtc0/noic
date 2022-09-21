package container

import (
	"github.com/mrtc0/noic/pkg/process"
	"github.com/sirupsen/logrus"
)

type Container struct {
	rootfs string
	state  Status
}

func (c *Container) Run() {
	logrus.Info("Run")
	// s := specs.SetupSpec()

	cmd, writePipe, err := process.NewParentProcess()
	if err != nil {
		logrus.Error("Failed")
	}

	if err := cmd.Start(); err != nil {
		logrus.Error(err)
	}

	writePipe.WriteString("ps aux")
	writePipe.Close()

	cmd.Wait()
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
