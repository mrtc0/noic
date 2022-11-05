package specs

import (
	"encoding/json"
	"fmt"
	"os"

	specsgo "github.com/opencontainers/runtime-spec/specs-go"
)

const DefaultSpecConfigFilename = "config.json"

func LoadSpec(specConfigPath string) (*specsgo.Spec, error) {
	var spec *specsgo.Spec
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
