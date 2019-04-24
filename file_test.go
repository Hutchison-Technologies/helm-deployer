package main

import (
	"github.com/stretchr/testify/assert"
	"runtime"
	"testing"
)

func Test_FileExists_Returns_False_When_File_Does_Not_Exist(t *testing.T) {
	assert.False(t, FileExists("/some/nonexistent/file"))
}
func Test_FileExists_Returns_True_When_File_Does_Exist(t *testing.T) {
	_, filename, _, _ := runtime.Caller(1)
	assert.True(t, FileExists(filename))
}

func Test_IsDirectory_Returns_False_When_Directory_Does_Not_Exist(t *testing.T) {
	assert.False(t, IsDirectory("/some/nonexistent/dir"))
}
func Test_IsDirectory_Returns_False_When_Path_Is_File(t *testing.T) {
	_, filename, _, _ := runtime.Caller(1)
	assert.False(t, IsDirectory(filename))
}
func Test_IsDirectory_Returns_True_When_Directory_Exists(t *testing.T) {
	assert.True(t, IsDirectory("/etc"))
}
