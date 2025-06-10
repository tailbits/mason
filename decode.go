package mason

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"

	"github.com/magicbell/mason/model"
)

type decodeOptions struct {
	wrappingKey *string
}

type wrapped[T model.WithSchema] map[string]T

type DecodeOption func(options *decodeOptions) error

// WithWrappingKey is a DecodeOption that indicates that we expect the incoming JSON to be wrapped in an object with key.
func WithWrappingKey(key string) DecodeOption {
	return func(options *decodeOptions) error {
		options.wrappingKey = &key
		return nil
	}
}

func DecodeRequest[T model.Entity](api *API, r *http.Request, opts ...DecodeOption) (ent T, err error) {
	if ent.Name() == "NilEntity" {
		return ent, nil
	}

	var options decodeOptions
	for _, opt := range opts {
		err := opt(&options)
		if err != nil {
			return ent, err
		}
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return ent, fmt.Errorf("unable to read the body: %w", err)
	}
	// restore the body for the next handler in the chain
	r.Body = io.NopCloser(io.Reader(bytes.NewBuffer(body)))

	schema, err := api.DereferenceSchema(ent.Schema())
	if err != nil {
		return ent, fmt.Errorf("dereferenceSchema ent[%s]: %w", ent.Name(), err)
	}

	if options.wrappingKey != nil {
		schema = []byte(fmt.Sprintf(`{"type": "object", "properties": {"%s": %s }, "required": ["%s"]}`, *options.wrappingKey, schema, *options.wrappingKey))
	}

	if err := model.Validate(schema, body); err != nil {
		return ent, fmt.Errorf("validate.Validate: %w", err)
	}

	// do the unmarshalling. note: if the entity is a pointer,
	// we need to create a new instance of the entity else "ent" will be a nil pointer.
	// todo: defer to the unmarshal function that already exists on T
	switch {
	case options.wrappingKey != nil:
		wrapped := make(wrapped[T])
		if err := json.Unmarshal(body, &wrapped); err != nil {
			return ent, fmt.Errorf("wrapped.Unmarshal: %w", err)
		}

		return wrapped[*options.wrappingKey], nil
	case reflect.TypeOf(ent).Kind() == reflect.Ptr:
		elemType := reflect.TypeOf(ent).Elem()
		newEnt := reflect.New(elemType).Interface()

		var ok bool
		if ent, ok = newEnt.(T); !ok {
			return ent, fmt.Errorf("type assertion failed for entity of type %T", newEnt)
		}

		if err := json.Unmarshal(body, ent); err != nil {
			return ent, fmt.Errorf("unable to unmarshal the data: %w", err)
		}

		return ent, nil
	default:
		if err := json.Unmarshal(body, &ent); err != nil {
			return ent, fmt.Errorf("unable to unmarshal the data: %w", err)
		}

		return ent, nil
	}
}
