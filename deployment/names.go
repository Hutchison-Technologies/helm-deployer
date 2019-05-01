package deployment

import (
	"fmt"
)

func BlueGreenDeploymentName(targetEnv, colour, appName string) string {
	return fmt.Sprintf("%s-%s-%s", targetEnv, colour, appName)
}

func StandardChartDeploymentName(targetEnv, appName string) string {
	return fmt.Sprintf("%s-%s", targetEnv, appName)
}

func OfflineServiceName(targetEnv, appName string) string {
	return fmt.Sprintf("%s-%s-offline", targetEnv, appName)
}

func ServiceReleaseName(targetEnv, appName string) string {
	return fmt.Sprintf("%s-service-%s", targetEnv, appName)
}
