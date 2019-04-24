package main

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"regexp"
	"testing"
)

func TestOfflineServiceNameReturnsValidAppName(t *testing.T) {
	assert.True(t, IsValidAppName(OfflineServiceName("prod", "some-api")))
}

func TestOfflineServiceNameReturnsNamePrefixedWithTargetEnv(t *testing.T) {
	targetEnv := "prod"
	assert.Regexp(t, regexp.MustCompile(fmt.Sprintf("^%s-.*", targetEnv)), OfflineServiceName(targetEnv, "some-api"))
}

func TestOfflineServiceNameReturnsNameContainingAppName(t *testing.T) {
	appName := "some-api"
	assert.Regexp(t, regexp.MustCompile(fmt.Sprintf(".*-%s-.*", appName)), OfflineServiceName("prod", appName))
}

func TestOfflineServiceNameReturnsNameAffixedWithOffline(t *testing.T) {
	assert.Regexp(t, regexp.MustCompile(".*-offline$"), OfflineServiceName("prod", "some-api"))
}

func TestServiceSelectorColourReturnsBlueWhenGivenNilService(t *testing.T) {
	assert.Equal(t, "blue", ServiceSelectorColour(nil, nil))
}

func TestServiceSelectorColourReturnsBlueWhenGivenError(t *testing.T) {
	assert.Equal(t, "blue", ServiceSelectorColour(&corev1.Service{
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"thing": "thing",
			},
		},
	}, errors.New("something pooped")))
}

func TestServiceSelectorColourReturnsBlueWhenGivenServiceHasNoColour(t *testing.T) {
	assert.Equal(t, "blue", ServiceSelectorColour(&corev1.Service{
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"thing": "thing",
			},
		},
	}, nil))
}

func TestServiceSelectorColourReturnsBlueWhenGivenServiceHasNoSelectors(t *testing.T) {
	assert.Equal(t, "blue", ServiceSelectorColour(&corev1.Service{
		Spec: corev1.ServiceSpec{},
	}, nil))
}

func TestServiceSelectorColourReturnsBlueWhenGivenServiceHasNoSpec(t *testing.T) {
	assert.Equal(t, "blue", ServiceSelectorColour(&corev1.Service{}, nil))
}

func TestServiceSelectorColourReturnsGreenWhenGivenServiceHasGreenSelector(t *testing.T) {
	selectorColour := "green"
	assert.Equal(t, selectorColour, ServiceSelectorColour(&corev1.Service{
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"colour": selectorColour,
			},
		},
	}, nil))
}
