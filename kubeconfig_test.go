package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestKubeConfigPathPanicsWhenHomeEmpty(t *testing.T) {
	assert.Panics(t, func() { KubeConfigPath("") })
}

func TestKubeConfigPanicsWhenFileDoesNotExist(t *testing.T) {
	assert.Panics(t, func() { KubeConfig("/some/nonexistent/file/path") })
}
