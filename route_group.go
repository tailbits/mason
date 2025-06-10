package mason

import (
	"path"

	"github.com/magicbell/mason/internal/casing"
)

type RouteGroup struct {
	name         string
	rtm          *API
	parent       *RouteGroup
	skipValidate bool
}

func (g *RouteGroup) Name() string {
	return g.name
}

func (g *RouteGroup) FullPath() string {
	if g.name == "" {
		return ""
	}

	pth := casing.ToKebabCase(g.name)
	for p := g.parent; p != nil; p = p.parent {
		pth = path.Join(casing.ToKebabCase(p.name), pth)
	}

	return pth
}

func (g *RouteGroup) Register(builder Builder) {
	builder.WithGroup(g.FullPath()).Register(g.rtm)
}

// SkipRESTValidation relaxes the constraint that all routes in a group must handle the same resource.
func (g *RouteGroup) SkipRESTValidation(name string) *RouteGroup {
	if name == "" {
		panic("cannot skip rest validation without an explicit group name")
	}

	g.name = name

	g.skipValidate = true

	g.rtm.routeIndex[g.name] = g.name

	return g
}

func (g *RouteGroup) NewRouteGroup(name string) *RouteGroup {
	return &RouteGroup{
		name:   name,
		rtm:    g.rtm,
		parent: g,
	}
}
