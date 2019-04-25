package deployment

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"regexp"
	"testing"
)

func Test_DeploymentName_Returns_ValidAppName(t *testing.T) {
	assert.True(t, IsValidAppName(DeploymentName("prod", "blue", "some-api")))
}

func Test_DeploymentName_Returns_Name_PrefixedWith_TargetEnv(t *testing.T) {
	targetEnv := "prod"
	assert.Regexp(t, regexp.MustCompile(fmt.Sprintf("^%s-.*", targetEnv)), DeploymentName(targetEnv, "blue", "some-api"))
}

func Test_DeploymentName_Returns_Name_Containing_AppName(t *testing.T) {
	appName := "some-api"
	assert.Regexp(t, regexp.MustCompile(fmt.Sprintf(".*-%s.*", appName)), DeploymentName("prod", "blue", appName))
}

func Test_DeploymentName_Returns_Name_Containing_Colour(t *testing.T) {
	colour := "green"
	assert.Regexp(t, regexp.MustCompile(fmt.Sprintf(".*-%s.*", colour)), DeploymentName("prod", colour, "some-api"))
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

func Test_ServiceSelectorColour_Returns_Blue_When_Given_Nil_Service(t *testing.T) {
	assert.Equal(t, "blue", ServiceSelectorColour(nil, nil))
}

func Test_ServiceSelectorColour_Returns_Blue_When_Given_Error(t *testing.T) {
	assert.Equal(t, "blue", ServiceSelectorColour(&corev1.Service{
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"thing": "thing",
			},
		},
	}, errors.New("something pooped")))
}

func Test_ServiceSelectorColour_Returns_Blue_When_Given_Service_Has_No_Colour(t *testing.T) {
	assert.Equal(t, "blue", ServiceSelectorColour(&corev1.Service{
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"thing": "thing",
			},
		},
	}, nil))
}

func Test_ServiceSelectorColour_Returns_Blue_When_Given_Service_Has_No_Selectors(t *testing.T) {
	assert.Equal(t, "blue", ServiceSelectorColour(&corev1.Service{
		Spec: corev1.ServiceSpec{},
	}, nil))
}

func Test_ServiceSelectorColour_Returns_Blue_When_Given_Service_Has_No_Spec(t *testing.T) {
	assert.Equal(t, "blue", ServiceSelectorColour(&corev1.Service{}, nil))
}

func Test_ServiceSelectorColour_Returns_Green_When_Given_Service_Has_Green_Selector(t *testing.T) {
	selectorColour := "green"
	assert.Equal(t, selectorColour, ServiceSelectorColour(&corev1.Service{
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"colour": selectorColour,
			},
		},
	}, nil))
}
