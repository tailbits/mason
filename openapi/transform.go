package openapi

import (
	"fmt"
	"sort"

	"github.com/swaggest/jsonschema-go"
	"github.com/swaggest/openapi-go/openapi31"
)

const prefix string = "v2"
const version string = "2.0.0"

var name string = "MagicBell"
var url string = "https://magicbell.com"
var email string = "hello@magicbell.com"
var serverDescription string = "MagicBell REST API Base URL"

func ToSchema(records []Record, cfg openapiConfig) ([]byte, error) {
	var reflector = setupReflector()

	if err := reflector.ingest(records); err != nil {
		return nil, fmt.Errorf("failed to ingest records: %w", err)
	}

	collectedTags := []string{}
	for tag := range reflector.allTags {
		collectedTags = append(collectedTags, tag)
	}
	for _, inferredTag := range cfg.allTags {
		if _, ok := reflector.allTags[inferredTag]; !ok {
			collectedTags = append(collectedTags, inferredTag)
		}
	}

	sort.Strings(collectedTags)
	reflector.collectTags(collectedTags)
	if err := reflector.collectDefinitions(); err != nil {
		return nil, fmt.Errorf("failed to collect definitions: %w", err)
	}

	if cfg.validate {
		return reflector.marshalJSON()
	}

	if err := reflector.validate(); err != nil {
		return nil, fmt.Errorf("failed to validate the generated spec: %w", err)
	}

	return reflector.marshalJSON()
}

func setupReflector() *ReflectorWrapper {
	reflector := openapi31.NewReflector()
	reflector.Spec = &openapi31.Spec{Openapi: "3.1.0"}
	reflector.Spec.Info.
		WithTitle("MagicBell API").
		WithVersion(version).
		WithDescription("OpenAPI 3.1.0 Specification for MagicBell API.").
		WithContact(openapi31.Contact{Name: &name, Email: &email, URL: &url})
	reflector.Spec.WithServers(openapi31.Server{
		URL:         "https://api.magicbell.com/" + prefix,
		Description: &serverDescription,
	})

	reflector.Reflector.DefaultOptions = append(reflector.Reflector.DefaultOptions, jsonschema.DefinitionsPrefix("#/components/schemas/"))

	return &ReflectorWrapper{reflector, make(definitionsMap), make(map[string]bool)}
}
