package cli

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"regexp"
	"testing"
)

func Test_Green_Returns_String_Prefixed_With_ANSI_Escape_Green_Code(t *testing.T) {
	str := "some-str"
	assert.Regexp(t, regexp.MustCompile(fmt.Sprintf("^\\033\\[32m%s.*", str)), Green(str))
}

func Test_Green_Returns_String_Affixed_With_ANSI_Escape_White_Code(t *testing.T) {
	str := "some-str"
	assert.Regexp(t, regexp.MustCompile(fmt.Sprintf("^.*%s\\033\\[97m$", str)), Green(str))
}

func Test_Orange_Returns_String_Prefixed_With_ANSI_Escape_Orange_Code(t *testing.T) {
	str := "some-str"
	assert.Regexp(t, regexp.MustCompile(fmt.Sprintf("^\\033\\[33m%s.*", str)), Orange(str))
}

func Test_Orange_Returns_String_Affixed_With_ANSI_Escape_White_Code(t *testing.T) {
	str := "some-str"
	assert.Regexp(t, regexp.MustCompile(fmt.Sprintf("^.*%s\\033\\[97m$", str)), Orange(str))
}
