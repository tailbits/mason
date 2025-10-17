package mason

import (
	"context"
	"net/http"

	"github.com/tailbits/mason/model"
)

type groupMap map[string]string

type WebHandler func(ctx context.Context, w http.ResponseWriter, r *http.Request) error

type Middleware interface {
	GetHandler(builder Builder) func(WebHandler) WebHandler
}

type API struct {
	Runtime
	registry   Registry
	models     map[string]model.Entity
	routeIndex groupMap
}

func NewAPI(runtime Runtime) *API {
	return &API{
		Runtime:    runtime,
		registry:   make(Registry),
		models:     make(map[string]model.Entity),
		routeIndex: make(groupMap),
	}
}

func (a *API) NewRouteGroup(name string) *RouteGroup {
	return &RouteGroup{
		rtm:  a,
		name: name,
	}
}

func (a *API) registerModel(mdl model.Entity) {
	a.models[mdl.Name()] = mdl
}

func (a *API) GetModel(name string) (model.Entity, bool) {
	e, ok := a.models[name]

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

	api.registerModel(i)
	api.registerModel(o)

	api.registerOp(m, group)
}
