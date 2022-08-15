package container

import (
	"github.com/sirupsen/logrus"
)

type Container struct {
	rootfs string
	state  Status
}

func (c *Container) Run() {
	logrus.Debug("Run")
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
