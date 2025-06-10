// Package model contains the data models for the Mason API.
package model

import (
	"fmt"
	"reflect"
)

type DerivedType interface {
	Unwrap() WithSchema
}

func New[T any]() T {
	var t T
	if reflect.TypeOf(t).Kind() == reflect.Ptr {
		newT := reflect.New(reflect.TypeOf(t).Elem()).Interface()
		if result, ok := newT.(T); ok {
			t = result
		} else {
			// This should never happen due to Go's type system,
			// but we handle it to satisfy the linter
			panic(fmt.Sprintf("Unexpected type assertion failure in Initialize[T]: expected %T, got %T\n", t, newT))
		}
	}
	if t, ok := any(t).(Entity); ok {
		if err := t.Unmarshal(t.Example()); err != nil {
			// Handle the error, e.g., log it or return a default value
			panic(err)
		}
	}

	return t
}
