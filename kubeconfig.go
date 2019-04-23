package main

import (
	"fmt"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"path/filepath"
)

func KubeConfigPath(homeDir string) string {
	if homeDir == "" {
		panic("home dir must not be empty")
	}
	return filepath.Join(homeDir, ".kube", "config")
}

func KubeConfig(configPath string) *rest.Config {
	if !FileExists(configPath) {
		panic(fmt.Sprintf("kubeconfig does not exist: %s", configPath))
	}

	config, err := clientcmd.BuildConfigFromFlags("", configPath)
	if err != nil {
		panic(err.Error())
	}
	return config
}
