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

func (a *API) NewRouteGroup(name string) *RouteGroup {
	return &RouteGroup{
		rtm:  a,
		name: name,
	}
}
