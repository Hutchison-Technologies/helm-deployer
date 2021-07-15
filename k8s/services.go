package k8s

import (
	"context"
	"errors"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/typed/core/v1"
)

func ServiceSelectorColour(service *corev1.Service) string {
	if service != nil && service.Spec.Selector != nil && len(service.Spec.Selector) > 0 {
		if colour, ok := service.Spec.Selector["colour"]; ok {
			return colour
		}
	}
	return ""
}

func GetService(kubeClient v1.CoreV1Interface, serviceName string) (*corev1.Service, error) {
	service, err := kubeClient.Services("default").Get(context.TODO(), serviceName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error getting service \033[32m%s\033[97m, %s", serviceName, err.Error()))
	}
	return service, nil
}
