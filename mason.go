package mason

import (
	"context"
	"net/http"

	"github.com/magicbell/mason/model"
)

type groupMap map[string]string

type WebHandler func(ctx context.Context, w http.ResponseWriter, r *http.Request) error

type Middleware interface {
	GetHandler(builder Builder) func(WebHandler) WebHandler
}

type API struct {
	Runtime
	registry   Registry
	entities   map[string]model.Entity
	routeIndex groupMap
}

func NewAPI(runtime Runtime) *API {
	return &API{
		Runtime:    runtime,
		registry:   make(Registry),
		entities:   make(map[string]model.Entity),
		routeIndex: make(groupMap),
	}
}

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

func (a *API) NewRouteGroup(name string) *RouteGroup {
	return &RouteGroup{
		rtm:  a,
		name: name,
	}
}

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
