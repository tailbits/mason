package mason

import (
	"strings"

	"github.com/magicbell/mason/model"
)

func (a *API) registerEntity(entity model.Entity) {
	a.entities[entity.Name()] = entity
}

func (a *API) GetEntity(name string) (model.Entity, bool) {
	e, ok := a.entities[name]

	return e, ok
}

func registerModel[I, O model.Entity, Q any](api *API, method string, group string, path string, opts ...Option) {
	i := model.New[I]()
	o := model.New[O]()
	q := model.New[Q]()

	m := Operation{
		Method:      method,
		Path:        path,
		Input:       i,
		Output:      o,
		QueryParams: q,
	}

	for _, opt := range opts {
		opt(&m)
	}

	api.registerEntity(i)
	api.registerEntity(o)

	api.registerOp(m, group)
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

	api.registerEntity(o)

	api.registerOp(m, group)
}

func (a *API) Registry() Registry {
	return a.registry
}

func (a *API) Operations() []Operation {
	return a.registry.Ops()
}

func (a *API) Routes() []string {
	return a.registry.Endpoints(func(key string) string {
		_, path := fromKey(key)
		return path
	})
}

func (a *API) GetOperation(method string, path string) (Operation, bool) {
	return a.registry.FindOp(method, path)
}

func (a *API) HasOperation(method string, path string) bool {
	_, ok := a.GetOperation(method, path)
	return ok
}

func toKey(method string, path string) string {
	return method + ":" + path
}

func fromKey(key string) (method string, path string) {
	split := strings.Split(key, ":")
	if len(split) != 2 {
		return "", ""
	}
	return split[0], split[1]
}
