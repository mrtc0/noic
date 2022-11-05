package cgroups

import (
	"fmt"
	"os"
	"strings"
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

type CgroupConfig struct {
	UseSystemd bool
	CgroupPath string
	Resources  *specs.LinuxResources
	Name       string
	Pid        int

	scopePrefix string
	parent      string
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

func New(config *CgroupConfig) (*Manager, error) {
	if config.CgroupPath == "" {
		config.scopePrefix = "noic"
		config.parent = "/"
	} else {
		// e.g. system.slice:docker:123456
		parts := strings.Split(config.CgroupPath, ":")
		if len(parts) != 3 {
			return nil, fmt.Errorf("expect cgroupsPath to be format \"slice:prefix:name\"")
		}

		config.parent = parts[0]
		config.scopePrefix = parts[1]
		config.Name = parts[2]
	}

	r := cgroupsv2.ToResources(config.Resources)
	// Workaround.
	// https://github.com/containerd/cgroups/blob/724eb82fe759f3b3b9c5f07d22d2fab93467dc56/v2/utils.go#L164
	if shares := config.Resources.CPU.Shares; shares != nil {
		convertedWeight := 1 + ((*shares)*9999)/262142
		w := uint64(convertedWeight)
		r.CPU.Weight = &w
	}

	m, err := cgroupsv2.NewSystemd("", getUnitName(config), config.Pid, r)
	if err != nil {
		return nil, err
	}

	return &Manager{v2: m}, nil

	// m, err := cgroupsv2.NewManager(path, fmt.Sprintf("/%s-cgroup.slice", name), cgroupsv2.ToResources(&resources))
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

func getUnitName(config *CgroupConfig) string {
	if !strings.HasSuffix(config.Name, ".slice") {
		return config.scopePrefix + "-" + config.Name + ".scope"
	}

	return config.Name
}
