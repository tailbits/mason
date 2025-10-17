package mason_test

import (
	"testing"

	"github.com/tailbits/mason"
	"gotest.tools/assert"
	"gotest.tools/assert/cmp"
)

var _ mason.Builder = &MockBuilder{}

type MockEntity struct {
	name string
}

func (m *MockEntity) Name() string {
	return m.name
}

type MockBuilder struct {
	resourceID  string
	operationID string
	groupPath   string
}

func (m *MockBuilder) ResourceID() string {
	return m.resourceID
}

func (m *MockBuilder) OpID() string {
	return m.operationID
}

func (m *MockBuilder) WithGroup(path string) mason.Builder {
	m.groupPath = path
	return m
}

func (m *MockBuilder) Register(api *mason.API) {}

func TestGroup_New(t *testing.T) {
	entity := &MockEntity{
		name: "test",
	}

	api := mason.NewAPI(mason.NewHTTPRuntime())
	group := api.NewRouteGroup(entity.Name())

	assert.Assert(t, group != nil, "Named group creation should succeed")
}

func TestGroup_Nesting(t *testing.T) {
	t.Run("nests groups correctly", func(t *testing.T) {
		// Create parent group with a named entity
		api := mason.NewAPI(mason.NewHTTPRuntime())
		parentEntity := &MockEntity{name: "parent"}
		parent := api.NewRouteGroup(parentEntity.Name())

		// Create child group with its own entity
		childEntity := &MockEntity{name: "child"}
		child := parent.NewRouteGroup(childEntity.Name())

		assert.Assert(t, child != nil, "Child group creation should succeed")
	})

	t.Run("creates nested group path", func(t *testing.T) {
		api := mason.NewAPI(mason.NewHTTPRuntime())
		// Setup our test environment
		parentEntity := &MockEntity{name: "parent"}
		childEntity := &MockEntity{name: "child"}

		// Create our nested group structure
		parent := api.NewRouteGroup(parentEntity.Name())
		child := parent.NewRouteGroup(childEntity.Name())

		// Create a builder to capture the group path
		builder := &MockBuilder{
			resourceID:  "test-resource",
			operationID: "test-operation",
		}

		child.Register(builder)

		// Verify the group path contains both parent and child names
		assert.Assert(t, cmp.Contains(builder.groupPath, "parent"))
		assert.Assert(t, cmp.Contains(builder.groupPath, "child"))
	})
}

func TestGroupRegistration(t *testing.T) {
	entity := &MockEntity{name: "test-resource"}

	api := mason.NewAPI(mason.NewHTTPRuntime())
	grp := api.NewRouteGroup(entity.Name())

	builder := &MockBuilder{
		resourceID:  "test-resource",
		operationID: "test-operation",
	}

	// Verify registration doesn't panic
	assert.Assert(t, cmp.Nil(func() error {
		grp.Register(builder)
		return nil
	}()))
}

func TestSkipRESTValidation(t *testing.T) {
	t.Run("allows different resources with skip validation", func(t *testing.T) {
		api := mason.NewAPI(mason.NewHTTPRuntime())

		entity := &MockEntity{name: "test"}

		group := api.NewRouteGroup(entity.Name())

		// Skip REST validation with explicit group name
		group = group.SkipRESTValidation("custom-group")

		// First registration
		bldr_1 := &MockBuilder{
			resourceID:  "resource1",
			operationID: "op1",
		}
		group.Register(bldr_1)

		// Second registration with different resource should not panic
		bldr_2 := &MockBuilder{
			resourceID:  "resource2",
			operationID: "op2",
		}

		// Verify no panic occurs
		assert.Assert(t, cmp.Nil(func() error {
			group.Register(bldr_2)
			return nil
		}()))
	})
}

// Path implements apiv2.Builder.
func (m *MockBuilder) Path(p string) mason.Builder {
	panic("unimplemented")
}

// RegisterBeta implements apiv2.Builder.
func (m *MockBuilder) RegisterBeta(api *mason.API) {
	m.Register(api)
}

// SkipIf implements apiv2.Builder.
func (m *MockBuilder) SkipIf(skip bool) mason.Builder {
	panic("unimplemented")
}

// WithDesc implements apiv2.Builder.
func (m *MockBuilder) WithDesc(d string) mason.Builder {
	panic("unimplemented")
}

// WithMapOfAnything implements apiv2.Builder.
func (m *MockBuilder) WithMapOfAnything(val map[string]interface{}) mason.Builder {
	panic("unimplemented")
}

// WithExtensions implements apiv2.Builder.
func (m *MockBuilder) WithExtensions(key string, val interface{}) mason.Builder {
	panic("unimplemented")
}

// WithMWs implements apiv2.Builder.
func (m *MockBuilder) WithMWs(mw ...mason.Middleware) mason.Builder {
	panic("unimplemented")
}

// WithOpID implements apiv2.Builder.
func (m *MockBuilder) WithOpID(id ...string) mason.Builder {
	panic("unimplemented")
}

// WithSuccessCode implements apiv2.Builder.
func (m *MockBuilder) WithSuccessCode(code int) mason.Builder {
	panic("unimplemented")
}

// WithSummary implements apiv2.Builder.
func (m *MockBuilder) WithSummary(s string) mason.Builder {
	panic("unimplemented")
}

// WithTags implements apiv2.Builder.
func (m *MockBuilder) WithTags(tags ...string) mason.Builder {
	panic("unimplemented")
}
