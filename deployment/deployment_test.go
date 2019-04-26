package deployment

import (
	"github.com/stretchr/testify/assert"
	"k8s.io/helm/pkg/proto/hapi/release"
	storageerrors "k8s.io/helm/pkg/storage/errors"
	"testing"
)

func Test_DetermineReleaseCourse_Returns_INSTALL_When_Error_Contains_Not_Found_Error(t *testing.T) {
	releaseName := "best-api"
	assert.Equal(t, ReleaseCourse.INSTALL, DetermineReleaseCourse(releaseName, release.Status_DEPLOYED, storageerrors.ErrReleaseNotFound(releaseName)))
}

func Test_DetermineReleaseCourse_Returns_UPGRADE_WITH_DIFF_CHECK_When_Error_Is_Nil_And_Status_Code_Is_Not_DELETED(t *testing.T) {
	assert.Equal(t, ReleaseCourse.UPGRADE_WITH_DIFF_CHECK, DetermineReleaseCourse("best-api", release.Status_DEPLOYED, nil))
	assert.Equal(t, ReleaseCourse.UPGRADE_WITH_DIFF_CHECK, DetermineReleaseCourse("best-api", release.Status_DELETING, nil))
	assert.Equal(t, ReleaseCourse.UPGRADE_WITH_DIFF_CHECK, DetermineReleaseCourse("best-api", release.Status_FAILED, nil))
	assert.Equal(t, ReleaseCourse.UPGRADE_WITH_DIFF_CHECK, DetermineReleaseCourse("best-api", release.Status_UNKNOWN, nil))
	assert.Equal(t, ReleaseCourse.UPGRADE_WITH_DIFF_CHECK, DetermineReleaseCourse("best-api", release.Status_SUPERSEDED, nil))
	assert.Equal(t, ReleaseCourse.UPGRADE_WITH_DIFF_CHECK, DetermineReleaseCourse("best-api", release.Status_PENDING_INSTALL, nil))
	assert.Equal(t, ReleaseCourse.UPGRADE_WITH_DIFF_CHECK, DetermineReleaseCourse("best-api", release.Status_PENDING_ROLLBACK, nil))
	assert.Equal(t, ReleaseCourse.UPGRADE_WITH_DIFF_CHECK, DetermineReleaseCourse("best-api", release.Status_PENDING_UPGRADE, nil))
}

func Test_DetermineReleaseCourse_Returns_UPGRADE_When_Error_Is_Nil_And_Status_Code_Is_DELETED(t *testing.T) {
	assert.Equal(t, ReleaseCourse.UPGRADE, DetermineReleaseCourse("best-api", release.Status_DELETED, nil))
}
