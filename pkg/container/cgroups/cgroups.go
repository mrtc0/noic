package cgroups

import (
	"fmt"

	cgroupsv1 "github.com/containerd/cgroups"
	cgroupsv2 "github.com/containerd/cgroups/v2"
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

type Manager struct {
	v1 cgroupsv1.Cgroup
	v2 *cgroupsv2.Manager
}

func IsVersion2() bool {
	return cgroupsv1.Mode() == cgroupsv1.Unified
}

func New(name string, resources specs.LinuxResources) (*Manager, error) {
	if IsVersion2() {
		m, err := cgroupsv2.NewSystemd("/", fmt.Sprintf("%s-cgroup.slice", name), -1, cgroupsv2.ToResources(&resources))
		if err != nil {
			return nil, err
		}

		return &Manager{v2: m}, nil
	}
	control, err := cgroupsv1.New(cgroupsv1.Systemd, cgroupsv1.Slice("user.slice", name), &resources)
	if err != nil {
		return nil, err
	}

	return &Manager{v1: control}, nil
}

func (m Manager) Add(pid uint64) error {
	if IsVersion2() {
		return m.v2.AddProc(pid)
	}

	return m.v1.Add(cgroupsv1.Process{Pid: int(pid)})
}
