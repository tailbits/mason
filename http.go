package mason

import (
	"net/http"

	"github.com/magicbell/mason/model"
)

func HandlePost[T model.Entity, O model.Entity, Q any](handler HandlerWithBody[T, O, Q]) *RouteBuilderWithBody[T, O, Q] {
	return &RouteBuilderWithBody[T, O, Q]{
		RouteBuilderBase: RouteBuilderBase{
			method:  http.MethodPost,
			keyVals: make(map[string]interface{}),
		},
		handler: handler,
	}
}

func HandlePut[T model.Entity, O model.Entity, Q any](handler HandlerWithBody[T, O, Q]) *RouteBuilderWithBody[T, O, Q] {
	return &RouteBuilderWithBody[T, O, Q]{
		RouteBuilderBase: RouteBuilderBase{
			method:  http.MethodPut,
			keyVals: make(map[string]interface{}),
		},
		handler: handler,
	}
}

func HandlePatch[T model.Entity, O model.Entity, Q any](handler HandlerWithBody[T, O, Q]) *RouteBuilderWithBody[T, O, Q] {
	return &RouteBuilderWithBody[T, O, Q]{
		RouteBuilderBase: RouteBuilderBase{
			method:  http.MethodPatch,
			keyVals: make(map[string]interface{}),
		},
		handler: handler,
	}
}

func HandleDelete[T model.Entity, O model.Entity, Q any](handler HandlerWithBody[T, O, Q]) *RouteBuilderWithBody[T, O, Q] {
	return &RouteBuilderWithBody[T, O, Q]{
		RouteBuilderBase: RouteBuilderBase{
			method:  http.MethodDelete,
			keyVals: make(map[string]interface{}),
		},
		handler: handler,
	}
}

func HandleGet[T model.Entity, Q any](handler HandlerNoBody[T, Q]) *RouteBuilderNoBody[T, Q] {
	return &RouteBuilderNoBody[T, Q]{
		RouteBuilderBase: RouteBuilderBase{
			method:  http.MethodGet,
			keyVals: make(map[string]interface{}),
		},
		handler: handler,
	}
}
