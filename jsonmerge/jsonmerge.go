package jsonmerge

import (
	"encoding/json"
	"fmt"
	"sort"
)

// orderedMap maintains insertion order of keys for JSON marshaling
type orderedMap struct {
	keys   []string
	values map[string]interface{}
}

func newOrderedMap() *orderedMap {
	return &orderedMap{
		keys:   make([]string, 0),
		values: make(map[string]interface{}),
	}
}

func (o *orderedMap) Set(key string, value interface{}) {
	if _, exists := o.values[key]; !exists {
		o.keys = append(o.keys, key)
	}
	o.values[key] = value
}

func (o *orderedMap) Get(key string) (interface{}, bool) {
	v, ok := o.values[key]
	return v, ok
}

func (o *orderedMap) MarshalJSON() ([]byte, error) {
	// Sort keys for deterministic output
	sort.Strings(o.keys)

	buf := []byte{'{'}
	for i, k := range o.keys {
		if i > 0 {
			buf = append(buf, ',')
		}
		// Marshal key
		key, err := json.Marshal(k)
		if err != nil {
			return nil, err
		}
		buf = append(buf, key...)
		buf = append(buf, ':')

		// Marshal value
		val, err := json.Marshal(o.values[k])
		if err != nil {
			return nil, err
		}
		buf = append(buf, val...)
	}
	buf = append(buf, '}')
	return buf, nil
}

// convertToOrderedMap recursively converts map[string]interface{} to orderedMap
func convertToOrderedMap(v interface{}) interface{} {
	switch v := v.(type) {
	case map[string]interface{}:
		om := newOrderedMap()
		for k, val := range v {
			om.Set(k, convertToOrderedMap(val))
		}
		return om
	case []interface{}:
		for i, val := range v {
			v[i] = convertToOrderedMap(val)
		}
	}
	return v
}

// Merger interface remains the same
type Merger interface {
	MergeSchemas(schemas ...[]byte) ([]byte, error)
	MergeExamples(examples ...[]byte) ([]byte, error)
}

type Options struct {
	SchemasMergeStrategy SchemaMergeStrategy
}

type SchemaMergeStrategy int

const (
	OverwriteDuplicates SchemaMergeStrategy = iota
	ErrorOnDuplicates
	KeepExisting
)

func New() Merger {
	return NewWithOptions(Options{
		SchemasMergeStrategy: OverwriteDuplicates,
	})
}

func NewWithOptions(opts Options) Merger {
	return &merger{opts: opts}
}

type merger struct {
	opts Options
}

func (m *merger) MergeSchemas(schemas ...[]byte) ([]byte, error) {
	if len(schemas) == 0 {
		return []byte("{}"), nil
	}

	result := newOrderedMap()
	result.Set("type", "object")

	properties := newOrderedMap()
	result.Set("properties", properties)

	required := make([]string, 0)

	for _, schema := range schemas {
		var current map[string]interface{}
		if err := json.Unmarshal(schema, &current); err != nil {
			return nil, fmt.Errorf("failed to unmarshal schema: %w", err)
		}

		if err := m.mergeSchemaProperties(properties, current); err != nil {
			return nil, err
		}

		if err := m.mergeRequiredFields(&required, current); err != nil {
			return nil, err
		}
	}

	if len(required) > 0 {
		sort.Strings(required)
		result.Set("required", required)
	}

	return json.MarshalIndent(result, "", "  ")
}

func (m *merger) MergeExamples(examples ...[]byte) ([]byte, error) {
	if len(examples) == 0 {
		return []byte("{}"), nil
	}

	result := newOrderedMap()
	for _, example := range examples {
		var current map[string]interface{}
		if err := json.Unmarshal(example, &current); err != nil {
			return nil, fmt.Errorf("failed to unmarshal example: %w", err)
		}

		for k, v := range current {
			result.Set(k, convertToOrderedMap(v))
		}
	}

	return json.MarshalIndent(result, "", "  ")
}

func (m *merger) mergeSchemaProperties(properties *orderedMap, schema map[string]interface{}) error {
	if props, ok := schema["properties"].(map[string]interface{}); ok {
		for k, v := range props {
			if _, exists := properties.Get(k); exists {
				switch m.opts.SchemasMergeStrategy {
				case ErrorOnDuplicates:
					return fmt.Errorf("duplicate property found: %s", k)
				case KeepExisting:
					continue
				}
			}
			properties.Set(k, convertToOrderedMap(v))
		}
	}
	return nil
}

func (m *merger) mergeRequiredFields(required *[]string, schema map[string]interface{}) error {
	if req, ok := schema["required"].([]interface{}); ok {
		reqMap := make(map[string]bool)

		// Add existing required fields
		for _, r := range *required {
			reqMap[r] = true
		}

		// Add new required fields
		for _, r := range req {
			if str, ok := r.(string); ok {
				reqMap[str] = true
			}
		}

		// Convert back to sorted slice
		*required = make([]string, 0, len(reqMap))
		for k := range reqMap {
			*required = append(*required, k)
		}
	}
	return nil
}
