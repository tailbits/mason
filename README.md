# Mason

Mason is an API framework for writing HTTP handlers with Input/Output models described by JSON schema.
It was created to serve the API (v2) at [MagicBell](https://www.magicbell.com), and guided by the following design goals:

- **JSON schema first** - The Input/Output models are described by JSON schema, with an example. By implementing the `model.Entity` interface I/O model definition is tested against the schema so they are never out of sync.
- **Incremental adoption** - Mason should be easy to add to an existing project, by giving it a `mason.Runtime` implementation that can `Handle` the `Operation` created by Mason, and `Respond` to a HTTP request.
- **Support Resource grouping & querying** - REST API resources and endpoints are a map to an API/product's feature offerings. For example `/integrations/slack`, and `/integrations/web_push` are two different resources, but to get all `integration` resources, the integration `RouteGroup` comes in handy.

## Usage

Add it to your project with

```bash
  go get github.com/magicbell/mason@latest
```

You'll need a `Runtime` implementation to start using Mason in your existing project, but for new projects, or to kick the tires, you can use the `HTTPRuntime`.

```go
  api := mason.NewAPI(mason.NewHTTPRuntime())
```

Let's add a new `GET /ping` endpoint that returns the current timestamp. To do this, we need to define the output struct

```go
  var _ model.Entity = (*Response)(nil)

  type Response struct {
    Timestamp time.Time `json:"timestamp"`
  }

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

```

Now we can use it to define the Handler (note that we use `model.Nil` for decoding query params. Since this is GET request, there is no input struct, but `model.Nil` can also be used for `POST/PUT` handlers that accept no request body.

```go
  func PingHandler(ctx context.Context, r *http.Request, params model.Nil) (rsp *Response, err error) {
    return &Response{
      Timestamp: time.Now().UTC(),
    }, nil
  }
```

Create a new RouteGroup

```
  api.NewRouteGroup("ping")
```

Register the Handler

```go
	grp.Register(mason.HandleGet(PingHandler).
		Path("/ping").
		WithOpID("ping").
		WithSummary("Ping the server").
		WithDesc("Ping the server when you are unsure of the time"))
```

You can try this example by running [example/ping/main.go](/example/ping/main.go). The example also mounts a handler to serve the OpenAPI file.

```go
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
```
