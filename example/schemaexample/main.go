package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/tailbits/mason"
	"github.com/tailbits/mason/model"
	"github.com/tailbits/mason/openapi"
)

var _ model.Entity = (*Input)(nil)

// Input Model
type Input struct {
	Increment *int `json:"increment"`
}

// Example implements model.Entity.
func (r *Input) Example() []byte {
	return []byte(`{
		"increment": 5 
	}`)
}

func (r *Input) Marshal() (json.RawMessage, error) {
	return json.Marshal(r)
}

func (r *Input) Name() string {
	return "IncrementInput"
}

func (r *Input) Schema() []byte {
	return []byte(`{
		"type": "object",
		"properties": {
			"increment": {
				"type": ["integer", "null"]
			}
		},
		"additionalProperties": false
	}`)
}

func (r *Input) Unmarshal(data json.RawMessage) error {
	return json.Unmarshal(data, r)
}

// Output Model
var _ model.Entity = (*Response)(nil)

type Response struct {
	Count  int            `json:"count"`
	Server HealthResponse `json:"server"`
}

// Example implements model.Entity.
func (r *Response) Example() []byte {
	return []byte(`{
		"count": 5
	}`)
}

func (r *Response) Marshal() (json.RawMessage, error) {
	return json.Marshal(r)
}

func (r *Response) Name() string {
	return "CountResponse"
}

func (r *Response) Schema() []byte {
	return []byte(`{
		"type": "object",
		"properties": {
			"count": {
				"type": "integer"
			},
			"server": {
				"$ref": "#/definitions/HealthResponse"
			}
		},
		"required": ["count"]
	}`)
}

func (r *Response) Unmarshal(data json.RawMessage) error {
	return json.Unmarshal(data, r)
}

// HealthResponse Model
var _ model.Entity = (*Response)(nil)

type HealthResponse struct {
	Timestamp time.Time `json:"timestamp"`
}

// Example implements model.Entity.
func (r *HealthResponse) Example() []byte {
	return []byte(`{
		"timestamp": "2023-10-01T12:00:00Z"
	}`)
}

func (r *HealthResponse) Marshal() (json.RawMessage, error) {
	return json.Marshal(r)
}

func (r *HealthResponse) Name() string {
	return "HealthResponse"
}

func (r *HealthResponse) Schema() []byte {
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

func (r *HealthResponse) Unmarshal(data json.RawMessage) error {
	return json.Unmarshal(data, r)
}

func HealthCheckHandler(ctx context.Context, r *http.Request, params model.Nil) (rsp *HealthResponse, err error) {
	return &HealthResponse{
		Timestamp: time.Now().UTC(),
	}, nil
}

// =============================================================================
// Handlers
var count int

func IncrementHandler(ctx context.Context, r *http.Request, inp *Input, params model.Nil) (rsp *Response, err error) {
	inc := 1
	if inp.Increment != nil {
		inc = *inp.Increment
	}
	count += inc

	return &Response{
		Count: count,
		Server: HealthResponse{
			Timestamp: time.Now().UTC(),
		},
	}, nil
}

func main() {
	rtm := mason.NewHTTPRuntime()
	api := mason.NewAPI(rtm)
	grp := api.NewRouteGroup("counter")

	grp.Register(mason.HandlePost(IncrementHandler).
		Path("/increment").
		WithOpID("increment").
		WithSummary("Increment the counter").
		WithDesc("Increment the counter by one, or the supplied increment"))

	grp.Register(mason.HandleGet(HealthCheckHandler).
		Path("/healthcheck").
		WithOpID("healthcheck").
		WithSummary("Get the server status"))

	// Generate the OpenAPI schema
	gen, err := openapi.NewGenerator(api)
	if err != nil {
		panic(fmt.Errorf("failed to create OpenAPI generator: %w", err))
	}
	gen.Spec.Info.WithTitle("Healthy Counter API")

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
