package openapi

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/swaggest/jsonschema-go"
	"github.com/swaggest/openapi-go"
	"github.com/swaggest/openapi-go/openapi31"
	"github.com/tailbits/mason"
)

type ContextWrapper struct {
	openapi.OperationContext
	*openapi31.Operation
	reflector *Reflector
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

	forEachQueryParam(record.QueryParams, func(name string, t string, format string, desc string) {
		pathParams = append(pathParams, makeOptionalQueryParam(name, t, format, desc))
	})

	c.WithParameters(pathParams...)

	c.WithID(record.ID)
	c.WithTags(record.Tags...)
	for _, tag := range record.Tags {
		c.reflector.tags[tag] = true
	}
	c.SetDescription(record.Description)

	if record.Summary != "" {
		c.SetSummary(record.Summary)
	}

	if record.PathSummary != "" || record.PathDescription != "" {
		path := c.PathPattern()
		pathItem := c.reflector.Spec.PathsEns().MapOfPathItemValues[path]
		if record.PathSummary != "" {
			pathItem.WithSummary(record.PathSummary)
		}
		if record.PathDescription != "" {
			pathItem.WithDescription(record.PathDescription)
		}
		c.reflector.Spec.PathsEns().WithMapOfPathItemValuesItem(path, pathItem)
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

func NewContextWrapper(ctx openapi.OperationContext, r *Reflector) *ContextWrapper {
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

func forEachQueryParam(queryParams any, f func(string, string, string, string)) {
	if queryParams == nil {
		return
	}

	t := reflect.TypeOf(queryParams)
	if t.Kind() != reflect.Struct {
		return
	}

	descriptions := queryParamDescriptions(t)
	timeType := reflect.TypeOf(time.Time{})
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("json")
		tag = strings.Split(tag, ",")[0]
		if tag == "" {
			continue
		}

		desc := field.Tag.Get("doc")
		if desc == "" {
			desc = descriptions[field.Name]
		}
		switch field.Type.Kind() {
		case reflect.String:
			f(tag, "string", "", desc)
		case reflect.Int:
			f(tag, "integer", "", desc)
		case reflect.Bool:
			f(tag, "boolean", "", desc)
		case reflect.Struct:
			if field.Type == timeType {
				f(tag, "string", "date-time", desc)
			}
		case reflect.Ptr:
			switch field.Type.Elem().Kind() {
			case reflect.String:
				f(tag, "string", "", desc)
			case reflect.Int:
				f(tag, "integer", "", desc)
			case reflect.Bool:
				f(tag, "boolean", "", desc)
			case reflect.Struct:
				if field.Type.Elem() == timeType {
					f(tag, "string", "date-time", desc)
				}
			}
		}
	}
}

func makeOptionalQueryParam(name string, t string, format string, desc string) openapi31.ParameterOrReference {
	req := false
	var schema jsonschema.Schema
	if t != "" {
		var jt jsonschema.Type
		switch t {
		case "string":
			jt.WithSimpleTypes(jsonschema.String)
		case "integer":
			jt.WithSimpleTypes(jsonschema.Integer)
		case "boolean":
			jt.WithSimpleTypes(jsonschema.Boolean)
		case "number":
			jt.WithSimpleTypes(jsonschema.Number)
		}
		schema.WithType(jt)
	}
	if format != "" {
		schema.Format = &format
	}
	s, err := schema.ToSchemaOrBool().ToSimpleMap()
	if err != nil {
		return openapi31.ParameterOrReference{}
	}

	param := &openapi31.Parameter{
		Name:     name,
		In:       openapi31.ParameterInQuery,
		Required: &req,
		Schema:   s,
	}
	if desc != "" {
		param.WithDescription(desc)
	}
	return openapi31.ParameterOrReference{Parameter: param}
}
