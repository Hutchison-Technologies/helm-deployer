package main

import (
	"github.com/stretchr/testify/assert"
	"runtime"
	"testing"
)

func TestFileExistsReturnsFalseWhenFileDoesNotExist(t *testing.T) {
	assert.False(t, FileExists("/some/nonexistent/file"))
}
func TestFileExistsReturnsTrueWhenFileDoesExist(t *testing.T) {
	_, filename, _, _ := runtime.Caller(1)
	assert.True(t, FileExists(filename))
}
