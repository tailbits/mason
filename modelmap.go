package mason

import (
	"github.com/magicbell/mason/model"
)

type Resource map[string]Operation
type Registry map[string]Resource

func (a *API) registerOp(m Operation, group string) {
	path := m.Path
	method := m.Method

	if grp, ok := (*&a.registry)[group]; ok {
		grp[toKey(method, path)] = m

		return
	}

	rsc := Resource{
		toKey(method, path): m,
	}
	(a.registry)[group] = rsc
}

// TaggedOps returns all models that have all the tags provided
func (mgm *Registry) TaggedOps(tags ...string) []Operation {
	models := make([]Operation, 0, len(*mgm)*2)

	for _, grp := range *mgm {
		for _, model := range grp {
			if len(model.Tags) < len(tags) {
				continue
			}

			hasAllTags := true
			for _, requiredTag := range tags {
				found := false
				for _, modelTag := range model.Tags {
					if modelTag == requiredTag {
						found = true
						break
					}
				}
				if !found {
					hasAllTags = false
					break
				}
			}

			if hasAllTags {
				models = append(models, model)
			}
		}
	}
	return models
}

func (mgm *Registry) FindOp(method string, path string) (Operation, bool) {
	for _, modelGroup := range *mgm {
		if model, ok := modelGroup[toKey(method, path)]; ok {
			return model, true
		}
	}
	return Operation{}, false
}

func (mgm *Registry) Ops() []Operation {
	var models []Operation
	for _, modelGroup := range *mgm {
		for _, model := range modelGroup {
			models = append(models, model)
		}
	}
	return models
}

func (mgm *Registry) Endpoints(transform func(string) string) []string {
	unique := make(map[string]bool)

	for _, modelGroup := range *mgm {
		for key := range modelGroup {
			unique[transform(key)] = true
		}
	}

	keys := make([]string, 0, len(unique))
	for key := range unique {
		keys = append(keys, key)
	}

	return keys
}

func (mg *Resource) FirstOp() (Operation, bool) {
	for _, model := range *mg {
		return model, true
	}
	return Operation{}, false
}

type Operation struct {
	OperationID string                 `json:"operationID,omitempty"`
	Input       model.Entity           `json:"input,omitempty"`
	Output      model.Entity           `json:"output,omitempty"`
	Method      string                 `json:"method,omitempty"`
	Path        string                 `json:"path,omitempty"`
	QueryParams any                    `json:"queryParams,omitempty"`
	Description string                 `json:"description,omitempty"`
	Summary     string                 `json:"summary,omitempty"`
	SuccessCode int                    `json:"code,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Extensions  map[string]interface{} `json:"mapOfAnything,omitempty"`
}

type Option func(*Operation)
type AuthType string

func WithOperationID(opID string) Option {
	return func(m *Operation) {
		m.OperationID = opID
	}
}

func WithDescription(desc string) Option {
	return func(m *Operation) {
		m.Description = desc
	}
}

func WithSuccessCode(code int) Option {
	return func(m *Operation) {
		m.SuccessCode = code
	}
}

func WithSummary(summary string) Option {
	return func(m *Operation) {
		m.Summary = summary
	}
}

func WithTags(tags ...string) Option {
	return func(m *Operation) {
		nonEmptyTags := make([]string, 0)
		for _, tag := range tags {
			if tag != "" {
				nonEmptyTags = append(nonEmptyTags, tag)
			}
		}
		m.Tags = nonEmptyTags
	}
}

func WithExtension(val map[string]interface{}) Option {
	return func(m *Operation) {
		m.Extensions = val
	}
}
