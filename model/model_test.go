package model_test

import (
	"testing"

	"github.com/magicbell/mason/model"
	"gotest.tools/v3/assert"
)

func TestErrorsAreSorted(t *testing.T) {
	e := model.ValidationError{Errors: []model.FieldError{
		{Message: "bbb"},
		{Message: "aaa"},
	}}
	want := model.ValidationError{Errors: []model.FieldError{
		{Message: "aaa"},
		{Message: "bbb"},
	}}

	model.SortErrors(&e)

	assert.DeepEqual(t, e, want)
}
