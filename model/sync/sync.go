// Package sync provides utilities for vetting models against their declared schemas.
package sync

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/swaggest/jsonschema-go"
	"github.com/tailbits/mason"
	"github.com/tailbits/mason/model"
)

type ShouldSkip interface {
	SkipSchemaValidation() bool
}

type Validator struct {
	Sch   *jsonschema.Schema
	Model any
	Name  string
}

func New(api *mason.API, model model.Entity) (*Validator, error) {
	sch, err := api.DereferenceSchema(model.Schema())
	if err != nil {
		return nil, fmt.Errorf("error dereferencing schema for %s: %w", model.Name(), err)
	}

	return new(model.Name(), model, sch)
}

// serverDefined returns true if we expect the field to be on the struct, but not in the schema (i.e. it is defined on the server rather than in the incoming payload)
func (v *Validator) serverDefined(fieldName string) bool {
	switch fieldName {
	case "id":
		return true
	default:
		return false
	}
}

func (v *Validator) IsSynced() error {
	return v.traverse(v.Sch, reflect.ValueOf(v.Model), false, v.Name)
}

func (v *Validator) isByteArray(val reflect.Value) bool {
	return (val.Kind() == reflect.Slice || val.Kind() == reflect.Array) && val.Type().Elem().Kind() == reflect.Uint8
}

func (v *Validator) isJSONRawMessage(val reflect.Value) bool {
	return v.isByteArray(val) && val.Type() == reflect.TypeOf(json.RawMessage{})
}

func (v *Validator) isTimestamp(val reflect.Value) bool {
	return val.Kind() == reflect.Struct && val.Type() == reflect.TypeOf(time.Time{})
}

func (v *Validator) isInterface(val reflect.Value) bool {
	return val.Kind() == reflect.Interface
}

func (v *Validator) getType(sch *jsonschema.Schema) (t string, nullable bool, err error) {
	if sch.Type == nil {
		return "", false, fmt.Errorf("schema is missing a type")
	}

	// if the type is a simple type, just coerce it to a string and return it
	if sch.Type.SimpleTypes != nil {
		return string(*sch.Type.SimpleTypes), false, nil
	}

	// otherwise it is a type enum
	// the only type enums we support are of the form ["null", "type"]
	nullable = false
	types := []string{}
	for _, t := range sch.Type.SliceOfSimpleTypeValues {
		if t == "null" {
			nullable = true
		} else {
			types = append(types, string(t))
		}
	}

	if len(types) > 1 {
		return "", false, fmt.Errorf("multiple types are not supported")
	}

	return types[0], nullable, nil
}

func (v *Validator) ensureDereference(sch *jsonschema.Schema) (*jsonschema.Schema, bool, error) {
	if sch.Ref != nil {
		defPrefix := "#/definitions/"
		if !strings.HasPrefix(*sch.Ref, defPrefix) {
			return nil, false, fmt.Errorf("references must be prefixed with %s", defPrefix)
		}
		key := strings.TrimPrefix(*sch.Ref, defPrefix)
		ref, ok := v.Sch.Definitions[key]
		if !ok || ref.TypeObject == nil {
			return nil, false, fmt.Errorf("could not find reference %s", *sch.Ref)
		}
		return ref.TypeObject, false, nil
	}

	nullable := false
	inner := sch
	var err error

	// allow oneOf{[null, ref]}
	if sch.OneOf != nil {
		for _, s := range sch.OneOf {
			if s.TypeObject == nil {
				continue
			}
			if s.TypeObject.Type != nil && s.TypeObject.Type.SimpleTypes != nil && *s.TypeObject.Type.SimpleTypes == "null" {
				nullable = true
			} else if s.TypeObject.Ref != nil {
				nullableRef := false
				inner, nullableRef, err = v.ensureDereference(s.TypeObject)
				if err != nil {
					return nil, false, err
				}
				nullable = nullable || nullableRef
			}
		}
	}

	return inner, nullable, nil
}

func (v *Validator) checkArray(sch []jsonschema.SchemaOrBool, val reflect.Value, breadcrumbs string) error {
	for i, s := range sch {
		if err := v.traverse(s.TypeObject, val, false, breadcrumbs+"."+strconv.Itoa(i)); err != nil {
			return err
		}
	}
	return nil
}

