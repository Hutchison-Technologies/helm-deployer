package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_DetermineCommand_Returns_UNKNOWN_When_Given_Empty_String(t *testing.T) {
	assert.Equal(t, Command.UNKNOWN, DetermineCommand(""))
}

func Test_DetermineCommand_Returns_UNKNOWN_When_Given_Unrecognised_String(t *testing.T) {
	assert.Equal(t, Command.UNKNOWN, DetermineCommand("this is not a command"))
}

func Test_DetermineCommand_Returns_BLUEGREEN_When_Given_Bluegreen_String(t *testing.T) {
	assert.Equal(t, Command.BLUEGREEN, DetermineCommand("bluegreen"))
}

func Test_DetermineCommand_Returns_STANDARD_When_Given_Standard_String(t *testing.T) {
	assert.Equal(t, Command.STANDARD, DetermineCommand("standard"))
}
