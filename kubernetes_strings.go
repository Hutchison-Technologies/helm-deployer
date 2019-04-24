package main

import (
	"fmt"
)

func OfflineServiceName(targetEnv, appName string) string {
	return fmt.Sprintf("%s-%s-offline", targetEnv, appName)
}
