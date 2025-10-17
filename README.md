# Mason

Mason is a JSON schema-backed API framework for writing HTTP handlers in Go. It offers input validation and decoding and generates an OpenAPI 3.1 spec. By design, it has a small API, so it can be adopted by teams with an existing codebase.

## Background

It was created to serve the API (v2) at [MagicBell](https://www.magicbell.com?utm_source=github&utm_campaign=mason), and guided by the following design goals:

- **JSON schema first** - The Input/Output models are described by JSON schema, with an example. By implementing the `model.Entity` interface, the model definition is tested against the schema so they are never out of sync.
- **Incremental adoption** - Mason should be easy to add to an existing project, by giving it a `mason.Runtime` implementation that can `Handle` the `Operation` created by Mason, and `Respond` to a HTTP request.
- **Support Resource grouping & querying** - REST API resources and endpoints are a map to an API/product's feature offerings. For example `/integrations/slack`, and `/integrations/web_push` are two different resources, but to get all `integration` resources, the integration `RouteGroup` comes in handy.

## Usage

Add it to your project:

```bash
  go get github.com/tailbits/mason@latest
```

You'll need a `Runtime` implementation to start using Mason in your existing project, but for new projects, or to kick the tires, you can use the `[HTTPRuntime](runtime.go)`.

```go
  api := mason.NewAPI(mason.NewHTTPRuntime())
```

##\* `GET` Handler

Let's add a new `GET /ping` endpoint that returns the current timestamp. To do this, we need to define the output struct.

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

Now we can use it to define the Handler (note that we use [model.Nil](model/nil.go) for decoding query params, which signals that we don't care about them. Since this is GET request, there is no input struct, but `model.Nil` can also be used for `POST/PUT` handlers that accept no request body.

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

### `POST` Handler

Let's move on to more the more exciting stuff, and build a small counter API. Let's add a `POST` handler that will increment the counter by 1, or by an `increment`, which is an optional field in the request body. Once again, we start by defining a model that confirms to the `platform.Entity` interface.

```go

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
			"timestamp": {
				"type": ["integer", "null"]
			}
		}
	}`)
}

func (r *Input) Unmarshal(data json.RawMessage) error {
	return json.Unmarshal(data, r)
}
```

Let's define the `CountResponse` as the Output model for the `POST` (as well as the `GET`) handler.

```go
// Output Model
var _ model.Entity = (*Response)(nil)

type Response struct {
	Count int `json:"count"`
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
			}
		},
		"required": ["count"]
	}`)
}

func (r *Response) Unmarshal(data json.RawMessage) error {
	return json.Unmarshal(data, r)
}
```

Finally, we can define the `POST` handler, and accept the validated and decoded `Input` model in our code.

```go
  var count int

  func IncrementHandler(ctx context.Context, r *http.Request, inp *Input, params model.Nil) (rsp *Response, err error) {
    inc := 1
    if inp.Increment != nil {
      inc = *inp.Increment
    }
    count += inc

    return &Response{
      Count: count,
    }, nil
  }
```

Registering the handler on the route group

```go
	rtm := mason.NewHTTPRuntime()
	api := mason.NewAPI(rtm)
	grp := api.NewRouteGroup("counter")

	grp.Register(mason.HandlePost(IncrementHandler).
		Path("/increment").
		WithOpID("increment").
		WithSummary("Increment the counter").
		WithDesc("Increment the counter by one, or the supplied increment"))
```

The code for this example is in [example/counter/main.go](example/counter/main.go), and it contains a `GET` handler, as well the route for grabbing the OpenAPI file.

Let's start counting!

```bash
# --data sends a POST request with curl
curl http://localhost:9090/increment \
  --header 'Content-Type: application/json' \
  --data '{}'
{"count":1}
```

Let's increment by 2

```bash
curl http://localhost:9090/increment \
  --header 'Content-Type: application/json' \
  --data '{"increment": 2}'
{"count":3}
```

How about some invalid input?

```bash
# -v to see the response code and body
curl -v http://localhost:9090/increment \
  --header 'Content-Type: application/json' \
  --data '{"increment": "2"}'

* upload completely sent off: 18 bytes
< HTTP/1.1 422 Unprocessable Entity
< Content-Type: application/json
< Date: Fri, 13 Jun 2025 08:02:05 GMT
< Content-Length: 91
<
{"errors":[{"error":null,"message":"Param 'increment' should be of type [integer,null]"}]}
```

