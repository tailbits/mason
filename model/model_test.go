package model_test

import (
	"testing"

	"github.com/tailbits/mason/model"
	"gotest.tools/v3/assert"
)

func TestErrorsAreSorted(t *testing.T) {
	e := model.ValidationError{Errors: []model.FieldError{
		{Message: "bbb"},
		{Message: "aaa"},
	}}

	// Sort and assert on message order only to avoid
	// unexported fields in model.FieldError tripping equality.
	model.SortErrors(&e)
	got := []string{e.Errors[0].Message, e.Errors[1].Message}
	want := []string{"aaa", "bbb"}
	assert.DeepEqual(t, got, want)
}
