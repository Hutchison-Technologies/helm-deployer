package charts

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const TEST_CHART_PATH = "../testdata/Chart.yaml"

func Test_HasDependency_Returns_False_When_File_Does_Not_Exist(t *testing.T) {
	result := HasDependency("/some/nonexistent/path/to/a/file.yaml", "some-name", "some-alias")
	assert.False(t, result)
}

func Test_HasDependency_Returns_False_When_Dependency_Not_Present(t *testing.T) {
	result := HasDependency(TEST_CHART_PATH, "made-up", "not-good")
	assert.False(t, result)
}

func Test_HasDependency_Returns_False_When_Dependency_Is_Present_But_Alias_Is_Not_Present(t *testing.T) {
	result := HasDependency(TEST_CHART_PATH, "unaliased-dep", "gone-fishing")
	assert.False(t, result)
}

func Test_HasDependency_Returns_False_When_Dependency_Is_Present_And_Alias_Is_Present_But_Alias_Does_Not_Match(t *testing.T) {
	result := HasDependency(TEST_CHART_PATH, "aliased-dep", "garbage")
	assert.False(t, result)
}

func Test_HasDependency_Returns_True_When_Dependency_Is_Present_And_Alias_Is_Present_And_Alias_Matches(t *testing.T) {
	result := HasDependency(TEST_CHART_PATH, "aliased-dep", "aliaseddep")
	assert.True(t, result)
}
