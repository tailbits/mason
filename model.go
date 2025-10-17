package mason

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/swaggest/jsonschema-go"
	"github.com/tailbits/mason/model"
)

var _ jsonschema.Exposer = (*Model)(nil)

type Model struct {
	jsonschema.Struct
	model.WithSchema
}

func (m Model) IsNil() bool {
	_, isNil := m.WithSchema.(model.Nil)
	return isNil
}

func (m Model) JSONSchema() (jsonschema.Schema, error) {
	if m.Schema() == nil {
		return jsonschema.Schema{}, nil
	}

	var sch jsonschema.Schema
	if err := json.Unmarshal(m.Schema(), &sch); err != nil {
		return jsonschema.Schema{}, fmt.Errorf("error unmarshalling schema for %s: %w", m.Name(), err)
	}

	ex := make(map[string]interface{})
	if err := json.Unmarshal(m.Example(), &ex); err != nil {
		return jsonschema.Schema{}, fmt.Errorf("error unmarshalling example for %s : %w", m.Name(), err)
	}
	sch.WithExamples(ex)

	walkRefs(&sch, func(ref *string) {
		refID := strings.ReplaceAll(*ref, "#/definitions/", "#/components/schemas/")
		refID = strings.TrimPrefix(refID, "#/components/schemas/")

		*ref = "#/components/schemas/" + refID
	})

	return sch, nil
}

func NewModel(ent model.WithSchema) Model {
	m := Model{
		Struct: jsonschema.Struct{
			DefName: ent.Name(),
		},
		WithSchema: ent,
	}

	return m
}

// =============================================================================
func walkRefs(schema *jsonschema.Schema, f func(*string)) {
	if schema == nil {
		return
	}

	apply := func(sch *jsonschema.Schema) {
		if sch.Ref != nil {
			f(sch.Ref)
		}
	}

	walkSchema(&jsonschema.SchemaOrBool{TypeObject: schema}, apply)
	for _, def := range schema.Definitions {
		walkSchema(&def, apply)
	}
}

func walkSchema(schemaOrBool *jsonschema.SchemaOrBool, f func(*jsonschema.Schema)) {
	if schemaOrBool == nil || schemaOrBool.TypeObject == nil {
		return
	}

	schema := schemaOrBool.TypeObject

	f(schema)

	walkSchema(schema.AdditionalItems, f)

	if schema.Items != nil && schema.Items.SchemaArray != nil {
		items := schema.Items.SchemaArray
		for _, item := range items {
			walkSchema(&item, f)
		}
	}

	if schema.Items != nil && schema.Items.SchemaOrBool != nil {
		walkSchema(schema.Items.SchemaOrBool, f)
	}

	walkSchema(schema.Contains, f)

	walkSchema(schema.AdditionalProperties, f)

	for _, prop := range schema.Properties {
		walkSchema(&prop, f)
	}

	if schema.AllOf != nil {
		for _, s := range schema.AllOf {
			walkSchema(&s, f)
		}
	}

	if schema.AnyOf != nil {
		for _, s := range schema.AnyOf {
			walkSchema(&s, f)
		}
	}

	if schema.OneOf != nil {
		for _, s := range schema.OneOf {
			walkSchema(&s, f)
		}
	}

	walkSchema(schema.Not, f)
}
