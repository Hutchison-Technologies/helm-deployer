package charts

import (
	"fmt"
)

func RequirementsYamlPath(chartDir string) string {
	return fmt.Sprintf("%s/requirements.yaml", chartDir)
}
