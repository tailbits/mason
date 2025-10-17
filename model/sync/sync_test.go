package sync_test

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/swaggest/jsonschema-go"
	"github.com/tailbits/mason/model"
	"github.com/tailbits/mason/model/sync"
)

var _ model.Entity = (*TestModel)(nil)

type TestModel struct {
	Count       *int       `json:"count"`
	DiscardedAt *time.Time `json:"discarded_at"`
	Omittable   string     `json:"omittable,omitempty"`
}

// Example implements apiv2.Entity.
func (t *TestModel) Example() []byte {
	return []byte(`{"count": 1}`)
}

// Marshal implements apiv2.Entity.
func (t *TestModel) Marshal() (json.RawMessage, error) {
	return json.Marshal(t)
}

// Name implements apiv2.Entity.
func (t *TestModel) Name() string {
	return "TestCase"
}

// Schema implements apiv2.Entity.
func (t *TestModel) Schema() []byte {
	return []byte(`{}`)
}

// Unmarshal implements apiv2.Entity.
func (t *TestModel) Unmarshal(data json.RawMessage) error {
	return json.Unmarshal(data, t)
}

type TestCase struct {
	Name string
	Sch  []byte
	Err  error
}

var testCases = []TestCase{
	{
		Name: "valid schema",
		Sch:  []byte(`{"type":"object","properties":{"count":{"type":["integer", "null"]},"discarded_at":{"type":["string", "null"],"format":"date-time"},"omittable":{"type": "string"}},"required":["count","discarded_at"]}`),
		Err:  nil,
	},
	{
		Name: "error: not nullable",
		Sch:  []byte(`{"type":"object","properties":{"count":{"type":["integer"]},"discarded_at":{"type":["string", "null"],"format":"date-time"},"omittable":{"type": "string"}},"required":["count","discarded_at"]}`),
		Err:  &sync.NullableFieldError{Message: "count must be nullable"},
	},
	{
		Name: "error: string -> int",
		Sch:  []byte(`{"type":"object","properties":{"count":{"type":["string", "null"]},"discarded_at":{"type":["string", "null"],"format":"date-time"},"omittable":{"type": "string"}},"required":["count","discarded_at"]}`),
		Err:  &sync.SchemaTypeError{Expected: "string", Got: reflect.Int},
	},
	{
		Name: "error: map -> int",
		Sch:  []byte(`{"type":"object","properties":{"count":{"type":["object", "null"]},"discarded_at":{"type":["string", "null"],"format":"date-time"},"omittable":{"type": "string"}},"required":["count","discarded_at"]}`),
		Err:  &sync.SchemaTypeError{Expected: "map or struct", Got: reflect.Int},
	},
	{
		Name: "error: extra property",
		Sch:  []byte(`{"type":"object","properties":{"count":{"type":["integer", "null"]},"discarded_at":{"type":["string", "null"],"format":"date-time"},"omittable":{"type": "string"},"extra":{"type": "string"}},"required":["count","discarded_at"]}`),
		Err:  &sync.AdditionalPropertyError{Property: "extra"},
	},
	// {
	// 	Name: "error: missing marked as required",
	// 	Sch:  []byte(`{"type":"object","properties":{"count":{"type":["integer", "null"]},"discarded_at":{"type":["string", "null"],"format":"date-time"},"omittable":{"type": "string"}},"required":["count"]}`),
	// 	Err:  &sync.RequiredPropertyError{Property: "discarded_at"},
	// },
}

func TestSchemaSync(t *testing.T) {

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			validator, err := sync.New(&TestModel{})
			if err != nil {
				t.Fatalf("failed to create validator for %s: %v", "TestCase", err)
			}

			validator.Sch = &jsonschema.Schema{}
			if err := json.Unmarshal(testCase.Sch, validator.Sch); err != nil {
				t.Fatalf("failed to unmarshal schema for %s: %v", "TestCase", err)
			}

			err = validator.IsSynced()
			if err == nil && testCase.Err != nil {
				t.Fatalf("expected error %s, got nil", testCase.Err)
			}

			if err != nil {
				if testCase.Err == nil {
					t.Fatalf("unexpected error: %v", err)
				}

				switch testCaseErr := testCase.Err.(type) {
				case *sync.SchemaTypeError:
					if !errors.As(err, &testCaseErr) {
						t.Fatalf("expected error %s, got %v", testCaseErr, err)
					}
				case *sync.AdditionalPropertyError:
					if !errors.As(err, &testCaseErr) {
						t.Fatalf("expected error %s, got %v", testCaseErr, err)
					}
				case *sync.RequiredPropertyError:
					if !errors.As(err, &testCaseErr) {
						t.Fatalf("expected error %s, got %v", testCaseErr, err)
					}
				case *sync.NullableFieldError:
					if !errors.As(err, &testCaseErr) {
						t.Fatalf("expected error %s, got %v", testCaseErr, err)
					}

				default:
					t.Fatalf("unexpected error type: %T", err)
				}
			}
		})
	}

}
