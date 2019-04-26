package k8s

import (
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"testing"
)

func Test_ServiceSelectorColour_Returns_Empty_String_When_Given_Nil_Service(t *testing.T) {
	assert.Equal(t, "", ServiceSelectorColour(nil))
}

func Test_ServiceSelectorColour_Returns_Empty_String_When_Given_Service_Has_No_Colour(t *testing.T) {
	assert.Equal(t, "", ServiceSelectorColour(&corev1.Service{
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"thing": "thing",
			},
		},
	}))
}

func Test_ServiceSelectorColour_Returns_Empty_String_When_Given_Service_Has_No_Selectors(t *testing.T) {
	assert.Equal(t, "", ServiceSelectorColour(&corev1.Service{
		Spec: corev1.ServiceSpec{},
	}))
}

func Test_ServiceSelectorColour_Returns_Empty_String_When_Given_Service_Has_No_Spec(t *testing.T) {
	assert.Equal(t, "", ServiceSelectorColour(&corev1.Service{}))
}

func Test_ServiceSelectorColour_Returns_Green_When_Given_Service_Has_Green_Selector(t *testing.T) {
	selectorColour := "green"
	assert.Equal(t, selectorColour, ServiceSelectorColour(&corev1.Service{
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"colour": selectorColour,
			},
		},
	}))
}
