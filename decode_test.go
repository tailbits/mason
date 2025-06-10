package mason_test

import (
	"net/http"
	"reflect"
	"testing"

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
