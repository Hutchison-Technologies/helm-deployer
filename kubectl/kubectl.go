package kubectl

import (
	"fmt"
	"os"

	"k8s.io/client-go/kubernetes"
	appsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	autoscalingv1 "k8s.io/client-go/kubernetes/typed/autoscaling/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
)

//Client is used for core kube actions
func Client() (corev1.CoreV1Interface, error) {
	_, client, err := getKubeClient()
	if err != nil {
		return nil, err
	}
	return client.CoreV1(), nil
}

//AppsClient is used for interacting with deployments
func AppsClient() (appsv1.AppsV1Interface, error) {
	_, client, err := getKubeClient()
	if err != nil {
		return nil, err
	}
	return client.AppsV1(), nil
}

//HPAClient is used for interacting with horizontal pod autoscalers
func HPAClient() (autoscalingv1.AutoscalingV1Interface, error) {
	_, client, err := getKubeClient()
	if err != nil {
		return nil, err
	}
	return client.AutoscalingV1(), nil
}

func getKubeClient() (*rest.Config, kubernetes.Interface, error) {
	configPath, err := ConfigPath(os.Getenv("HOME"))
	if err != nil {
		return nil, nil, fmt.Errorf("Error getting kubectl config path: %s", err)
	}

	config, err := Config(configPath)
	if err != nil {
		return nil, nil, fmt.Errorf("Error getting kubeconfig: %s", err)
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("Error building kubectl client: %s", err)
	}
	return config, client, nil
}
