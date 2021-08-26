package legacy

import (
	"context"
	"net/http"
	"sort"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/routers"
	"github.com/stretchr/testify/require"
)

func TestRouter(t *testing.T) {
	helloCONNECT := &openapi3.Operation{Responses: openapi3.NewResponses()}
	helloDELETE := &openapi3.Operation{Responses: openapi3.NewResponses()}
	helloGET := &openapi3.Operation{Responses: openapi3.NewResponses()}
	helloHEAD := &openapi3.Operation{Responses: openapi3.NewResponses()}
	helloOPTIONS := &openapi3.Operation{Responses: openapi3.NewResponses()}
	helloPATCH := &openapi3.Operation{Responses: openapi3.NewResponses()}
	helloPOST := &openapi3.Operation{Responses: openapi3.NewResponses()}
	helloPUT := &openapi3.Operation{Responses: openapi3.NewResponses()}
	helloTRACE := &openapi3.Operation{Responses: openapi3.NewResponses()}
	paramsGET := &openapi3.Operation{Responses: openapi3.NewResponses()}
	booksPOST := &openapi3.Operation{Responses: openapi3.NewResponses()}
	partialGET := &openapi3.Operation{Responses: openapi3.NewResponses()}
	doc := &openapi3.T{
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
			"/params/{x}/{y}/{z.*}": &openapi3.PathItem{
				Get: paramsGET,
				Parameters: openapi3.Parameters{
					&openapi3.ParameterRef{Value: openapi3.NewPathParameter("x")},
					&openapi3.ParameterRef{Value: openapi3.NewPathParameter("y")},
					&openapi3.ParameterRef{Value: openapi3.NewPathParameter("z")},
				},
			},
			"/books/{bookid}": &openapi3.PathItem{
				Get: paramsGET,
				Parameters: openapi3.Parameters{
					&openapi3.ParameterRef{Value: openapi3.NewPathParameter("bookid")},
				},
			},
			"/books/{bookid2}.json": &openapi3.PathItem{
				Post: booksPOST,
				Parameters: openapi3.Parameters{
					&openapi3.ParameterRef{Value: openapi3.NewPathParameter("bookid2")},
				},
			},
			"/partial": &openapi3.PathItem{
				Get: partialGET,
			},
		},
	}

	expect := func(r routers.Router, method string, uri string, operation *openapi3.Operation, params map[string]string) {
		req, err := http.NewRequest(method, uri, nil)
		require.NoError(t, err)
		route, pathParams, err := r.FindRoute(req)
		if err != nil {
			if operation == nil {
				pathItem := doc.Paths[uri]
				if pathItem == nil {
					if err.Error() != routers.ErrPathNotFound.Error() {
						t.Fatalf("'%s %s': should have returned %q, but it returned an error: %v", method, uri, routers.ErrPathNotFound, err)
					}
					return
				}
				if pathItem.GetOperation(method) == nil {
					if err.Error() != routers.ErrMethodNotAllowed.Error() {
						t.Fatalf("'%s %s': should have returned %q, but it returned an error: %v", method, uri, routers.ErrMethodNotAllowed, err)
					}
				}
			} else {
				t.Fatalf("'%s %s': should have returned an operation, but it returned an error: %v", method, uri, err)
			}
		}
		if operation == nil && err == nil {
			t.Fatalf("'%s %s': should have failed, but returned\nroute = %+v\npathParams = %+v", method, uri, route, pathParams)
		}
		if route == nil {
			return
		}
		if route.Operation != operation {
			t.Fatalf("'%s %s': Returned wrong operation (%v)",
				method, uri, route.Operation)
		}
		if len(params) == 0 {
			if len(pathParams) != 0 {
				t.Fatalf("'%s %s': should return no path arguments, but found %+v", method, uri, pathParams)
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
					t.Fatalf("'%s %s': path parameter %q should be %q, but it's not defined.", method, uri, name, expected)
				}
				if actual != expected {
					t.Fatalf("'%s %s': path parameter %q should be %q, but it's %q", method, uri, name, expected, actual)
				}
			}
		}
	}

	err := doc.Validate(context.Background())
	require.NoError(t, err)
	r, err := NewRouter(doc)
	require.NoError(t, err)

	expect(r, http.MethodGet, "/not_existing", nil, nil)
	expect(r, http.MethodDelete, "/hello", helloDELETE, nil)
	expect(r, http.MethodGet, "/hello", helloGET, nil)
	expect(r, http.MethodHead, "/hello", helloHEAD, nil)
	expect(r, http.MethodPatch, "/hello", helloPATCH, nil)
	expect(r, http.MethodPost, "/hello", helloPOST, nil)
	expect(r, http.MethodPut, "/hello", helloPUT, nil)
	expect(r, http.MethodGet, "/params/a/b/", paramsGET, map[string]string{
		"x": "a",
		"y": "b",
		// "z": "",
	})
	expect(r, http.MethodGet, "/params/a/b/c%2Fd", paramsGET, map[string]string{
		"x": "a",
		"y": "b",
		// "z": "c/d",
	})
	expect(r, http.MethodGet, "/books/War.and.Peace", paramsGET, map[string]string{
		"bookid": "War.and.Peace",
	})
	{
		req, err := http.NewRequest(http.MethodPost, "/books/War.and.Peace.json", nil)
		require.NoError(t, err)
		_, _, err = r.FindRoute(req)
		require.EqualError(t, err, routers.ErrPathNotFound.Error())
	}
	expect(r, http.MethodPost, "/partial", nil, nil)

	doc.Servers = []*openapi3.Server{
		{URL: "https://www.example.com/api/v1"},
		{URL: "https://{d0}.{d1}.com/api/v1/", Variables: map[string]*openapi3.ServerVariable{
			"d0": {Default: "www"},
			"d1": {Default: "example", Enum: []string{"example"}},
		}},
	}
	err = doc.Validate(context.Background())
	require.NoError(t, err)
	r, err = NewRouter(doc)
	require.NoError(t, err)
	expect(r, http.MethodGet, "/hello", nil, nil)
	expect(r, http.MethodGet, "/api/v1/hello", nil, nil)
	expect(r, http.MethodGet, "www.example.com/api/v1/hello", nil, nil)
	expect(r, http.MethodGet, "https:///api/v1/hello", nil, nil)
	expect(r, http.MethodGet, "https://www.example.com/hello", nil, nil)
	expect(r, http.MethodGet, "https://www.example.com/api/v1/hello", helloGET, nil)
	expect(r, http.MethodGet, "https://domain0.domain1.com/api/v1/hello", helloGET, map[string]string{
		"d0": "domain0",
		"d1": "domain1",
	})

	{
		uri := "https://www.example.com/api/v1/onlyGET"
		expect(r, http.MethodGet, uri, helloGET, nil)
		req, err := http.NewRequest(http.MethodDelete, uri, nil)
		require.NoError(t, err)
		require.NotNil(t, req)
		route, pathParams, err := r.FindRoute(req)
		require.EqualError(t, err, routers.ErrMethodNotAllowed.Error())
		require.Nil(t, route)
		require.Nil(t, pathParams)
	}
}
