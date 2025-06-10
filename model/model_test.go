package model_test

import (
	"testing"

	"github.com/magicbell/mason/model"
	"gotest.tools/v3/assert"
)

func TestErrorsAreSorted(t *testing.T) {
	e := model.JSONFieldErrors{Errors: []model.JSONFieldError{
		{Message: "bbb"},
		{Message: "aaa"},
	}}
	want := model.JSONFieldErrors{Errors: []model.JSONFieldError{
		{Message: "aaa"},
		{Message: "bbb"},
	}}

	model.SortErrors(&e)

	assert.DeepEqual(t, e, want)
}
