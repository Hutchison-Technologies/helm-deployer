package main

import (
	"fmt"
)

func DeploymentName(targetEnv, colour, appName string) string {
	return fmt.Sprintf("%s-%s-%s", targetEnv, colour, appName)
}
