package cgroups

import (
	"fmt"
	"os"
	"syscall"

	cgroupsv1 "github.com/containerd/cgroups"
	cgroupsv2 "github.com/containerd/cgroups/v2"
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

const defaultMountFlags = syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV

type Manager struct {
	v1 cgroupsv1.Cgroup
	v2 *cgroupsv2.Manager
}

func mountCgroupV1(mountPath string) error {
	if err := os.MkdirAll(mountPath, 0o755); err != nil {
		return err
	}

	flags := defaultMountFlags | syscall.MS_RDONLY
	return syscall.Mount("cgroup", mountPath, "cgroup", uintptr(flags), "")
}

func mountCgroupV2(mountPath string) error {
	if err := os.MkdirAll(mountPath, 0o755); err != nil {
		return err
	}

	flags := defaultMountFlags
	return syscall.Mount("cgroup2", mountPath, "cgroup2", uintptr(flags), "")
}

func IsVersion2() bool {
	return cgroupsv1.Mode() == cgroupsv1.Unified
}

func New(name string, path string, resources specs.LinuxResources) (*Manager, error) {
	// if IsVersion2() {
	if path == "" {
		if err := mountCgroupV2("/sys/fs/cgroup"); err != nil {
			return nil, err
		}
		m, err := cgroupsv2.NewSystemd("/", fmt.Sprintf("%s-cgroup.slice", name), -1, cgroupsv2.ToResources(&resources))
		if err != nil {
			return nil, err
		}

		return &Manager{v2: m}, nil
	}

	if err := mountCgroupV2(path); err != nil {
		return nil, err
	}
	m, err := cgroupsv2.NewManager(path, fmt.Sprintf("/%s-cgroup.slice", name), cgroupsv2.ToResources(&resources))
	if err != nil {
		return nil, err
	}

	return &Manager{v2: m}, nil
	/*
		}
		control, err := cgroupsv1.New(cgroupsv1.V1, cgroupsv1.StaticPath(name), &resources)
		if err != nil {
			return nil, err
		}

		return &Manager{v1: control}, nil
	*/
}

func (m Manager) Add(pid uint64) error {
	if IsVersion2() {
		return m.v2.AddProc(pid)
	}

	return m.v1.Add(cgroupsv1.Process{Pid: int(pid)})
}
