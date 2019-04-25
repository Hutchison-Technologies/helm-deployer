package k8s

import (
	"fmt"
	"github.com/Hutchison-Technologies/bluegreen-deployer/filesystem"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"path/filepath"
)

func ConfigPath(homeDir string) string {
	if homeDir == "" {
		panic("home dir must not be empty")
	}
	return filepath.Join(homeDir, ".kube", "config")
}

func Config(configPath string) *rest.Config {
	if !filesystem.IsFile(configPath) {
		panic(fmt.Sprintf("kubeconfig does not exist: %s", configPath))
	}

	config, err := clientcmd.BuildConfigFromFlags("", configPath)
	if err != nil {
		panic(err.Error())
	}
	return config
}
