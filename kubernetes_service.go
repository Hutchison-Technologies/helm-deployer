package main

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
)

func OfflineServiceName(targetEnv, appName string) string {
	return fmt.Sprintf("%s-%s-offline", targetEnv, appName)
}

func ServiceSelectorColour(service *corev1.Service, err error) string {
	defaultColour := "blue"
	if err == nil && service != nil && service.Spec.Selector != nil && len(service.Spec.Selector) > 0 {
		if colour, ok := service.Spec.Selector["colour"]; ok {
			return colour
		}
	}
	return defaultColour
}
