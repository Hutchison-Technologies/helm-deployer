package charts

import (
	"fmt"
)

func ChartYamlPath(chartDir string) string {
	return fmt.Sprintf("%s/Chart.yaml", chartDir)
}
