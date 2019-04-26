package deployment

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"regexp"
	"testing"
)

func Test_ChartValuesPath_Returns_Path_Prefixed_With_Chart_Dir(t *testing.T) {
	chartDir := "/some/dir"
	assert.Regexp(t, regexp.MustCompile(fmt.Sprintf("^%s/.*", chartDir)), ChartValuesPath(chartDir, "prod"))
}

func Test_ChartValuesPath_Returns_Path_Affixed_With_Yaml_Extension(t *testing.T) {
	assert.Regexp(t, regexp.MustCompile("^.*\\.yaml$"), ChartValuesPath("/some/dir", "prod"))
}

func Test_ChartValuesPath_Returns_Path_To_TargetEnv_Yaml_File(t *testing.T) {
	targetEnv := "yayForDeployments"
	assert.Regexp(t, regexp.MustCompile(fmt.Sprintf("^.*/%s.yaml$", targetEnv)), ChartValuesPath("/some/dir", targetEnv))
}
