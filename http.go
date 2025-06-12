package mason

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/magicbell/mason/model"
)

type HandlerWithBody[T model.Entity, O model.Entity, Q any] func(ctx context.Context, r *http.Request, model T, params Q) (response O, err error)
type HandlerNoBody[O model.Entity, Q any] func(ctx context.Context, r *http.Request, params Q) (response O, err error)

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

func newHandlerWithBody[T model.Entity, O model.Entity, Q any](api *API, fn HandlerWithBody[T, O, Q], code int) WebHandler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		params, err := DecodeQueryParams[Q](r)
		if err != nil {
			return fmt.Errorf("decodeQueryParams: %w", err)
		}

		model, err := DecodeRequest[T](api, r)
		if err != nil {
			return fmt.Errorf("validateAndDecode: %w", err)
		}

		result, err := fn(ctx, r, model, params)
		if err != nil {
			return err
		}

		return api.Respond(ctx, w, result, code)
	}
}

func newHandler[T model.Entity, Q any](rsp WebResponder, fn HandlerNoBody[T, Q], code int) WebHandler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		params, err := DecodeQueryParams[Q](r)
		if err != nil {
			return fmt.Errorf("decodeQueryParams: %w", err)
		}

		result, err := fn(ctx, r, params)
		if err != nil {
			return err
		}

		return rsp.Respond(ctx, w, result, code)
	}
}

func DecodeQueryParams[Q any](r *http.Request) (Q, error) {
	var params Q

	if err := r.ParseForm(); err != nil {
		return params, fmt.Errorf("unable to parse query params: %w", err)
	}

	// loop through fields of params
	v := reflect.TypeOf(params)
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		tag := field.Tag.Get("json")
		tag = strings.Split(tag, ",")[0]
		if tag == "" {
			continue
		}

		value := r.Form.Get(tag)
		defaultValue := field.Tag.Get("default")

		if value == "" && defaultValue != "" {
			value = defaultValue
		}

		if value == "" {
			continue
		}

		// set the value of the field
		f := reflect.ValueOf(&params).Elem().Field(i)

		kind := field.Type.Kind()

		switch kind {
		case reflect.String:
			f.SetString(value)
		case reflect.Int:
			n, err := strconv.Atoi(value)
			if err != nil {
				return params, fmt.Errorf("unable to parse query params: %w", err)
			}
			f.SetInt(int64(n))
		case reflect.Bool:
			b, err := strconv.ParseBool(value)
			if err != nil {
				return params, fmt.Errorf("unable to parse query params: %w", err)
			}
			f.SetBool(b)
		case reflect.Ptr:
			switch field.Type.Elem().Kind() {
			case reflect.String:
				f.Set(reflect.ValueOf(&value))
			case reflect.Int:
				n, err := strconv.Atoi(value)
				if err != nil {
					return params, fmt.Errorf("unable to parse query params: %w", err)
				}
				f.Set(reflect.ValueOf(&n))
			case reflect.Bool:
				b, err := strconv.ParseBool(value)
				if err != nil {
					return params, fmt.Errorf("unable to parse query params: %w", err)
				}
				f.Set(reflect.ValueOf(&b))
			}
		default:
			return params, fmt.Errorf("unsupported query param type: %v", f.Kind())
		}
	}

	return params, nil
}
