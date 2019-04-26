package runtime 

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_PanicIfError_Panics_When_Error_Is_Not_Nil(t *testing.T) {
	assert.Panics(t, func() { PanicIfError(errors.New("you'd better panic")) })
}

func Test_PanicIfError_Does_Not_Panic_When_Error_Is_Nil(t *testing.T) {
	assert.NotPanics(t, func() { PanicIfError(nil ) })
}
