package cli

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_HandleParseFlags_Panics_When_Flags_Empty(t *testing.T) {
	assert.Panics(t, func() { HandleParseFlags(make(map[string]string), nil) })
}

func Test_HandleParseFlags_Panics_When_Flags_Nil(t *testing.T) {
	assert.Panics(t, func() { HandleParseFlags(nil, nil) })
}

func Test_HandleParseFlags_Panics_When_Error_Not_Nil(t *testing.T) {
	assert.Panics(t, func() { HandleParseFlags(nil, errors.New("Some poop happened")) })
}

func Test_HandleParseFlags_Returns_Given_Flags(t *testing.T) {
	someFlags := map[string]string{
		"thing1": "thing2",
		"thing3": "thing4",
		"thing5": "thing6",
	}
	assert.Equal(t, someFlags, HandleParseFlags(someFlags, nil))
}
