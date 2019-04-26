package deployment

import (
	"fmt"
)

func ChartValuesPath(chartDir, targetEnv string) string {
	return fmt.Sprintf("%s/%s.yaml", chartDir, targetEnv)
}
