package deployment

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_BlueGreenDeploymentName_Returns_ValidAppName(t *testing.T) {
	assert.True(t, IsValidAppName(BlueGreenDeploymentName("prod", "blue", "some-api")))
}

func Test_BlueGreenDeploymentName_Returns_Name_PrefixedWith_TargetEnv(t *testing.T) {
	targetEnv := "prod"
	assert.Regexp(t, regexp.MustCompile(fmt.Sprintf("^%s-.*", targetEnv)), BlueGreenDeploymentName(targetEnv, "blue", "some-api"))
}

func Test_BlueGreenDeploymentName_Returns_Name_Containing_AppName(t *testing.T) {
	appName := "some-api"
	assert.Regexp(t, regexp.MustCompile(fmt.Sprintf(".*-%s.*", appName)), BlueGreenDeploymentName("prod", "blue", appName))
}

func Test_BlueGreenDeploymentName_Returns_Name_Containing_Colour(t *testing.T) {
	colour := "green"
	assert.Regexp(t, regexp.MustCompile(fmt.Sprintf(".*-%s.*", colour)), BlueGreenDeploymentName("prod", colour, "some-api"))
}

func Test_StandardChartDeploymentName_Returns_ValidAppName(t *testing.T) {
	assert.True(t, IsValidAppName(StandardChartDeploymentName("prod", "some-api")))
}

func Test_StandardChartDeploymentName_Returns_Name_PrefixedWith_TargetEnv(t *testing.T) {
	targetEnv := "prod"
	assert.Regexp(t, regexp.MustCompile(fmt.Sprintf("^%s-.*", targetEnv)), StandardChartDeploymentName(targetEnv, "some-api"))
}

func Test_StandardChartDeploymentName_Returns_Name_Affixed_With_AppName(t *testing.T) {
	appName := "some-api"
	assert.Regexp(t, regexp.MustCompile(fmt.Sprintf(".*-%s$", appName)), StandardChartDeploymentName("prod", appName))
}

func Test_ServiceReleaseName_Returns_ValidAppName(t *testing.T) {
	assert.True(t, IsValidAppName(ServiceReleaseName("prod", "some-api")))
}

func Test_ServiceReleaseName_Returns_Name_PrefixedWith_TargetEnv(t *testing.T) {
	targetEnv := "prod"
	assert.Regexp(t, regexp.MustCompile(fmt.Sprintf("^%s-.*", targetEnv)), ServiceReleaseName(targetEnv, "some-api"))
}

func Test_ServiceReleaseName_Returns_Name_Containing_AppName(t *testing.T) {
	appName := "some-api"
	assert.Regexp(t, regexp.MustCompile(fmt.Sprintf(".*-%s.*", appName)), ServiceReleaseName("prod", appName))
}

func Test_ServiceReleaseName_Returns_Name_Containing_Service(t *testing.T) {
	assert.Regexp(t, regexp.MustCompile(".*-service-.*"), ServiceReleaseName("prod", "some-api"))
}

func Test_OfflineServiceName_Returns_Valid_AppName(t *testing.T) {
	assert.True(t, IsValidAppName(OfflineServiceName("prod", "some-api")))
}

func Test_OfflineServiceName_Returns_Name_Prefixed_With_TargetEnv(t *testing.T) {
	targetEnv := "prod"
	assert.Regexp(t, regexp.MustCompile(fmt.Sprintf("^%s-.*", targetEnv)), OfflineServiceName(targetEnv, "some-api"))
}

func Test_OfflineServiceName_Returns_Name_Containing_AppName(t *testing.T) {
	appName := "some-api"
	assert.Regexp(t, regexp.MustCompile(fmt.Sprintf(".*-%s-.*", appName)), OfflineServiceName("prod", appName))
}

func Test_OfflineServiceName_Returns_Name_Affixed_With_Offline(t *testing.T) {
	assert.Regexp(t, regexp.MustCompile(".*-offline$"), OfflineServiceName("prod", "some-api"))
}
