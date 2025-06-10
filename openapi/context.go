package openapi

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/magicbell/mason"
	"github.com/swaggest/jsonschema-go"
	"github.com/swaggest/openapi-go"
	"github.com/swaggest/openapi-go/openapi31"
)

type ContextWrapper struct {
	openapi.OperationContext
	*openapi31.Operation
	reflector *ReflectorWrapper
}

func (c ContextWrapper) addToReflector() error {
	return c.reflector.AddOperation(c.OperationContext)
}

// from takes a Record and uses it to populate the ContextWrapper with the necessary information to generate an OpenAPI operation.
func (c *ContextWrapper) from(record Record) error {
	if !record.Output.IsNil() {
		if err := c.addRespStructure(record.Output, openapi.WithHTTPStatus(record.SuccessStatus)); err != nil {
			return err
		}
	}

	if record.Input != nil && !record.Input.IsNil() {
		if err := c.addReqStructure(*record.Input); err != nil {
			return err
		}
	}

	pathParams := []openapi31.ParameterOrReference{}
	forEachPathParam(record.Method, record.Path, func(param string) {
		pathParams = append(pathParams, makeRequiredPathParam(param))
	})

	forEachQueryParam(record.QueryParams, func(name string, t string) {
		pathParams = append(pathParams, makeOptionalQueryParam(name, t))
	})

	c.WithParameters(pathParams...)

	c.WithID(record.ID)
	c.WithTags(record.Tags...)
	for _, tag := range record.Tags {
		c.reflector.allTags[tag] = true
	}
	c.SetDescription(record.Description)

	if record.Summary != "" {
		c.SetSummary(record.Summary)
	}

	if record.Extensions != nil {
		c.Operation.WithMapOfAnything(record.Extensions)
	}

	return nil
}

// addReqStructure provides duplicate-detection to the openapi-go AddReqStructure method.
func (c ContextWrapper) addReqStructure(o mason.Model, options ...openapi.ContentOption) error {
	if err := c.reflector.addModel(o); err != nil {
		return fmt.Errorf("failed to add definition for %s: %w", o.Name(), err)
	}

	c.OperationContext.AddReqStructure(o, options...)

	return nil
}

// addRespStructure provides duplicate-detection to the openapi-go AddRespStructure method.
func (c ContextWrapper) addRespStructure(o mason.Model, options ...openapi.ContentOption) error {
	if err := c.reflector.addModel(o); err != nil {
		return fmt.Errorf("failed to add definition for %s: %w", o.Name(), err)
	}

	c.OperationContext.AddRespStructure(o, options...)

	return nil
}

func NewContextWrapper(ctx openapi.OperationContext, r *ReflectorWrapper) *ContextWrapper {
	ctxWrapper := ContextWrapper{
		OperationContext: ctx,
		reflector:        r,
	}
	if opExp, ok := ctx.(openapi31.OperationExposer); ok {
		ctxWrapper.Operation = opExp.Operation()
	}

	return &ctxWrapper
}

/* -------------------------------------------------------------------------- */

func forEachPathParam(method string, path string, f func(string)) {
	_, _, params, _ := openapi.SanitizeMethodPath(method, path)
	for _, p := range params {
		f(p)
	}
}

func makeRequiredPathParam(param string) openapi31.ParameterOrReference {
	req := true
	s, err := jsonschema.String.ToSchemaOrBool().ToSimpleMap()
	if err != nil {
		return openapi31.ParameterOrReference{}
	}

	return openapi31.ParameterOrReference{
		Parameter: &openapi31.Parameter{
			Name:     param,
			In:       openapi31.ParameterInPath,
			Required: &req,
			Schema:   s,
		},
	}
}

func forEachQueryParam(queryParams any, f func(string, string)) {
	if queryParams == nil {
		return
	}

	t := reflect.TypeOf(queryParams)
	if t.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("json")
		tag = strings.Split(tag, ",")[0]
		if tag == "" {
			continue
		}

		switch field.Type.Kind() {
		case reflect.String:
			f(tag, "string")
		case reflect.Int:
			f(tag, "integer")
		case reflect.Bool:
			f(tag, "boolean")
		case reflect.Ptr:
			switch field.Type.Elem().Kind() {
			case reflect.String:
				f(tag, "string")
			case reflect.Int:
				f(tag, "integer")
			case reflect.Bool:
				f(tag, "boolean")
			}
		}
	}
}

func makeOptionalQueryParam(name string, t string) openapi31.ParameterOrReference {
	req := false
	s, err := jsonschema.SimpleType(t).ToSchemaOrBool().ToSimpleMap()
	if err != nil {
		return openapi31.ParameterOrReference{}
	}

	return openapi31.ParameterOrReference{
		Parameter: &openapi31.Parameter{
			Name:     name,
			In:       openapi31.ParameterInQuery,
			Required: &req,
			Schema:   s,
		},
	}
}
