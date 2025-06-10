package model

import (
	"encoding/json"
)

// Serializable is an interface for serializing and deserializing data.
// Serializble is used for data types sent and received over the wire.
type Serializable interface {
	Marshal() (json.RawMessage, error)
	Unmarshal(data json.RawMessage) error
}

// WithName is an interface for defining a name for a data type.
// The name is used for generating API documentation and for uniquely identifying the data type.
type WithName interface {
	Name() string
}

// WithSchema is an interface for defining a schema and example data for a data type.
// The schema is used for validating data and for generating API documentation, along with the example data.
type WithSchema interface {
	WithName
	Schema() []byte
	Example() []byte
}

// Entity is a domain model that can be serialized and has a schema.
// The term comes from Domain-Driven Design (DDD).
type Entity interface {
	WithSchema
	Serializable
}
