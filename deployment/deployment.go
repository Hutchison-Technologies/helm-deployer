package deployment

import (
	"fmt"
	"github.com/Hutchison-Technologies/bluegreen-deployer/k8s"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/typed/core/v1"
)

func GetOfflineService(kubeClient v1.CoreV1Interface, targetEnv, appName string) (*corev1.Service, error) {
	offlineServiceName := OfflineServiceName(targetEnv, appName)
	service, err := k8s.GetService(kubeClient, offlineServiceName)
	if err != nil {
		return nil, fmt.Errorf("Error looking for offline service: %s", err)
	}
	return service, nil
}
