package mason

import (
	"fmt"
	"net/http"
	"path"
	"strings"

	m "github.com/magicbell/mason/model"
)

var successCodes = map[string]int{
	http.MethodPost:   http.StatusCreated,
	http.MethodPut:    http.StatusOK,
	http.MethodPatch:  http.StatusCreated,
	http.MethodDelete: http.StatusOK,
	http.MethodGet:    http.StatusOK,
}

type Builder interface {
	ResourceID() string
	OpID() string
	Path(p string) Builder
	WithGroup(group string) Builder
	WithOpID(segments ...string) Builder
	WithDesc(d string) Builder
	WithTags(tags ...string) Builder
	WithSuccessCode(code int) Builder
	WithSummary(s string) Builder
	WithMWs(mw ...Middleware) Builder
	WithExtensions(key string, val interface{}) Builder
	SkipIf(skip bool) Builder
	RegisterBeta(api *API)
	Register(api *API)
}

type RouteBuilderBase struct {
	opID        string
	method      string
	path        string
	mw          []func(WebHandler) WebHandler
	desc        string
	tags        []string
	summary     string
	successCode int
	skipped     bool
	group       string
	keyVals     map[string]interface{}
}

func (rb *RouteBuilderBase) validate() error {
	if rb.opID == "" {
		return fmt.Errorf("operationID is required")
	}
	if rb.method == "" {
		return fmt.Errorf("method is required")
	}
	if rb.path == "" {
		return fmt.Errorf("path is required")
	}
	return nil
}

type RouteBuilderWithBody[T m.Entity, O m.Entity, Q any] struct {
	RouteBuilderBase
	handler HandlerWithBody[T, O, Q]
}

// ResourceID returns the resource ID for the route.
func (rb *RouteBuilderWithBody[T, O, Q]) ResourceID() string {
	t := m.New[T]()

	return RecursivelyUnwrap(t).Name()
}

// Path sets the path for the route. This can include path parameters like /users/{id}
func (rb *RouteBuilderWithBody[T, O, Q]) Path(p string) Builder {
	rb.path = p

	return rb
}

func (rb *RouteBuilderWithBody[T, O, Q]) WithGroup(group string) Builder {
	rb.group = group

	return rb
}

// WithOpID sets the operationID for the route. This is used primarily for documentation purposes.
func (rb *RouteBuilderWithBody[T, O, Q]) WithOpID(id ...string) Builder {
	rb.opID = strings.ReplaceAll(path.Join(id...), "/", "_")
	return rb
}

// OpID returns the operation ID for the route.
func (rb *RouteBuilderWithBody[T, O, Q]) OpID() string {
	return rb.opID
}

// WithDesc sets the description for the route. This is used primarily for documentation purposes.
func (rb *RouteBuilderWithBody[T, O, Q]) WithDesc(d string) Builder {
	rb.desc = d
	return rb
}

// WithTags sets the tags for the route. This is used primarily for documentation purposes.
func (rb *RouteBuilderWithBody[T, O, Q]) WithTags(tags ...string) Builder {
	rb.tags = tags
	return rb
}

// WithExtensions sets custom x- attributes for the route. This is used for adding OpenAPI extensions..
func (rb *RouteBuilderWithBody[T, O, Q]) WithExtensions(key string, val interface{}) Builder {
	if !strings.HasPrefix(key, "x-") {
		panic(fmt.Errorf("custom keys must start with 'x-', key '%s' does not start with 'x-'", key))
	}
	rb.keyVals[key] = val

	return rb
}

// WithSuccessCode sets the success code for the route. This can be used to override the default success code for the method.
func (rb *RouteBuilderWithBody[T, O, Q]) WithSuccessCode(code int) Builder {
	rb.successCode = code
	return rb
}

func (rb *RouteBuilderWithBody[T, O, Q]) WithSummary(s string) Builder {
	rb.summary = s
	return rb
}

// WithMWs defines a set of middlewares to add to the route.
func (rb *RouteBuilderWithBody[T, O, Q]) WithMWs(mw ...Middleware) Builder {
	for _, m := range mw {
		h := m.GetHandler(rb)
		rb.mw = append(rb.mw, h)
	}

	return rb
}

// SkipIf ensures that the route is not documented if the condition is true.
func (rb *RouteBuilderWithBody[T, O, Q]) SkipIf(skip bool) Builder {
	rb.skipped = skip
	return rb
}

// RegisterBeta registers the route and marks it as beta, meaning it will not be included in the OpenAPI documentation.
func (rb *RouteBuilderWithBody[T, O, Q]) RegisterBeta(api *API) {
	rb.SkipIf(true).Register(api)
}

