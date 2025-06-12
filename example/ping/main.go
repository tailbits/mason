package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/magicbell/mason"
	"github.com/magicbell/mason/model"
	"github.com/magicbell/mason/openapi"
)

var _ model.Entity = (*Response)(nil)

type Response struct {
	Timestamp time.Time `json:"timestamp"`
}

// Example implements model.Entity.
func (r *Response) Example() []byte {
	return []byte(`{
		"timestamp": "2023-10-01T12:00:00Z"
	}`)
}

func (r *Response) Marshal() (json.RawMessage, error) {
	return json.Marshal(r)
}

func (r *Response) Name() string {
	return "PingResponse"
}

func (r *Response) Schema() []byte {
	return []byte(`{
		"type": "object",
		"properties": {
			"timestamp": {
				"type": "string",
				"format": "date-time"
			}
		},
		"required": ["timestamp"]
	}`)
}

func (r *Response) Unmarshal(data json.RawMessage) error {
	return json.Unmarshal(data, r)
}

func PingHandler(ctx context.Context, r *http.Request, params model.Nil) (rsp *Response, err error) {
	return &Response{
		Timestamp: time.Now().UTC(),
	}, nil
}

func main() {
	rtm := mason.NewHTTPRuntime()

	api := mason.NewAPI(rtm)

	grp := api.NewRouteGroup("ping")

	grp.Register(mason.HandleGet(PingHandler).
		Path("/ping").
		WithOpID("ping").
		WithSummary("Ping the server").
		WithDesc("Ping the server when you are unsure of the time"))

	// let's generate the OpenAPI documentation
	gen, err := openapi.NewGenerator(api)
	if err != nil {
		panic(fmt.Errorf("failed to create OpenAPI generator: %w", err))
	}
	gen.Spec.Info.WithTitle("Ping API")

	schema, err := gen.Schema()
	if err != nil {
		panic(fmt.Errorf("failed to generate OpenAPI schema: %w", err))
	}

	// We can mix mason endpoints, with standard HTTP handlers
	rtm.Handle("GET", "/openapi.json", func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write(schema); err != nil {
			return fmt.Errorf("failed to write OpenAPI schema: %w", err)
		}

		return nil
	})

	server := &http.Server{
		Addr:    ":9090",
		Handler: rtm,
	}
	fmt.Println("API URL      : http://localhost:9090")
	fmt.Println("OpenAPI spec : http://localhost:9090/openapi.json")
	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}
