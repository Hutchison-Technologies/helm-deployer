package charts

import (
	"errors"
	"fmt"
	"github.com/Hutchison-Technologies/bluegreen-deployer/filesystem"
	"github.com/Hutchison-Technologies/bluegreen-deployer/gosexy/yaml"
)

func LoadValuesYaml(path string) (*yaml.Yaml, error) {
	if !filesystem.IsFile(path) {
		return nil, errors.New(fmt.Sprintf("Expected to find chart values yaml at: \033[31m%s\033[97m, but found nothing.", path))
	}

	values, err := yaml.Open(path)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Could not open chart values at \033[31m%s\033[97m, %s", path, err.Error()))
	}

	return values, nil
}

func EditValuesYaml(valuesYaml *yaml.Yaml, settings [][]interface{}) ([]byte, error) {
	for _, setting := range settings {
		err := valuesYaml.Set(setting...)
		if err != nil {
			return nil, err
		}
	}
	err := valuesYaml.Save()
	if err != nil {
		return nil, err
	}
	values, err := valuesYaml.ToByteArray()
	if err != nil {
		return nil, err
	}
	return values, nil
}
