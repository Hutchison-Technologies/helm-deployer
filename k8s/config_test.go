package k8s

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_ConfigPath_Panics_When_Home_Empty(t *testing.T) {
	assert.Panics(t, func() { ConfigPath("") })
}

func Test_Config_Panics_When_File_Does_Not_Exist(t *testing.T) {
	assert.Panics(t, func() { Config("/some/nonexistent/file/path") })
}
