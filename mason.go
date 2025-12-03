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

type GroupMetadata struct {
	Summary     string
	Description string
}

type API struct {
	Runtime
	registry   Registry
	models     map[string]model.Entity
	routeIndex groupMap
	groupMeta  map[string]GroupMetadata
}

func NewAPI(runtime Runtime) *API {
	return &API{
		Runtime:    runtime,
		registry:   make(Registry),
		models:     make(map[string]model.Entity),
		routeIndex: make(groupMap),
		groupMeta:  make(map[string]GroupMetadata),
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

func (a *API) ForEachOperation(fn func(group string, op Operation)) {
	for group, resource := range a.registry {
		for _, op := range resource {
			fn(group, op)
		}
	}
}

func (a *API) GroupMetadata(path string) (GroupMetadata, bool) {
	meta, ok := a.groupMeta[path]
	return meta, ok
}

func (a *API) setGroupSummary(path string, summary string) {
	a.updateGroupMetadata(path, func(meta *GroupMetadata) {
		meta.Summary = summary
	})
}

func (a *API) setGroupDescription(path string, description string) {
	a.updateGroupMetadata(path, func(meta *GroupMetadata) {
		meta.Description = description
	})
}

func (a *API) updateGroupMetadata(path string, update func(*GroupMetadata)) {
	if path == "" || update == nil {
		return
	}

	meta := a.groupMeta[path]
	update(&meta)
	a.groupMeta[path] = meta
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
