package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_IsValidAppName_Returns_False_When_Given_Invalid_AppName(t *testing.T) {
	assert.False(t, IsValidAppName(""))
	assert.False(t, IsValidAppName(" "))
	assert.False(t, IsValidAppName("contains -space"))
	assert.False(t, IsValidAppName("contains_underscore"))
	assert.False(t, IsValidAppName("contains-CAPS"))
	assert.False(t, IsValidAppName("contains-non-alpha-numeric-chars-£$%^&*()"))
	assert.False(t, IsValidAppName("contains-more-than-64-characters-because-that-shit-dont-fly-here-son"))
}

func Test_IsValidAppName_Returns_True_When_Given_Valid_AppName(t *testing.T) {
	assert.True(t, IsValidAppName("contains-no-space"))
	assert.True(t, IsValidAppName("containsnumb3r"))
	assert.True(t, IsValidAppName("containslowercase"))
}

func Test_IsValidAppVersion_Returns_False_When_Given_Invalid_AppVersion(t *testing.T) {
	assert.False(t, IsValidAppVersion(""))
	assert.False(t, IsValidAppVersion(" "))
	assert.False(t, IsValidAppVersion("letters"))
	assert.False(t, IsValidAppVersion("0"))
	assert.False(t, IsValidAppVersion("0.1"))
	assert.False(t, IsValidAppVersion("0.1.a"))
	assert.False(t, IsValidAppVersion("0.1.!"))
	assert.False(t, IsValidAppVersion("latest"))
	assert.False(t, IsValidAppVersion("master"))
	assert.False(t, IsValidAppVersion("v0.1. 2"))
}

func Test_IsValidAppVersion_Returns_True_When_Given_Valid_AppVersion(t *testing.T) {
	assert.True(t, IsValidAppVersion("v0.0.0"))
	assert.True(t, IsValidAppVersion("v1.0.0"))
	assert.True(t, IsValidAppVersion("1.0.0"))
	assert.True(t, IsValidAppVersion("1.0.10"))
}

func Test_IsValidTargetEnv_Returns_False_When_Given_Invalid_TargetEnv(t *testing.T) {
	assert.False(t, IsValidTargetEnv(""))
	assert.False(t, IsValidTargetEnv(" "))
	assert.False(t, IsValidTargetEnv("green"))
	assert.False(t, IsValidTargetEnv("124"))
	assert.False(t, IsValidTargetEnv("test"))
	assert.False(t, IsValidTargetEnv("PROD"))
	assert.False(t, IsValidTargetEnv("STAGING"))
	assert.False(t, IsValidTargetEnv("production"))
	assert.False(t, IsValidTargetEnv("stage"))
}

func Test_IsValidTargetEnv_Returns_True_When_Given_Valid_TargetEnv(t *testing.T) {
	assert.True(t, IsValidTargetEnv("staging"))
	assert.True(t, IsValidTargetEnv("prod"))
}
