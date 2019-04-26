package kubectl

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_ConfigPath_Returns_Error_When_Home_Empty(t *testing.T) {
	_, err := ConfigPath("")
	assert.NotNil(t, err)
}

func Test_Config_Returns_Error_When_File_Does_Not_Exist(t *testing.T) {
	_, err := Config("/some/nonexistent/file/path")
	assert.NotNil(t, err)
}
