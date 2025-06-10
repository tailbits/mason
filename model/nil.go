package model

import (
	"encoding/json"
)

// Nil is a model that represents an empty object.
type Nil struct{}

var _ Entity = (*Nil)(nil)

func (n Nil) Schema() []byte {
	return []byte(`
		{
			"type":"object",
			"properties":{}, 
			"additionalProperties": false, 
			"required": []
		}`,
	)
}

func (n Nil) Example() []byte {
	return []byte(`{}`)
}

func (n Nil) Name() string {
	return "NilEntity"
}

func (n Nil) Marshal() (json.RawMessage, error) {
	return []byte("{}"), nil
}

func (n Nil) Unmarshal(data json.RawMessage) error {
	return nil
}
