package charts

import (
	"github.com/Hutchison-Technologies/helm-deployer/filesystem"
	"github.com/Hutchison-Technologies/helm-deployer/gosexy/yaml"
)

func HasDependency(chartYamlPath, depName, depAlias string) bool {
	if !filesystem.IsFile(chartYamlPath) {
		return false
	}

	chart, err := yaml.Open(chartYamlPath)
	if err != nil {
		return false
	}

	deps := chart.Get("dependencies").([]interface{})
	for _, dep := range deps {
		asDep, ok := dep.(map[interface{}]interface{})
		if !ok {
			continue
		}
		name, nameOk := asDep["name"]
		if depAlias != "" {
			alias, aliasOk := asDep["alias"]
			if nameOk && aliasOk && name == depName && alias == depAlias {
				return true
			}
		} else {
			if nameOk && name == depName {
				return true
			}
		}
	}
	return false
}
