package openapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/magicbell/mason"
	"github.com/magicbell/mason/model"
	"github.com/magicbell/mason/openapi"
	"gotest.tools/v3/assert"
)

// TestCase defines a test case for OpenAPI generation
type TestCase struct {
	name            string
	setupFunc       func(*mason.API) error
	expectedOutcome Outcome
}

// Outcome represents the expected result of a test case
type Outcome interface {
	Assert(t *testing.T, schema []byte, err error)
}

// ExpectSuccess asserts that OpenAPI generation succeeds and matches expected output
type ExpectSuccess struct {
	expectedFile string
}

func (e ExpectSuccess) Assert(t *testing.T, schema []byte, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("expected success but got error: %v", err)
	}

	// Format the schema
	formatted, err := formatJSON(schema)
	if err != nil {
		t.Fatalf("error formatting schema: %v", err)
	}

	// Compare with expected file
	expected, err := os.ReadFile(e.expectedFile)
	if err != nil {
		t.Fatalf("error reading expected file: %v", err)
	}

	if os.Getenv("UPDATE_SCHEMA_SNAPSHOT") == "true" {
		if err := os.WriteFile(e.expectedFile, formatted, 0644); err != nil {
			t.Fatalf("error writing expected file: %v", err)
		}
		t.Logf("updated expected file: %s", e.expectedFile)
		return
	}

	assert.Equal(t, string(expected), string(formatted), "schema does not match snapshot - run with UPDATE_SCHEMA_SNAPSHOT=true to update")
}

// ExpectError asserts that OpenAPI generation fails with a specific error message
type ExpectError struct {
	errorContains string
}

func (e ExpectError) Assert(t *testing.T, schema []byte, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error but got success")
	}

	if !strings.Contains(err.Error(), e.errorContains) {
		t.Fatalf("expected error to contain %q but got: %v", e.errorContains, err)
	}
}

func TestOpenAPIGen(t *testing.T) {
	tests := []TestCase{
		{
			name: "Basic Schema Generation",
			setupFunc: func(api *mason.API) error {
				// Set up basic test groups
				grpA := api.NewRouteGroup("TestA").NewRouteGroup("Child")
				grpA.Register(
					mason.HandlePut(CreateResourceA).
						Path("/test-a").
						WithOpID("create_test_resource").
						WithDesc("Create a test resource").
						WithTags("A").
						WithExtensions("x-meta", 1),
				)
				grpA.Register(
					mason.HandleGet(GetResourceA).
						Path("/test-a").
						WithOpID("fetch_test_resource").
						WithDesc("Get a test resource").
						WithTags("A").
						WithExtensions("x-meta", 2),
				)

				grpB := api.NewRouteGroup("TestB").NewRouteGroup("Child")
				grpB.Register(
					mason.HandleGet(GetResourceB).
						Path("/test-b").
						WithOpID("fetch_test_resource_b").
						WithDesc("Get a test resource B").
						WithTags("B"),
				)

				return nil
			},
			expectedOutcome: ExpectSuccess{
				expectedFile: "testdata/basic_schema.json",
			},
		},
		{
			name: "Missing Resource Reference",
			setupFunc: func(api *mason.API) error {
				// Set up a group with a resource that references a non-existent type
				grpC := api.NewRouteGroup("TestC").NewRouteGroup("Child")
				grpC.Register(
					mason.HandleGet(GetResourceWithMissingRef).
						Path("/test-c").
						WithOpID("fetch_resource_with_missing_ref").
						WithDesc("Get a resource with missing reference").
						WithTags("C"),
				)

				return nil
			},
			expectedOutcome: ExpectError{
				errorContains: "TestResourceC", // The missing reference
			},
		},
		{
			name: "Case-Sensitive Duplicate Models",
			setupFunc: func(api *mason.API) error {
				// Set up a group with a resource name that conflicts with TestResourceA but in lowercase
				grpA := api.NewRouteGroup("CaseSensitive").NewRouteGroup("Child")
				grpA.Register(
					mason.HandleGet(GetResourceA).
						Path("/resource-a").
						WithOpID("fetch_resource_a").
						WithDesc("Get resource A").
						WithTags("A"),
				)

				grpB := api.NewRouteGroup("TestB").NewRouteGroup("Child")
				grpB.Register(
					mason.HandleGet(GetResourceB).
						Path("/test-b").
						WithOpID("fetch_test_resource_b").
						WithDesc("Get a test resource B").
						WithTags("B"),
				)

				// Register the lowercase version which should cause a conflict
				grpC := api.NewRouteGroup("LowerCase").NewRouteGroup("Child")
				grpC.Register(
					mason.HandleGet(GetLowerResourceA).
						Path("/lower-resource-a").
						WithOpID("fetch_lower_resource_a").
						WithDesc("Get lowercase resource A").
						WithTags("A"),
				)

				return nil
			},
			expectedOutcome: ExpectError{
				errorContains: "conflicting definitions", // Should fail due to case-insensitive name conflict
			},
		},
		{
			name: "Identical Schema Definitions",
			setupFunc: func(api *mason.API) error {
				// Register a handler using TestResourceA
				grpA := api.NewRouteGroup("OriginalA").NewRouteGroup("Child")
				grpA.Register(
					mason.HandleGet(GetResourceA).
						Path("/resource-a").
						WithOpID("fetch_resource_a").
						WithDesc("Get resource A").
						WithTags("A"),
				)

				grpB := api.NewRouteGroup("OriginalB").NewRouteGroup("Child")
				grpB.Register(
					mason.HandleGet(GetResourceB).
						Path("/test-b").
						WithOpID("fetch_test_resource_b").
						WithDesc("Get a test resource B").
						WithTags("B"),
				)

				// Register a handler using an identical clone of TestResourceA
				grpC := api.NewRouteGroup("CloneA").NewRouteGroup("Child")
				grpC.Register(
					mason.HandleGet(GetCloneResourceA).
						Path("/clone-resource-a").
						WithOpID("fetch_clone_resource_a").
						WithDesc("Get clone of resource A").
						WithTags("A"),
				)

				return nil
			},
			expectedOutcome: ExpectSuccess{
				expectedFile: "testdata/identical_schema.json", // Should succeed with identical schemas
			},
		},
		{
			name: "Conflicting Schema Definitions",
			setupFunc: func(api *mason.API) error {
				// Register a handler using TestResourceA
				grpA := api.NewRouteGroup("OriginalA").NewRouteGroup("Child")
				grpA.Register(
					mason.HandleGet(GetResourceA).
						Path("/resource-a").
						WithOpID("fetch_resource_a").
						WithDesc("Get resource A").
						WithTags("A"),
				)

				grpB := api.NewRouteGroup("TestB").NewRouteGroup("Child")
				grpB.Register(
					mason.HandleGet(GetResourceB).
						Path("/test-b").
						WithOpID("fetch_test_resource_b").
						WithDesc("Get a test resource B").
						WithTags("B"),
				)

				// Register a handler using a resource with the same name but different schema
				grpC := api.NewRouteGroup("ConflictingA").NewRouteGroup("Child")
				grpC.Register(
					mason.HandleGet(GetConflictingResourceA).
						Path("/conflicting-resource-a").
						WithOpID("fetch_conflicting_resource_a").
						WithDesc("Get conflicting resource A").
						WithTags("A"),
				)

				return nil
			},
			expectedOutcome: ExpectError{
				errorContains: "different definition", // Should fail due to schema conflict
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.name == "Full Schema Generation" {
				t.Skip("fix me!")
			}

			api := mason.NewAPI(mason.NewHTTPRuntime())

			// Set up the test
			err := tc.setupFunc(api)
			if err != nil {
				t.Fatalf("error setting up test: %v", err)
			}

			gen, err := openapi.NewGenerator(api, openapi.Validate(false))
			assert.NilError(t, err, "failed to create OpenAPI generator")

			schema, err := gen.ToSchema()

			tc.expectedOutcome.Assert(t, schema, err)
		})
	}
}

