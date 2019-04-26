package deployment

import (
	"fmt"
)

func DeploymentName(targetEnv, colour, appName string) string {
	return fmt.Sprintf("%s-%s-%s", targetEnv, colour, appName)
}

func OfflineServiceName(targetEnv, appName string) string {
	return fmt.Sprintf("%s-%s-offline", targetEnv, appName)
}
