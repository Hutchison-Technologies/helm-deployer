package main

import (
	"regexp"
)

func IsValidAppName(appName string) bool {
	return len(appName) < 64 && regexp.MustCompile(`^[a-z][a-z|0-9|-]+$`).MatchString(appName)
}

func IsValidAppVersion(appName string) bool {
	return regexp.MustCompile(`^v?[0-9]+.[0-9]+.[0-9]+$`).MatchString(appName)
}

func IsValidTargetEnv(targetEnv string) bool {
	return targetEnv == "prod" || targetEnv == "staging"
}
