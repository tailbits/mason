package openapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/daveshanley/vacuum/model"
	"github.com/daveshanley/vacuum/motor"
	"github.com/daveshanley/vacuum/rulesets"
	"github.com/magicbell/mason"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/swaggest/jsonschema-go"
	"github.com/swaggest/openapi-go/openapi31"
)

type definitionsMap map[string]jsonschema.Schema

type ReflectorWrapper struct {
	*openapi31.Reflector
	allDefs definitionsMap
	allTags map[string]bool
}

func (r *ReflectorWrapper) ingest(records []Record) error {
	for _, record := range records {
		ctx, err := r.newOperationContext(record.Method, record.Path)
		if err != nil {
			return fmt.Errorf("failed to create operation context: %w", err)
		}

		if err := ctx.from(record); err != nil {
			return fmt.Errorf("failed to populate operation context: %w", err)
		}

		if err := ctx.addToReflector(); err != nil {
			return fmt.Errorf("failed to add operation: %w", err)
		}
	}

	return nil
}

func (r *ReflectorWrapper) validate() error {
	specBytes, err := r.marshalJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// build and store built-in vacuum default RuleSets.
	defaultRS := rulesets.BuildDefaultRuleSets()

	// generate the 'recommended' RuleSet
	recommendedRS := defaultRS.GenerateOpenAPIRecommendedRuleSet()

	// apply the rules in the ruleset to the specification
	lintingResults := motor.ApplyRulesToRuleSet(
		&motor.RuleSetExecution{
			RuleSet: recommendedRS,
			Spec:    specBytes,
		})

	// create a new model.RuleResultSet from the results.
	// structure allows categorization, sorting and searching
	// in a simple and consistent way.
	resultSet := model.NewRuleResultSet(lintingResults.Results)

	// sort results by line number (so they are not all jumbled)
	resultSet.SortResultsByLineNumber()

	// .. do something interesting with the results
	// print only the results from the 'schemas' category
	schemasResults := resultSet.GetRuleResultsForCategory("schemas")

	errors := make([]error, 0)

	// for every rule that is violated, it contains a list of violations.
	// so first iterate through the schemas sesults
	for _, ruleResult := range schemasResults.RuleResults {

		// iterate over each violation of this rule
		for _, violation := range ruleResult.Results {

			errors = append(errors, fmt.Errorf(" - [%d:%d] %s", violation.StartNode.Line, violation.StartNode.Column, violation.Message))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation failed: %v", errors)
	}

	return nil
}

func (r *ReflectorWrapper) marshalJSON() ([]byte, error) {
	return r.Reflector.Spec.MarshalJSON()
}

// collectDefinitions takes all the definitions that have been collected in a cache from all the calls to AddReqStructure/AddRespStructure, and commits them to the reflector's OpenAPI spec.
func (r *ReflectorWrapper) collectDefinitions() error {
	// First, check for case-insensitive duplicates.
	seen := make(map[string]string) // map normalized key -> original key
	for defName := range r.allDefs {
		normalized := strings.ToLower(defName)
		if orig, exists := seen[normalized]; exists {
			return fmt.Errorf("conflicting definitions: %q and %q", orig, defName)
		}
		seen[normalized] = defName
	}

	for defName, def := range r.allDefs {
		if r.Reflector.Spec.Components == nil {
			return nil
		}
		def.Definitions = nil
		sm, err := def.ToSchemaOrBool().ToSimpleMap()
		if err != nil {
			continue
		}
		r.Reflector.Spec.Components.WithSchemasItem(defName, sm)
	}

	return nil
}

// collectTags takes a list of tags and saves them to the reflector so that they can be included in the final OpenAPI spec as a top-level key.
func (r *ReflectorWrapper) collectTags(tags []string) {
	r.Spec.Tags = make([]openapi31.Tag, len(tags))
	for i, tag := range tags {
		r.Spec.Tags[i] = openapi31.Tag{Name: tag}
	}
}

func (r *ReflectorWrapper) addModel(model mason.Model) error {
	if model.IsNil() {
		return nil
	}

	schema, err := model.JSONSchema()
	if err != nil {
		return fmt.Errorf("failed to get JSON schema: %w", err)
	}

	if err := r.addDefinition(model.Name(), schema); err != nil {
		return fmt.Errorf("failed to add definition: %w", err)
	}

	return nil
}

// addDefinition accepts a schema and a name, and adds the schema to the reflector so that it can be included in the final OpenAPI spec.
// If a definition with the same name already exists, it will be compared with the new definition to ensure they are identical, otherwise an error will be returned.
func (r *ReflectorWrapper) addDefinition(name string, schema jsonschema.Schema) error {
	if name == "" {
		return fmt.Errorf("definition name cannot be empty")
	}

	if existingDef, ok := r.allDefs[name]; ok {
		if !isSchemaIdentical(existingDef, schema) {
			return fmt.Errorf("definition with name [%s] already exists but with a different definition", name)
		}
		if len(existingDef.Examples) > 0 && len(schema.Examples) == 0 {
			return nil
		}
	}
	r.allDefs[name] = schema

	for nestedName, def := range schema.Definitions {
		if def.TypeObject != nil {
			if err := r.addDefinition(nestedName, *def.TypeObject); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *ReflectorWrapper) newOperationContext(method, path string) (*ContextWrapper, error) {
	oc, err := r.Reflector.NewOperationContext(method, path)
	if err != nil {
		return nil, fmt.Errorf("failed to create operation context: %w", err)
	}

	return NewContextWrapper(oc, r), nil
}

/* -------------------------------------------------------------------------- */

func printDiff(existingDef jsonschema.Schema, newDef jsonschema.Schema) {
	dmp := diffmatchpatch.New()

	existing, _ := existingDef.MarshalJSON()
	new, _ := newDef.MarshalJSON()

	diffs := dmp.DiffMain(string(pretty(existing)), string(pretty(new)), false)
	diffs = dmp.DiffCleanupSemantic(diffs)

	str := dmp.DiffPrettyText(diffs)

	fmt.Println(str) // nolint:forbidigo
}

func pretty(schema []byte) []byte {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, schema, "", "  "); err != nil {
		return schema // Return original if indentation fails
	}
	return prettyJSON.Bytes()
}

func isSchemaIdentical(a jsonschema.Schema, b jsonschema.Schema) bool {
	a.Examples = nil
	b.Examples = nil

	aa, _ := a.MarshalJSON()
	bb, _ := b.MarshalJSON()

	res := string(aa) == string(bb)
	if !res {
		printDiff(a, b)
	}
	return res
}