// Register registers the route with the mux, and finalizes the route configuration.
func (rb *RouteBuilderWithBody[T, O, Q]) Register(api *API) {
	if err := rb.validate(); err != nil {
		panic(err)
	}
	if rb.handler == nil {
		panic("handler is required")
	}
	if rb.group == "" {
		msg := fmt.Sprintf("route group name could not be inferred for %s %s; consider using group.WithDefaultName() to set it explicitly", rb.method, rb.path)
		panic(msg)
	}

	var output O
	if rb.successCode == 0 {
		rb.successCode = DefaultSuccessCode(rb.method, output)
	}

	if !rb.skipped {
		registerModel[T, O, Q](
			api,
			rb.method,
			rb.group,
			rb.path,
			WithOperationID(rb.opID),
			WithSuccessCode((rb.successCode)),
			WithDescription(rb.desc),
			WithSummary(rb.summary),
			WithTags(rb.tags...),
			WithExtension(rb.keyVals),
		)
	}

	h := newHandlerWithBody(api, rb.handler, rb.successCode)

	api.Handle(rb.method, rb.path, h, rb.mw...)
}

type RouteBuilderNoBody[T m.Entity, Q any] struct {
	RouteBuilderBase
	handler HandlerNoBody[T, Q]
}

func (rb *RouteBuilderNoBody[T, Q]) ResourceID() string {
	t := m.New[T]()

	return RecursivelyUnwrap(t).Name()
}

// Path sets the path for the route. This can include path parameters like /users/{id}
func (rb *RouteBuilderNoBody[T, Q]) Path(p string) Builder {
	rb.path = p
	return rb
}

func (rb *RouteBuilderNoBody[T, Q]) WithGroup(group string) Builder {
	rb.group = group
	return rb
}

// WithOpID sets the operationID for the route. This is used primarily for documentation purposes.
func (rb *RouteBuilderNoBody[T, Q]) WithOpID(id ...string) Builder {
	rb.opID = strings.ReplaceAll(path.Join(id...), "/", "_")
	return rb
}

// OpID returns the operation ID for the route.
func (rb *RouteBuilderNoBody[T, Q]) OpID() string {
	return rb.opID
}

// WithDesc sets the description for the route. This is used primarily for documentation purposes.
func (rb *RouteBuilderNoBody[T, Q]) WithDesc(d string) Builder {
	rb.desc = d
	return rb
}

// WithTags sets the tags for the route. This is used primarily for documentation purposes.
func (rb *RouteBuilderNoBody[T, Q]) WithTags(tags ...string) Builder {
	rb.tags = tags
	return rb
}

// WithExtensions sets custom x- attributes for the route. This is used for adding OpenAPI extensions..
func (rb *RouteBuilderNoBody[T, Q]) WithExtensions(key string, val interface{}) Builder {
	if !strings.HasPrefix(key, "x-") {
		panic(fmt.Errorf("invalid key [%s]: custom keys must start with 'x-'", key))
	}
	rb.keyVals[key] = val

	return rb
}

// WithSuccessCode sets the success code for the route. This can be used to override the default success code for the method.
func (rb *RouteBuilderNoBody[T, Q]) WithSuccessCode(code int) Builder {
	rb.successCode = code
	return rb
}

func (rb *RouteBuilderNoBody[T, Q]) WithSummary(s string) Builder {
	rb.summary = s
	return rb
}

// WithMWs defines a set of middlewares to add to the route.
func (rb *RouteBuilderNoBody[T, Q]) WithMWs(mw ...Middleware) Builder {
	for _, m := range mw {
		h := m.GetHandler(rb)
		rb.mw = append(rb.mw, h)
	}

	return rb
}

// SkipIf ensures that the route is not documented if the condition is true.
func (rb *RouteBuilderNoBody[T, Q]) SkipIf(skip bool) Builder {
	rb.skipped = skip
	return rb
}

// RegisterBeta registers the route and marks it as beta, meaning it will not be included in the OpenAPI documentation.
func (rb *RouteBuilderNoBody[T, Q]) RegisterBeta(api *API) {
	rb.SkipIf(true).Register(api)
}

// Register registers the route with the mux, and finalizes the route configuration.
func (rb *RouteBuilderNoBody[T, Q]) Register(api *API) {
	if err := rb.validate(); err != nil {
		panic(err)
	}
	if rb.handler == nil {
		panic("handler is required")
	}
	if rb.group == "" {
		panic("group is required")
	}

	var output T
	if rb.successCode == 0 {
		rb.successCode = DefaultSuccessCode(rb.method, output)
	}

	if !rb.skipped {
		registerResponseEntity[T, Q](
			api,
			rb.method,
			rb.group, rb.path,
			WithOperationID(rb.opID),
			WithSuccessCode((rb.successCode)),
			WithDescription(rb.desc),
			WithSummary(rb.summary),
			WithTags(rb.tags...),
			WithExtension(rb.keyVals),
		)
	}

	h := newHandler(api, rb.handler, rb.successCode)

	api.Handle(rb.method, rb.path, h, rb.mw...)
}

func DefaultSuccessCode(method string, output m.WithSchema) int {
	if _, ok := any(output).(m.Nil); ok {
		return http.StatusNoContent
	}
	return successCodes[method]
}

func RecursivelyUnwrap(current m.WithSchema) m.WithSchema {
	for {
		unwrapper, ok := current.(m.DerivedType)
		if !ok {
			return current
		}
		current = unwrapper.Unwrap()
	}
}
