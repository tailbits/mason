package model

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	"github.com/xeipuuv/gojsonschema"
)

// FieldError is used to indicate an error with a specific request field.
type FieldError struct {
	field   string
	details map[string]interface{}
	Message string `json:"message"`
}

func (fe FieldError) Field() string {
	return fe.field
}

func (fe FieldError) Details() map[string]interface{} {
	return fe.details
}

// ValidationError represents a collection of field errors.
type ValidationError struct {
	Errors []FieldError `json:"errors"`
}

// Error implements the error interface on FieldErrors.
func (fe ValidationError) Error() string {
	d, err := json.Marshal(fe)
	if err != nil {
		return err.Error()
	}

	return string(d)
}

func ToValidationError(result *gojsonschema.Result) ValidationError {
	errs := make([]FieldError, 0, len(result.Errors()))
	for _, res := range result.Errors() {
		switch res.(type) {
		case *gojsonschema.NumberAllOfError, *gojsonschema.NumberAnyOfError, *gojsonschema.NumberOneOfError:
			continue
		default:
			errs = append(errs, FieldError{
				field:   res.Field(),
				details: res.Details(),
				Message: newErrorMessage(res),
			})
		}
	}

	res := ValidationError{
		Errors: errs,
	}
	SortErrors(&res)

	return res
}

func SortErrors(e *ValidationError) {
	slices.SortFunc(e.Errors, func(a, b FieldError) int { return cmp.Compare(a.Message, b.Message) })
}

// IsJSONFieldError checks if an error of type FieldErrors exists.
func IsJSONFieldError(err error) bool {
	var fe ValidationError
	return errors.As(err, &fe)
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
