package charts

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const testDataPath = "../testdata/prod.yaml"

func Test_LoadValuesYaml_Returns_Error_When_File_Does_Not_Exist(t *testing.T) {
	_, err := LoadValuesYaml("/some/nonexistent/path/to/a/file.yaml")
	assert.NotNil(t, err)
}

func Test_LoadValuesYaml_Returns_Nil_Error_When_File_Exists(t *testing.T) {
	_, err := LoadValuesYaml(testDataPath)
	assert.Nil(t, err)
}

func Test_LoadValuesYaml_Returns_FileContents(t *testing.T) {
	fileContents, _ := LoadValuesYaml(testDataPath)
	assert.NotNil(t, fileContents)
}
