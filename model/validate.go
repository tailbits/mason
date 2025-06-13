package model

import (
	"errors"
	"fmt"

	"github.com/xeipuuv/gojsonschema"
)

// ErrBodyEmpty occurs when the body of the reponse was empty.
var ErrBodyEmpty = errors.New("body empty")

// Validate validates the provided model against it's declared tags.
func Validate(schema []byte, body []byte) error {
	if len(body) == 0 {
		return fmt.Errorf(
			"validateBodySchema: %w %w",
			ValidationError{Errors: []FieldError{{Message: "body is empty"}}},
			ErrBodyEmpty,
		)
	}

	doc := gojsonschema.NewBytesLoader(schema)
	sch, err := gojsonschema.NewSchema(doc)
	if err != nil {
		return fmt.Errorf("gojsonschema.NewSchema: [%s] %w", doc, err)
	}

	res, err := sch.Validate(gojsonschema.NewBytesLoader(body))
	if err != nil {
		return fmt.Errorf("json schema validate: %w", err)
	}

	if !res.Valid() {
		return ToValidationError(res)
	}

	return nil
}
