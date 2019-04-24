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

func TestIsDirectoryReturnsFalseWhenDirectoryDoesNotExist(t *testing.T) {
	assert.False(t, IsDirectory("/some/nonexistent/dir"))
}
func TestIsDirectoryReturnsFalseWhenPathIsFile(t *testing.T) {
	_, filename, _, _ := runtime.Caller(1)
	assert.False(t, IsDirectory(filename))
}
func TestIsDirectoryReturnsTrueWhenDirectoryExists(t *testing.T) {
	assert.True(t, IsDirectory("/etc"))
}
