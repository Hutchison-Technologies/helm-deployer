package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_KubeConfigPath_Panics_When_Home_Empty(t *testing.T) {
	assert.Panics(t, func() { KubeConfigPath("") })
}

func Test_KubeConfig_Panics_When_File_Does_Not_Exist(t *testing.T) {
	assert.Panics(t, func() { KubeConfig("/some/nonexistent/file/path") })
}
