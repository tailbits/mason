package sync

import (
	"fmt"
	"reflect"
)

type ValidationError struct {
	Breadcrumbs string
	Err         error
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %v", e.Breadcrumbs, e.Err)
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

// Specific error types
type SchemaTypeError struct {
	Expected    string
	Got         reflect.Kind
	Breadcrumbs string
}

func (e *SchemaTypeError) Error() string {
	return fmt.Sprintf("got %s when schema expects %s", e.Got, e.Expected)
}

type NullableFieldError struct {
	Message     string
	Breadcrumbs string
}

func (e *NullableFieldError) Error() string {
	return e.Message
}

type MissingPropertyError struct {
	Property string
}

func (e *MissingPropertyError) Error() string {
	return fmt.Sprintf("schema is missing property %s", e.Property)
}

type AdditionalPropertyError struct {
	Property    string
	Breadcrumbs string
}

func (e *AdditionalPropertyError) Error() string {
	return fmt.Sprintf("schema has an additional property %s", e.Property)
}

type RequiredPropertyError struct {
	Property    string
	Breadcrumbs string
}

func (e *RequiredPropertyError) Error() string {
	return fmt.Sprintf("schema is missing required property (alternatively mark the field with omitempty) %s", e.Property)
}