// Helper function to format JSON
func formatJSON(b []byte) ([]byte, error) {
	var prettyJSON bytes.Buffer
	error := json.Indent(&prettyJSON, b, "", "  ")
	if error != nil {
		return nil, error
	}
	return prettyJSON.Bytes(), nil
}

/* -------------------------------------------------------------------------- */
// Handler functions
/* -------------------------------------------------------------------------- */

func CreateResourceA(ctx context.Context, _ *http.Request, resource *TestResourceA, query TestQuery) (*TestResourceA, error) {
	return resource, nil
}

func GetResourceA(ctx context.Context, _ *http.Request, params TestParams) (*TestResourceA, error) {
	return &TestResourceA{}, nil
}

func GetResourceB(ctx context.Context, _ *http.Request, params TestParams) (*TestResourceB, error) {
	return &TestResourceB{}, nil
}

func GetResourceWithMissingRef(ctx context.Context, _ *http.Request, params TestParams) (*ResourceWithMissingRef, error) {
	return &ResourceWithMissingRef{}, nil
}

func GetLowerResourceA(ctx context.Context, _ *http.Request, params TestParams) (*TestResourceALowerCase, error) {
	return &TestResourceALowerCase{}, nil
}

func GetCloneResourceA(ctx context.Context, _ *http.Request, params TestParams) (*TestResourceAClone, error) {
	return &TestResourceAClone{}, nil
}

func GetConflictingResourceA(ctx context.Context, _ *http.Request, params TestParams) (*TestResourceAConflicting, error) {
	return &TestResourceAConflicting{}, nil
}

/* -------------------------------------------------------------------------- */
// Resource types and implementations
/* -------------------------------------------------------------------------- */

// TestQuery and TestParams are common parameter types
type TestQuery struct {
	X string `json:"x"`
}

type TestParams struct {
	ID string `json:"id"`
}

// TestResourceA represents a test resource
type TestResourceA struct{}

func (t *TestResourceA) Example() []byte {
	return []byte(`
	{
		"x": "example",
		"y": {
			"y": "example"
		}
	}
	`)
}

func (t *TestResourceA) Marshal() (json.RawMessage, error) {
	return json.Marshal(t)
}

func (t *TestResourceA) Name() string {
	return "TestResourceA"
}

