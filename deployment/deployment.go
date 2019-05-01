package deployment

import (
	"fmt"
	"strings"

	"github.com/Hutchison-Technologies/helm-deployer/k8s"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/helm/pkg/proto/hapi/release"
	storageerrors "k8s.io/helm/pkg/storage/errors"
)

type alias = int

type list struct {
	INSTALL                 alias
	UPGRADE_WITH_DIFF_CHECK alias
	UPGRADE                 alias
}

var ReleaseCourse = &list{
	INSTALL:                 0,
	UPGRADE_WITH_DIFF_CHECK: 1,
	UPGRADE:                 2,
}

func GetOfflineService(kubeClient v1.CoreV1Interface, targetEnv, appName string) (*corev1.Service, error) {
	offlineServiceName := OfflineServiceName(targetEnv, appName)
	service, err := k8s.GetService(kubeClient, offlineServiceName)
	if err != nil {
		return nil, fmt.Errorf("Error looking for offline service: %s", err)
	}
	return service, nil
}

func DetermineReleaseCourse(releaseName string, statusCode release.Status_Code, err error) int {
	if err != nil && strings.Contains(err.Error(), storageerrors.ErrReleaseNotFound(releaseName).Error()) {
		return ReleaseCourse.INSTALL
	} else if statusCode == release.Status_DELETED {
		return ReleaseCourse.UPGRADE
	} else {
		return ReleaseCourse.UPGRADE_WITH_DIFF_CHECK
	}
}

func ChartValuesForDeployment(deployColour, appVersion string) [][]interface{} {
	return [][]interface{}{
		[]interface{}{"bluegreen", "is_service_release", false},
		[]interface{}{"bluegreen", "deployment", "colour", deployColour},
		[]interface{}{"bluegreen", "deployment", "version", appVersion},
	}
}

func ChartValuesForServiceRelease(deployColour string) [][]interface{} {
	return [][]interface{}{
		[]interface{}{"bluegreen", "is_service_release", true},
		[]interface{}{"bluegreen", "service", "selector", "colour", deployColour},
	}
}
