package kubectl

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/Hutchison-Technologies/helm-deployer/filesystem"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func ConfigPath(homeDir string) (string, error) {
	if homeDir == "" {
		return "", errors.New("Home dir must not be empty")
	}
	return filepath.Join(homeDir, ".kube", "config"), nil
}

func Config(configPath string) (*rest.Config, error) {
	if !filesystem.IsFile(configPath) {
		return nil, errors.New(fmt.Sprintf("kubeconfig does not exist at path: %s", configPath))
	}

	config, err := clientcmd.BuildConfigFromFlags("", configPath)
	if err != nil {
		return nil, err
	}
	return config, nil
}
