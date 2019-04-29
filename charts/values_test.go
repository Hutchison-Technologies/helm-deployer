package charts

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const TEST_VALUES_PATH = "../testdata/prod.yaml"

func Test_LoadValuesYaml_Returns_Error_When_File_Does_Not_Exist(t *testing.T) {
	_, err := LoadValuesYaml("/some/nonexistent/path/to/a/file.yaml")
	assert.NotNil(t, err)
}

func Test_LoadValuesYaml_Returns_Nil_Error_When_File_Exists(t *testing.T) {
	_, err := LoadValuesYaml(TEST_VALUES_PATH)
	assert.Nil(t, err)
}

func Test_LoadValuesYaml_Returns_FileContents(t *testing.T) {
	fileContents, _ := LoadValuesYaml(TEST_VALUES_PATH)
	assert.NotNil(t, fileContents)
}
