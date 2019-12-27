package openapi3filter_test

import (
	"net/http"
	"sort"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/stretchr/testify/require"
)

func TestRouter(t *testing.T) {
	var (
		pathNotFound = "Path was not found"
		methodNotAllowed = "Path doesn't support the HTTP method"
		doesNotMatchAnyServer = "Does not match any server"
	)

	// Build swagger
	helloCONNECT := &openapi3.Operation{Responses: make(openapi3.Responses)}
	helloDELETE := &openapi3.Operation{Responses: make(openapi3.Responses)}
	helloGET := &openapi3.Operation{Responses: make(openapi3.Responses)}
	helloHEAD := &openapi3.Operation{Responses: make(openapi3.Responses)}
	helloOPTIONS := &openapi3.Operation{Responses: make(openapi3.Responses)}
	helloPATCH := &openapi3.Operation{Responses: make(openapi3.Responses)}
	helloPOST := &openapi3.Operation{Responses: make(openapi3.Responses)}
	helloPUT := &openapi3.Operation{Responses: make(openapi3.Responses)}
	helloTRACE := &openapi3.Operation{Responses: make(openapi3.Responses)}
	paramsGET := &openapi3.Operation{Responses: make(openapi3.Responses)}
	partialGET := &openapi3.Operation{Responses: make(openapi3.Responses)}
	swagger := &openapi3.Swagger{
		OpenAPI: "3.0.0",
		Info: &openapi3.Info{
			Title:   "MyAPI",
			Version: "0.1",
		},
		Paths: openapi3.Paths{
			"/hello": &openapi3.PathItem{
				Connect: helloCONNECT,
				Delete:  helloDELETE,
				Get:     helloGET,
				Head:    helloHEAD,
				Options: helloOPTIONS,
				Patch:   helloPATCH,
				Post:    helloPOST,
				Put:     helloPUT,
				Trace:   helloTRACE,
			},
			"/onlyGET": &openapi3.PathItem{
				Get: helloGET,
			},
			"/params/{x}/{y}/{z*}": &openapi3.PathItem{
				Get: paramsGET,
			},
			"/partial": &openapi3.PathItem{
				Get: partialGET,
			},
		},
	}

	// Build router
	router := openapi3filter.NewRouter().WithSwagger(swagger)

	// Declare a helper function
	expect := func(method string, uri string, operation *openapi3.Operation, params map[string]string) {
		req, err := http.NewRequest(method, uri, nil)
		require.NoError(t, err)
		route, pathParams, err := router.FindRoute(req.Method, req.URL)
		if err != nil {
			if operation == nil {
				if err.Error() == doesNotMatchAnyServer {
					return
				}

				pathItem := swagger.Paths[uri]
				if pathItem == nil {
					if err.Error() != pathNotFound {
						t.Fatalf("'%s %s': should have returned '%s', but it returned an error: %v",
							method, uri, pathNotFound, err)
					}
					return
				}
				if pathItem.GetOperation(method) == nil {
					if err.Error() != methodNotAllowed {
						t.Fatalf("'%s %s': should have returned '%s', but it returned an error: %v",
							method, uri, methodNotAllowed, err)
					}
				}
			} else {
				t.Fatalf("'%s %s': should have returned an operation, but it returned an error: %v",
					method, uri, err)
			}
		}
		if operation == nil && err == nil {
			t.Fatalf("'%s %s': should have returned an error, but didn't",
				method, uri)
		}
		if route == nil {
			return
		}
		if route.Operation != operation {
			t.Fatalf("'%s %s': Returned wrong operation (%v)",
				method, uri, route.Operation)
		}
		if params == nil {
			if len(pathParams) != 0 {
				t.Fatalf("'%s %s': should return no path arguments, but found some",
					method, uri)
			}
		} else {
			names := make([]string, 0, len(params))
			for name := range params {
				names = append(names, name)
			}
			sort.Strings(names)
			for _, name := range names {
				expected := params[name]
				actual, exists := pathParams[name]
				if !exists {
					t.Fatalf("'%s %s': path parameter '%s' should be '%s', but it's not defined.",
						method, uri, name, expected)
				}
				if actual != expected {
					t.Fatalf("'%s %s': path parameter '%s' should be '%s', but it's '%s'",
						method, uri, name, expected, actual)
				}
			}
		}
	}
	expect("GET", "/not_existing", nil, nil)
	expect("DELETE", "/hello", helloDELETE, nil)
	expect("GET", "/hello", helloGET, nil)
	expect("HEAD", "/hello", helloHEAD, nil)
	expect("PATCH", "/hello", helloPATCH, nil)
	expect("POST", "/hello", helloPOST, nil)
	expect("PUT", "/hello", helloPUT, nil)
	expect("GET", "/params/a/b/c/d", paramsGET, map[string]string{
		"x": "a",
		"y": "b",
		"z": "c/d",
	})
	expect("POST", "/partial", nil, nil)
	swagger.Servers = append(swagger.Servers, &openapi3.Server{
		URL: "https://www.example.com/api/v1/",
	}, &openapi3.Server{
		URL: "https://{d0}.{d1}.com/api/v1/",
	})
	expect("GET", "/hello", nil, nil)
	expect("GET", "/api/v1/hello", nil, nil)
	expect("GET", "www.example.com/api/v1/hello", nil, nil)
	expect("GET", "https:///api/v1/hello", nil, nil)
	expect("GET", "https://www.example.com/hello", nil, nil)
	expect("GET", "https://www.example.com/api/v1/hello", helloGET, map[string]string{})
	expect("GET", "https://domain0.domain1.com/api/v1/hello", helloGET, map[string]string{
		"d0": "domain0",
		"d1": "domain1",
	})

	{
		uri := "https://www.example.com/api/v1/onlyGET"
		expect(http.MethodGet, uri, helloGET, nil)
		req, err := http.NewRequest(http.MethodDelete, uri, nil)
		require.NoError(t, err)
		require.NotNil(t, req)
		route, pathParams, err := router.FindRoute(req.Method, req.URL)
		require.Error(t, err)
		require.Equal(t, err.(*openapi3filter.RouteError).Reason, "Path doesn't support the HTTP method")
		require.Nil(t, route)
		require.Nil(t, pathParams)
	}
}