func (t *TestResourceA) Schema() []byte {
	return []byte(`
	{
		"type":"object",
		"properties": {
			"x": {"type":"string"},
			"y": {"$ref": "#/definitions/TestResourceB"}
		}
	}
	`)
}

func (t *TestResourceA) Unmarshal(data json.RawMessage) error {
	return json.Unmarshal(data, t)
}

var _ model.Entity = (*TestResourceA)(nil)

// TestResourceB represents another test resource
type TestResourceB struct{}

func (t *TestResourceB) Example() []byte {
	return []byte(`
	{
		"y": "example"
	}
	`)
}

func (t *TestResourceB) Marshal() (json.RawMessage, error) {
	return json.Marshal(t)
}

func (t *TestResourceB) Name() string {
	return "TestResourceB"
}

func (t *TestResourceB) Schema() []byte {
	return []byte(`
	{
		"type":"object",
		"properties": {
			"y": {"type":"string"}
		}
	}
	`)
}

func (t *TestResourceB) Unmarshal(data json.RawMessage) error {
	return json.Unmarshal(data, t)
}

var _ model.Entity = (*TestResourceB)(nil)

// =============================================================================
// ResourceWithMissingRef has a schema that references a non-existent resource
var _ model.Entity = (*ResourceWithMissingRef)(nil)

type ResourceWithMissingRef struct{}

func (t *ResourceWithMissingRef) Example() []byte {
	return []byte(`
	{
		"z": "example"
	}
	`)
}

func (t *ResourceWithMissingRef) Marshal() (json.RawMessage, error) {
	return json.Marshal(t)
}

func (t *ResourceWithMissingRef) Name() string {
	return "ResourceWithMissingRef"
}

func (t *ResourceWithMissingRef) Schema() []byte {
	return []byte(`
	{
		"type":"object",
		"properties": {
			"z": {"type":"string"},
			"missingRef": {"$ref": "#/definitions/TestResourceC"}
		}
	}
	`)
}

func (t *ResourceWithMissingRef) Unmarshal(data json.RawMessage) error {
	return json.Unmarshal(data, t)
}

// =========================================================================
// TestResourceALowerCase is a duplicate of TestResourceA but with lowercase name
// This should cause a case-sensitivity conflict
var _ model.Entity = (*TestResourceALowerCase)(nil)

type TestResourceALowerCase struct{}

func (t *TestResourceALowerCase) Example() []byte {
	return []byte(`
	{
		"x": "lowercase example",
		"y": {
			"y": "example"
		}
	}
	`)
}

func (t *TestResourceALowerCase) Marshal() (json.RawMessage, error) {
	return json.Marshal(t)
}

func (t *TestResourceALowerCase) Name() string {
	return "testresourcea" // Note: lowercase name conflicts with TestResourceA
}

func (t *TestResourceALowerCase) Schema() []byte {
	return []byte(`
	{
		"type":"object",
		"properties": {
			"x": {"type":"string"},
			"y": {"$ref": "#/definitions/TestResourceB"}
		}
	}
	`)
}

func (t *TestResourceALowerCase) Unmarshal(data json.RawMessage) error {
	return json.Unmarshal(data, t)
}

// =========================================================================
// TestResourceAClone is an identical clone of TestResourceA
// This should be fine because the schema is identical
var _ model.Entity = (*TestResourceAClone)(nil)

type TestResourceAClone struct{}

func (t *TestResourceAClone) Example() []byte {
	return []byte(`
	{
		"x": "example",
		"y": {
			"y": "example"
		}
	}
	`)
}

func (t *TestResourceAClone) Marshal() (json.RawMessage, error) {
	return json.Marshal(t)
}

func (t *TestResourceAClone) Name() string {
	return "TestResourceA" // Same name as TestResourceA
}

func (t *TestResourceAClone) Schema() []byte {
	return []byte(`
	{
		"type":"object",
		"properties": {
			"x": {"type":"string"},
			"y": {"$ref": "#/definitions/TestResourceB"}
		}
	}
	`)
}

func (t *TestResourceAClone) Unmarshal(data json.RawMessage) error {
	return json.Unmarshal(data, t)
}

// =========================================================================
// TestResourceAConflicting uses the same name as TestResourceA but has a different schema
// This should cause a schema conflict error
var _ model.Entity = (*TestResourceAConflicting)(nil)

type TestResourceAConflicting struct{}

func (t *TestResourceAConflicting) Example() []byte {
	return []byte(`
	{
		"x": "example",
		"z": 123
	}
	`)
}

func (t *TestResourceAConflicting) Marshal() (json.RawMessage, error) {
	return json.Marshal(t)
}

func (t *TestResourceAConflicting) Name() string {
	return "TestResourceA" // Same name as TestResourceA
}

func (t *TestResourceAConflicting) Schema() []byte {
	return []byte(`
	{
		"type":"object",
		"properties": {
			"x": {"type":"string"},
			"z": {"type":"integer"} 
		}
	}
	`)
}

func (t *TestResourceAConflicting) Unmarshal(data json.RawMessage) error {
	return json.Unmarshal(data, t)
}