func (v *Validator) checkMap(sch map[string]jsonschema.SchemaOrBool, val reflect.Value, breadcrumbs string) error {
	for k, s := range sch {
		if err := v.traverse(s.TypeObject, val, false, breadcrumbs+"."+k); err != nil {
			return err
		}
	}
	return nil
}

func (v *Validator) isIntegerValue(val reflect.Value) bool {
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	default:
		return false
	}
}

func (v *Validator) isStrict(sch *jsonschema.Schema) bool {
	return sch.AdditionalProperties != nil && sch.AdditionalProperties.TypeBoolean != nil && !*sch.AdditionalProperties.TypeBoolean
}

func (v *Validator) isNumericValue(val reflect.Value) bool {
	if v.isIntegerValue(val) {
		return true
	}
	switch val.Kind() {
	case reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}

func (v *Validator) traverse(sch *jsonschema.Schema, val reflect.Value, omitEmpty bool, breadcrumbs string) error {
	// Check if the value's type implements ShouldSkip
	if val.IsValid() {
		var shouldSkipType = reflect.TypeOf((*ShouldSkip)(nil)).Elem()
		var receiver reflect.Value

		if val.Type().Implements(shouldSkipType) {
			receiver = val
		} else if val.CanAddr() {
			addr := val.Addr()
			if addr.Type().Implements(shouldSkipType) {
				receiver = addr
			}
		}

		if receiver.IsValid() {
			if shouldSkip, ok := receiver.Interface().(ShouldSkip); ok {
				if shouldSkip.SkipSchemaValidation() {
					return nil // Skip validation for this value
				}
			}
		}
	}

	if sch == nil {
		if v.isInterface(val) {
			return fmt.Errorf("%s: a non-interface value (%s) should have a definite schema", breadcrumbs, val.Kind())
		}
		return nil
	}

	sch, nullableRefType, err := v.ensureDereference(sch)
	if err != nil {
		return fmt.Errorf("%s: %w", breadcrumbs, err)
	}

	t, nullableSimpleType, err := v.getType(sch)
	if err != nil {
		return fmt.Errorf("%s: %w", breadcrumbs, err)
	}

	_, isEntity := val.Interface().(model.WithSchema)

	nullable := nullableRefType || nullableSimpleType

	if val.Kind() == reflect.Ptr {
		if breadcrumbs != v.Name && !isEntity && !nullable && !omitEmpty {
			// all field pointers are potentially nullable: exclude root object and model.WithSchema field values
			return &NullableFieldError{Message: fmt.Sprintf("%s: must be nullable", breadcrumbs)}
		}
		if val.IsNil() {
			val = reflect.New(val.Type().Elem()).Elem()
		} else {
			val = val.Elem()
		}
	}

	if val.Kind() == reflect.Map && !isEntity && !nullable && !omitEmpty {
		return &NullableFieldError{Message: fmt.Sprintf("%s: must be nullable", breadcrumbs)}
	}

	if v.isJSONRawMessage(val) {
		// if the field is a json.RawMessage, then we leave it up to the schema to declare structure
		return nil
	}

	if v.isInterface(val) {
		// if the field is an interface, then we leave it up to the schema to declare structure
		return nil
	}

	switch t {
	case "boolean":
		if val.Kind() != reflect.Bool {
			return &SchemaTypeError{Expected: "boolean", Got: val.Kind(), Breadcrumbs: breadcrumbs}
		}
	case "integer":
		if !v.isIntegerValue(val) {
			return &SchemaTypeError{Expected: "integer", Got: val.Kind(), Breadcrumbs: breadcrumbs}
		}
	case "number":
		if !v.isNumericValue(val) {
			return &SchemaTypeError{Expected: "number", Got: val.Kind(), Breadcrumbs: breadcrumbs}
		}
	case "string":
		if val.Kind() != reflect.String && !v.isByteArray(val) && !v.isTimestamp(val) {
			return &SchemaTypeError{Expected: "string", Got: val.Kind(), Breadcrumbs: breadcrumbs}
		}
	case "object":
		switch val.Kind() {
		case reflect.Struct:
			for i := 0; i < val.NumField(); i++ {
				field := val.Field(i)
				fieldType := val.Type().Field(i)
				tag := fieldType.Tag.Get("json")
				if tag == "" || tag == "-" {
					continue
				}
				name, opts := parseTag(tag)
				if name != "" {
					tag = name
				}

				schOrBool, ok := sch.Properties[tag]
				if !ok && !v.serverDefined(tag) {
					return fmt.Errorf("%s: schema is missing property %s", breadcrumbs, tag)
				}

				// isRequired := false
				// for _, r := range sch.Required {
				// 	if r == tag {
				// 		isRequired = true
				// 		break
				// 	}
				// }

				// fmt.Println(opts, isRequired)

				// only allow non-required fields if they are also marked with omitempty
				// if !opts.Contains("omitempty") && !isRequired {
				// 	return &RequiredPropertyError{Property: tag, Breadcrumbs: breadcrumbs}
				// }

				if sch.AdditionalProperties != nil && *sch.AdditionalProperties.TypeBoolean {
					return fmt.Errorf("%s: struct schemas should not allow additional properties", breadcrumbs)
				}

				sch := schOrBool.TypeObject
				if sch == nil {
					continue
				}

				if err := v.traverse(sch, field, opts.Contains("omitempty"), breadcrumbs+"."+tag); err != nil {
					return err
				}
			}
			for k := range sch.Properties {
				found := false
				for i := 0; i < val.NumField(); i++ {
					fieldType := val.Type().Field(i)
					tag := fieldType.Tag.Get("json")
					if tag == "" {
						continue
					}
					tag = strings.Split(tag, ",")[0]
					if tag == k {
						found = true
						break
					}
				}
				if !found {
					return &AdditionalPropertyError{Property: k, Breadcrumbs: breadcrumbs}
				}
			}
		case reflect.Map:
			if v.isStrict(sch) {
				return fmt.Errorf("%s: schema strictly enumerates all valid keys (e.g. %s); the appropriate data type for unmarshalling would be a struct, not a map", breadcrumbs, getMapKeys(sch.Properties))
			}
			// get map value type
			mr := val.Type().Elem()
			mapValue := reflect.New(mr).Elem()

			if sch.AdditionalProperties != nil && sch.AdditionalProperties.TypeObject != nil {
				if err := v.traverse(sch.AdditionalProperties.TypeObject, mapValue, false, breadcrumbs+"[key]"); err != nil {
					return err
				}
			}

			if err := v.checkMap(sch.Properties, mapValue, breadcrumbs+"[key]"); err != nil {
				return err
			}

		default:
			return &SchemaTypeError{Expected: "map or a struct", Got: val.Kind(), Breadcrumbs: breadcrumbs}
		}
	case "array":
		if val.Kind() != reflect.Slice {
			return &SchemaTypeError{Expected: "array or slice", Got: val.Kind(), Breadcrumbs: breadcrumbs}
		}
		// get array value type
		ar := val.Type().Elem()
		elementVal := reflect.New(ar).Elem()

		if sch.AllOf != nil {
			if err := v.checkArray(sch.AllOf, elementVal, breadcrumbs); err != nil {
				return fmt.Errorf("%s: allOf failed %s", breadcrumbs, err)
			}
		}

		if sch.OneOf != nil {
			if err := v.checkArray(sch.OneOf, elementVal, breadcrumbs); err != nil {
				return fmt.Errorf("%s: oneOf failed %s", breadcrumbs, err)
			}
		}

		if sch.AnyOf != nil {
			if err := v.checkArray(sch.AnyOf, elementVal, breadcrumbs); err != nil {
				return fmt.Errorf("%s: anyOf failed %s", breadcrumbs, err)
			}
		}

		if sch.Items != nil {
			if err := v.traverse(sch.Items.SchemaOrBool.TypeObject, elementVal, false, breadcrumbs+".0"); err != nil {
				return err
			}
		}
	default:
		t, _, _ := v.getType(sch)
		return fmt.Errorf("%s: unknown type %s", breadcrumbs, t)
	}
	return nil
}

func new(name string, model any, sch []byte) (*Validator, error) {
	parsed := jsonschema.Schema{} // nolint:golint,exhaustruct
	err := parsed.UnmarshalJSON(sch)
	if err != nil {
		return nil, err
	}
	return &Validator{
		Sch:   &parsed,
		Model: model,
		Name:  name,
	}, nil
}

func getMapKeys[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
