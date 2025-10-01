package mason_test

import (
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/magicbell/mason"
)

type decodeTest[T any] struct {
	Name        string
	decodeTests []struct {
		Name        string
		QueryString string
		Expected    T
		ExpectError bool
	}
}

func TestDecodeQueryParams(t *testing.T) {
	nameAge := decodeTest[struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}]{
		Name: "String and Int params",
		decodeTests: []struct {
			Name        string
			QueryString string
			Expected    struct {
				Name string `json:"name"`
				Age  int    `json:"age"`
			}
			ExpectError bool
		}{
			{
				Name:        "Valid params",
				QueryString: "name=John&age=30",
				Expected: struct {
					Name string `json:"name"`
					Age  int    `json:"age"`
				}{Name: "John", Age: 30},
				ExpectError: false,
			},
			{
				Name:        "Missing age",
				QueryString: "name=John",
				Expected: struct {
					Name string `json:"name"`
					Age  int    `json:"age"`
				}{Name: "John", Age: 0},
				ExpectError: false,
			},
			{
				Name:        "Invalid age",
				QueryString: "name=John&age=invalid",
				Expected: struct {
					Name string `json:"name"`
					Age  int    `json:"age"`
				}{},
				ExpectError: true,
			},
		},
	}

	boolPtr := decodeTest[struct {
		Active   bool    `json:"active"`
		Nickname *string `json:"nickname"`
	}]{
		Name: "Boolean and Pointer params",
		decodeTests: []struct {
			Name        string
			QueryString string
			Expected    struct {
				Active   bool    `json:"active"`
				Nickname *string `json:"nickname"`
			}
			ExpectError bool
		}{
			{
				Name:        "Valid params",
				QueryString: "active=true&nickname=Johnny",
				Expected: struct {
					Active   bool    `json:"active"`
					Nickname *string `json:"nickname"`
				}{Active: true, Nickname: ptr("Johnny")},
				ExpectError: false,
			},
			{
				Name:        "Invalid boolean",
				QueryString: "active=invalid&nickname=Johnny",
				Expected: struct {
					Active   bool    `json:"active"`
					Nickname *string `json:"nickname"`
				}{},
				ExpectError: true,
			},
		},
	}

	withDefaults := decodeTest[struct {
		Name string `json:"name"`
		Age  int    `json:"age" default:"10"`
	}]{
		Name: "With defaults",
		decodeTests: []struct {
			Name        string
			QueryString string
			Expected    struct {
				Name string `json:"name"`
				Age  int    `json:"age" default:"10"`
			}
			ExpectError bool
		}{
			{
				Name:        "Valid params",
				QueryString: "name=John&age=30",
				Expected: struct {
					Name string `json:"name"`
					Age  int    `json:"age" default:"10"`
				}{Name: "John", Age: 30},
				ExpectError: false,
			},
			{
				Name:        "Missing age",
				QueryString: "name=John",
				Expected: struct {
					Name string `json:"name"`
					Age  int    `json:"age" default:"10"`
				}{Name: "John", Age: 10},
				ExpectError: false,
			},
		},
	}

	run(nameAge, t)
	run(boolPtr, t)
	run(withDefaults, t)

	// Time parsing tests
	timeTests := decodeTest[struct {
		From time.Time  `json:"from"`
		To   *time.Time `json:"to"`
	}]{
		Name: "Time params (RFC3339 and Unix)",
		decodeTests: []struct {
			Name        string
			QueryString string
			Expected    struct {
				From time.Time  `json:"from"`
				To   *time.Time `json:"to"`
			}
			ExpectError bool
		}{
			{
				Name:        "RFC3339 both",
				QueryString: "from=2025-10-01T09:36:00Z&to=2025-10-01T10:00:00Z",
				Expected: func() struct {
					From time.Time  `json:"from"`
					To   *time.Time `json:"to"`
				} {
					from, _ := time.Parse(time.RFC3339, "2025-10-01T09:36:00Z")
					to, _ := time.Parse(time.RFC3339, "2025-10-01T10:00:00Z")
					return struct {
						From time.Time  `json:"from"`
						To   *time.Time `json:"to"`
					}{From: from, To: &to}
				}(),
				ExpectError: false,
			},
			{
				Name:        "Unix seconds",
				QueryString: "from=1730379600&to=1730383200",
				Expected: func() struct {
					From time.Time  `json:"from"`
					To   *time.Time `json:"to"`
				} {
					from := time.Unix(1730379600, 0).UTC()
					to := time.Unix(1730383200, 0).UTC()
					return struct {
						From time.Time  `json:"from"`
						To   *time.Time `json:"to"`
					}{From: from, To: &to}
				}(),
				ExpectError: false,
			},
			{
				Name:        "Unix milliseconds",
				QueryString: "from=1730379600000&to=1730383200000",
				Expected: func() struct {
					From time.Time  `json:"from"`
					To   *time.Time `json:"to"`
				} {
					from := time.Unix(1730379600, 0).UTC()
					to := time.Unix(1730383200, 0).UTC()
					return struct {
						From time.Time  `json:"from"`
						To   *time.Time `json:"to"`
					}{From: from, To: &to}
				}(),
				ExpectError: false,
			},
			{
				Name:        "Invalid time",
				QueryString: "from=notatime",
				Expected:    struct { From time.Time `json:"from"`; To *time.Time `json:"to"` }{},
				ExpectError: true,
			},
		},
	}
	run(timeTests, t)

	// Tolerant string time variants
	tolerant := decodeTest[struct {
		At time.Time `json:"at"`
	}]{
		Name: "Time params tolerant formats",
		decodeTests: []struct {
			Name        string
			QueryString string
			Expected    struct {
				At time.Time `json:"at"`
			}
			ExpectError bool
		}{
			{
				Name:        "RFC3339 Z",
				QueryString: "at=2025-09-20T12:00:00Z",
				Expected: func() struct{ At time.Time `json:"at"` } {
					at, _ := time.Parse(time.RFC3339, "2025-09-20T12:00:00Z")
					return struct{ At time.Time `json:"at"` }{At: at}
				}(),
				ExpectError: false,
			},
			{
				Name:        "No zone with seconds",
				QueryString: "at=2025-09-20T12:00:00",
				Expected: struct{ At time.Time `json:"at"` }{At: time.Date(2025, 9, 20, 12, 0, 0, 0, time.UTC)},
				ExpectError: false,
			},
			{
				Name:        "No zone minutes",
				QueryString: "at=2025-09-20T12:00",
				Expected: struct{ At time.Time `json:"at"` }{At: time.Date(2025, 9, 20, 12, 0, 0, 0, time.UTC)},
				ExpectError: false,
			},
			{
				Name:        "Date only",
				QueryString: "at=2025-09-20",
				Expected: struct{ At time.Time `json:"at"` }{At: time.Date(2025, 9, 20, 0, 0, 0, 0, time.UTC)},
				ExpectError: false,
			},
		},
	}
	run(tolerant, t)
}

func run[Q any](decodeTest decodeTest[Q], t *testing.T) {
	for _, tt := range decodeTest.decodeTests {
		t.Run(tt.Name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/?"+tt.QueryString, nil) // nolint: noctx
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			result, err := mason.DecodeQueryParams[Q](req)
			if tt.ExpectError {
				if err == nil {
					t.Errorf("Expected an error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if !reflect.DeepEqual(result, tt.Expected) {
					t.Errorf("Expected %+v, but got %+v", tt.Expected, result)
				}
			}
		})
	}
}

func ptr[T any](v T) *T {
	return &v
}
