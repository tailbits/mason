package model

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	"github.com/xeipuuv/gojsonschema"
)

// JSONFieldError is used to indicate an error with a specific request field.
type JSONFieldError struct {
	Message string `json:"message"`
}

// JSONFieldErrors represents a collection of field errors.
type JSONFieldErrors struct {
	Errors []JSONFieldError `json:"errors"`
}

// Error implements the error interface on FieldErrors.
func (fe JSONFieldErrors) Error() string {
	d, err := json.Marshal(fe)
	if err != nil {
		return err.Error()
	}
	return string(d)
}

func NewJSONFieldErrors(msgs []string) JSONFieldErrors {
	errs := make([]JSONFieldError, 0, len(msgs))
	for _, msg := range msgs {
		errs = append(errs, JSONFieldError{
			Message: msg,
		})
	}
	ret := JSONFieldErrors{Errors: errs}
	SortErrors(&ret)
	return ret
}

func toFieldErrors(result *gojsonschema.Result) JSONFieldErrors {
	errs := make([]JSONFieldError, 0, len(result.Errors()))
	for _, res := range result.Errors() {
		switch res.(type) {
		case *gojsonschema.NumberAllOfError, *gojsonschema.NumberAnyOfError, *gojsonschema.NumberOneOfError:
			continue
		default:
			errs = append(errs, JSONFieldError{
				Message: newErrorMessage(res),
			})
		}
	}

	res := JSONFieldErrors{
		Errors: errs,
	}
	SortErrors(&res)

	return res
}

func SortErrors(e *JSONFieldErrors) {
	slices.SortFunc(e.Errors, func(a, b JSONFieldError) int { return cmp.Compare(a.Message, b.Message) })
}

// IsJSONFieldError checks if an error of type FieldErrors exists.
func IsJSONFieldError(err error) bool {
	var fe JSONFieldErrors
	return errors.As(err, &fe)
}

// GetJSONFieldErrors returns a copy of the FieldErrors pointer.
func GetJSONFieldErrors(err error) JSONFieldErrors {
	var fe JSONFieldErrors
	if !errors.As(err, &fe) {
		return JSONFieldErrors{}
	}

	return fe
}

// =============================================================================

func newErrorMessage(resErr gojsonschema.ResultError) string {
	switch resErr.(type) {
	case *gojsonschema.RequiredError:
		return fmt.Sprintf("Param '%s' is missing", resErr.Details()["property"])
	case *gojsonschema.StringLengthGTEError:
		return fmt.Sprintf("Param '%s' is too short", resErr.Field())
	case *gojsonschema.StringLengthLTEError:
		return fmt.Sprintf("Param '%s' is too long", resErr.Field())
	case *gojsonschema.ArrayMinItemsError:
		return fmt.Sprintf("Param '%s' must contain atleast %d items", resErr.Field(), resErr.Details()["min"])
	case *gojsonschema.ArrayMaxItemsError:
		return fmt.Sprintf("Param '%s' must contain at most %d items", resErr.Field(), resErr.Details()["max"])
	case *gojsonschema.AdditionalPropertyNotAllowedError:
		return fmt.Sprintf("Param '%s' doesn't allow key: %s", resErr.Field(), resErr.Details()["property"])
	case *gojsonschema.InvalidTypeError:
		return fmt.Sprintf("Param '%s' should be of type %s", resErr.Field(), resErr.Details()["expected"])
	case *gojsonschema.DoesNotMatchPatternError:
		return fmt.Sprintf("Param '%s' should match pattern %s", resErr.Field(), resErr.Details()["pattern"])
	case *gojsonschema.DoesNotMatchFormatError:
		return fmt.Sprintf("Param '%s' should be a valid %s", resErr.Field(), resErr.Details()["format"])

	// case *gojsonschema.FalseError:
	// case *gojsonschema.InvalidTypeError:
	// case *gojsonschema.NumberAnyOfError:
	// case *gojsonschema.NumberOneOfError:
	// case *gojsonschema.NumberAllOfError:
	// case *gojsonschema.NumberNotError:
	// case *gojsonschema.MissingDependencyError:
	// case *gojsonschema.InternalError:
	// case *gojsonschema.ConstError:
	// case *gojsonschema.EnumError:
	// case *gojsonschema.ArrayNoAdditionalItemsError:
	// case *gojsonschema.ArrayMaxItemsError:
	// case *gojsonschema.ItemsMustBeUniqueError:
	// case *gojsonschema.ArrayContainsError:
	// case *gojsonschema.ArrayMinPropertiesError:
	// case *gojsonschema.ArrayMaxPropertiesError:
	// case *gojsonschema.InvalidPropertyPatternError:
	// case *gojsonschema.InvalidPropertyNameError:
	// case *gojsonschema.MultipleOfError:
	// case *gojsonschema.NumberGTEError:
	// case *gojsonschema.NumberGTError:
	// case *gojsonschema.NumberLTEError:
	// case *gojsonschema.NumberLTError:
	// case *gojsonschema.ConditionThenError:
	// case *gojsonschema.ConditionElseError:

	default:
		return fmt.Sprintf("[%T]: %s", resErr, resErr.Description())
	}
}