The `Respond` method handles the error formatting in the `HTTPRuntime` by checking if the error is of type [model.ValidationError](model/error.go). This code can be the starting point for returning formatted errors to your users, in your Runtime implementation, or a middleware.

```go
  func (r *HTTPRuntime) Handle(method string, path string, handler WebHandler, mws ...func(WebHandler) WebHandler) {
    r.HandleFunc(fmt.Sprintf("%s %s", method, path), func(w http.ResponseWriter, req *http.Request) {
      if req.Method != method {
        http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
        return
      }

      ctx := req.Context()
      if err := handler(ctx, w, req); err != nil {
        // Format validation Error
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
```

## Schema Registry

Mason collects the schema of every I/O model registered via the Handlers in a registry. This allows for resolving schema references, and other goodies!

### Schema Dereference

To illustrate the derefeencing, take a look at the [example/schemaexample/main.go](example/schemaexample/main.go), which recreates the `POST /increment` handler from the counter example, but this time, returns a `server` key in the response. The server key contains a timestamp, and we also add a `GET /healthcheck` endpoint that returns the same key in it's response.

First, let's setup the `GET /healthcheck` output model and handler.

```go
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
```

Is the API healthy? Let's find out!

```
➜  ext git:(chore/mason-update) ✗ curl http://localhost:9090/healthcheck
{"timestamp":"2025-06-13T09:26:31.615403Z"}
```

We can use the schema contributed by the `HealthResponse`, and identified by it's `Name()` in the `Response` model for the `POST /increment` handler.

```go
  type Response struct {
    Count  int            `json:"count"`
    Server HealthResponse `json:"server"`
  }


  func (r *Response) Schema() []byte {
    return []byte(`{
      "type": "object",
      "properties": {
        "count": {
          "type": "integer"
        },
        "server": {
          "$ref": "#/components/schemas/HealthResponse"
        }
      },
      "required": ["count"]
    }`)
  }
```

Let's increment the counter!

```bash
curl http://localhost:9090/increment \
  --header 'Content-Type: application/json' \
  --data '{"increment": 2}'
{"count":2,"server":{"timestamp":"2025-06-13T09:25:37.031658Z"}}
```

The generated OpenAPI schema uses the registry references, and if we were composing an input model, the dereferenced schema would also be used for validation.

```json
...
  "paths": {
      "/healthcheck": {
        "get": {
          "summary": "Get the server status",
          "description": "",
          "operationId": "healthcheck",
          "responses": {
            "200": {
              "description": "OK",
              "content": {
                "application/json": {
                  "schema": {
                    "$ref": "#/components/schemas/HealthResponse"
                  }
                }
              }
            }
          }
        }
      },
      "/increment": {
        "post": {
          "summary": "Increment the counter",
          "description": "Increment the counter by one, or the supplied increment",
          "operationId": "increment",
          "requestBody": {
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/IncrementInput"
                }
              }
            }
          },
          "responses": {
            "201": {
              "description": "Created",
              "content": {
                "application/json": {
                  "schema": {
                    "$ref": "#/components/schemas/CountResponse"
                  }
                }
              }
            }
          },
          "x-forbid-unknown-cookie": true,
          "x-forbid-unknown-header": true,
          "x-forbid-unknown-path": true,
          "x-forbid-unknown-query": true
        }
      }
    },
    "components": {
      "schemas": {
        "CountResponse": {
          "examples": [
            {
              "count": 5
            }
          ],
...
```

If you need the dereferenced schema, you can grab it from the api instance.

```go
	sch, err := api.DereferenceSchema(schema_with_references)
```

### Model, Schema, and Example Sync

The [sync](model/sync) package can check if the model's struct, schema, and example are in sync.

## Development Status

As mentioned in the intro, Mason is in active development and usage at [MagicBell](https://www.magicbell.com?utm_source=github&utm_campaign=mason). In open-sourcing it, we want to give back to the incredible Go community, and also receive feedback, contributions, and ideas for improvements.

Plesae create issues with your questions, and if you use Mason, we'd love to hear from you, too!
