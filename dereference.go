package mason

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/magicbell/mason/model"
	"github.com/swaggest/jsonschema-go"
)

func (a *API) DereferenceSchema(schema []byte) ([]byte, error) {
	var sch jsonschema.Schema
	if err := json.Unmarshal(schema, &sch); err != nil {
		return nil, fmt.Errorf("json.Unmarshal: schema[%s] %w", string(schema), err)
	}

	var err error
	walkRefs(&sch, func(ref *string) {
		id := strings.TrimPrefix(*ref, "#/definitions/")

		if _, ok := sch.Definitions[id]; !ok {
			// that means we have an external reference
			// we need to dereference it
			e, ok := a.GetEntity(id)
			if !ok {
				err = fmt.Errorf("entity %s not found", id)
				return
			}

			ent, ok := e.(model.WithSchema)
			if !ok {
				err = fmt.Errorf("entity %s does not implement platform.WithSchema", id)
				return
			}

			entSchBytes := ent.Schema()

			var entSch jsonschema.Schema
			if err = json.Unmarshal(entSchBytes, &entSch); err != nil {
				return
			}

			sch.WithDefinitionsItem(id, entSch.ToSchemaOrBool())
		}
	})

	if err != nil {
		return nil, err
	}

	return json.Marshal(sch)
}
