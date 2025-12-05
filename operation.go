package mason

import "github.com/tailbits/mason/model"

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

func (a *API) registerOp(m Operation, group string) {
	path := m.Path
	method := m.Method

	if grp, ok := (a.registry)[group]; ok {
		grp[toKey(method, path)] = m

		return
	}

	rsc := Resource{
		toKey(method, path): m,
	}
	(a.registry)[group] = rsc
}

func registerResponseEntity[O model.Entity, Q any](api *API, method string, group string, path string, opts ...Option) {
	o := model.New[O]()
	q := model.New[Q]()

	m := Operation{
		Method:      method,
		Path:        path,
		Output:      o,
		QueryParams: q,
	}

	for _, opt := range opts {
		opt(&m)
	}

	api.registerModel(o)

	api.registerOp(m, group)
}
