package mason

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"

	"github.com/tailbits/mason/model"
)

type decodeOptions struct{}

type DecodeOption func(options *decodeOptions) error

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

	if err := model.Validate(schema, body); err != nil {
		return ent, fmt.Errorf("model.Validate: %w", err)
	}

	// If the entity is a pointer, we need to create a new instance of the entity,
	// or else "ent" will be a nil pointer.
	switch {
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
