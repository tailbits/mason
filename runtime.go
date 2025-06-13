package mason

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/magicbell/mason/model"
)

type RouteHandler interface {
	Handle(method string, path string, handler WebHandler, mws ...func(WebHandler) WebHandler)
}

type WebResponder interface {
	Respond(ctx context.Context, w http.ResponseWriter, data any, status int) error
}

type Runtime interface {
	RouteHandler
	WebResponder
}

// ==========================================================================
// HTTPRuntime is a concrete implementation of the Runtime interface for HTTP-based applications.

var _ Runtime = (*HTTPRuntime)(nil)

type HTTPRuntime struct {
	*http.ServeMux
}

func (r *HTTPRuntime) Handle(method string, path string, handler WebHandler, mws ...func(WebHandler) WebHandler) {
	r.HandleFunc(fmt.Sprintf("%s %s", method, path), func(w http.ResponseWriter, req *http.Request) {
		if req.Method != method {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		ctx := req.Context()
		if err := handler(ctx, w, req); err != nil {
			var fe model.ValidationError
			if errors.As(err, &fe) {
				// Return well-formatted validation errors
				if err := r.Respond(ctx, w, fe, http.StatusUnprocessableEntity); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}

				return
			}

			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}

func (r *HTTPRuntime) Respond(ctx context.Context, w http.ResponseWriter, data any, status int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusInternalServerError)
			return fmt.Errorf("failed to encode response data: %w", err)
		}
	}

	return nil
}

func NewHTTPRuntime() *HTTPRuntime {
	return &HTTPRuntime{
		ServeMux: http.NewServeMux(),
	}
}
