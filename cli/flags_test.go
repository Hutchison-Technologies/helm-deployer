package cli

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_HandleParseFlags_Returns_Error_When_Flags_Empty(t *testing.T) {
	_, err := HandleParseFlags(make(map[string]string), nil)
	assert.NotNil(t, err)
}

func Test_HandleParseFlags_Returns_Error_When_Flags_Nil(t *testing.T) {
	_, err := HandleParseFlags(nil, nil)
	assert.NotNil(t, err)
}

func Test_HandleParseFlags_Returns_Error_When_Error_Not_Nil(t *testing.T) {
	_, err := HandleParseFlags(nil, errors.New("Some poop happened"))
	assert.NotNil(t, err)
}

func Test_HandleParseFlags_Returns_Given_Flags(t *testing.T) {
	someFlags := map[string]string{
		"thing1": "thing2",
		"thing3": "thing4",
		"thing5": "thing6",
	}
	returnedFlags, err := HandleParseFlags(someFlags, nil)
	assert.Equal(t, someFlags, returnedFlags)
	assert.Nil(t, err)
}
