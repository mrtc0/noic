package container

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

const StateFilename = "state.json"
const StateDir = "/var/run/noic"

func StateFilePath(id string) (string, error) {
	d, err := StateDirPath(id)
	if err != nil {
		return "", err
	}

	path := filepath.Join(d, StateFilename)

	return path, nil
}

func StateDirPath(id string) (string, error) {
	path := filepath.Join(StateDir, id)
	if err := os.MkdirAll(path, 0777); err != nil {
		return path, err
	}

	return path, nil
}

func SaveStateFile(container *Container) error {
	path, err := StateFilePath(container.ID)
	if err != nil {
		return err
	}

	c, err := json.Marshal(container)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path, c, 0644)
	if err != nil {
		return err
	}

	return nil
}
