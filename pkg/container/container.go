package container

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/mrtc0/noic/pkg/process"
	specs "github.com/opencontainers/runtime-spec/specs-go"
	gopsutil "github.com/shirou/gopsutil/process"
)

const execFifoFilename = "exec.fifo"

type Container struct {
	ID                 string
	Root               string
	ExecFifoPath       string
	Spec               *specs.Spec
	InitProcess        *process.InitProcess
	State              specs.State
	StateRootDirectory string
	UseSystemdCgroups  bool
}

func Exists(stateRootDirectory, containerID string) bool {
	d := filepath.Join(stateRootDirectory, containerID)
	_, err := os.Stat(d)
	return err == nil
}

func (c *Container) StateFilePath() string {
	return filepath.Join(c.StateRootDirectory, c.ID, "state.json")
}

func (c *Container) StateDirectory() string {
	return filepath.Join(c.StateRootDirectory, c.ID)
}

func (c *Container) SaveState() error {
	if err := os.MkdirAll(c.StateDirectory(), 0o700); err != nil {
		return err
	}

	j, err := json.Marshal(c)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(c.StateFilePath(), j, 0644)
	if err != nil {
		return err
	}

	return nil
}

func FindByID(id string, stateRootDirectory string) (*Container, error) {
	if !Exists(stateRootDirectory, id) {
		return nil, fmt.Errorf("container %s does not exists", id)
	}

	path := filepath.Join(stateRootDirectory, id, "state.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("state.json does not exists: %s", err)
	}

	var container *Container
	if err = json.Unmarshal(raw, &container); err != nil {
		return nil, fmt.Errorf("faild unmarshal: %s", err)
	}

	return container, nil
}

func (c *Container) Run() error {
	parent, writePipe, err := c.NewParentProcess()
	if err != nil {
		return fmt.Errorf("faild NewParentProcess: %s", err)
	}

	if err := parent.Start(); err != nil {
		return fmt.Errorf("failed start parent Process: %s", err)
	}

	c.InitProcess = &process.InitProcess{Pid: parent.Process.Pid}
	c.State.Pid = parent.Process.Pid
	c.State.Status = specs.ContainerState(c.CurrentStatus().String())

	b, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("faild marshal: %s", err)
	}
	writePipe.Write(b)
	writePipe.Close()

	return nil
}

func (c *Container) Destroy() error {
	if err := os.RemoveAll(c.StateDirectory()); err != nil {
		return err
	}

	return nil
}

func (c *Container) Kill() error {
	ps, err := gopsutil.NewProcess(int32(c.InitProcess.Pid))
	if err != nil {
		return err
	}

	return ps.Kill()
}

func (c *Container) CreatePIDFile(path string) error {
	var (
		tmpDir  = filepath.Dir(path)
		tmpName = filepath.Join(tmpDir, "."+filepath.Base(path))
	)
	f, err := os.OpenFile(tmpName, os.O_RDWR|os.O_CREATE|os.O_EXCL|os.O_SYNC, 0o666)
	if err != nil {
		return err
	}
	_, err = f.WriteString(strconv.Itoa(c.InitProcess.Pid))
	f.Close()
	if err != nil {
		return err
	}
	return os.Rename(tmpName, path)
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

	if _, err := os.Stat(filepath.Join(c.StateDirectory(), execFifoFilename)); err == nil {
		return Created
	}

	return Running
}
