package specs

import (
	"encoding/json"
	"fmt"
	"os"
)

const DefaultSpecConfigFilename = "config.json"

func LoadSpec(specConfigPath string) (*Spec, error) {
	var spec *Spec
	specConf, err := os.Open(specConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("spec file %s not found", specConfigPath)
		}

		return nil, err
	}

	defer specConf.Close()

	if err = json.NewDecoder(specConf).Decode(&spec); err != nil {
		return nil, fmt.Errorf("failed decode spec: %w", err)
	}

	return spec, nil
}
