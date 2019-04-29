package charts

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"regexp"
	"testing"
)

func Test_RequirementsYamlPath_Returns_Path_Prefixed_With_Chart_Dir(t *testing.T) {
	chartDir := "/some/dir"
	assert.Regexp(t, regexp.MustCompile(fmt.Sprintf("^%s/.*", chartDir)), RequirementsYamlPath(chartDir))
}

func Test_RequirementsYamlPath_Returns_Path_Affixed_With_Yaml_Extension(t *testing.T) {
	assert.Regexp(t, regexp.MustCompile("^.*\\.yaml$"), RequirementsYamlPath("/some/dir"))
}

func Test_RequirementsYamlPath_Returns_Path_To_Requirements_Yaml_File(t *testing.T) {
	assert.Regexp(t, regexp.MustCompile("^.*/requirements.yaml$"), RequirementsYamlPath("/some/dir"))
}
