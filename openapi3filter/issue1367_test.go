package openapi3filter_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

func setupTestRouter(t *testing.T, spec string) routers.Router {
	t.Helper()
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(spec))
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	router, err := gorillamux.NewRouter(doc)
	require.NoError(t, err)

	return router
}

func testIssue1367Request(t *testing.T, spec, url, errContains string) {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)

	router := setupTestRouter(t, spec)
	route, pathParams, err := router.FindRoute(req)
	require.NoError(t, err)

	err = openapi3filter.ValidateRequest(t.Context(), &openapi3filter.RequestValidationInput{
		Request:    req,
		PathParams: pathParams,
		Route:      route,
	})
	if errContains != "" {
		require.ErrorContains(t, err, errContains)
	} else {
		require.NoError(t, err)
	}
}

func TestIssue1367TypelessConstAndEnumPrimitives(t *testing.T) {
	t.Parallel()

	const spec = `
openapi: 3.1.0
info:
  title: repro
  version: 1.0.0
paths:
  /items:
    get:
      parameters:
        - name: kind
          in: query
          required: true
          schema:
            const: widget
        - name: withenum
          in: query
          schema:
            enum: [widget]
      responses:
        '200':
          description: ok
`
	const rootURL = "http://example.com/items"

	testcases := []struct {
		name        string
		url         string
		errContains string
	}{{
		name: "const matches",
		url:  "?kind=widget",
	}, {
		name: "const and enum match",
		url:  "?kind=widget&withenum=widget",
	}, {
		name:        "enum mismatch",
		url:         "?kind=widget&withenum=gadget",
		errContains: "parameter \"withenum\" in query has an error: validation failed due to: at '': value must be 'widget'",
	}, {
		name:        "const mismatch",
		url:         "?kind=gadget",
		errContains: "parameter \"kind\" in query has an error: validation failed due to: at '': value must be 'widget'",
	}, {
		name:        "empty enum value",
		url:         "?kind=widget&withenum=",
		errContains: "parameter \"withenum\" in query has an error: validation failed due to: at '': value must be 'widget'",
	}}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			testIssue1367Request(t, spec, rootURL+tc.url, tc.errContains)
		})
	}
}

func TestIssue1367TypelessConstAndEnumNumberAndBoolean(t *testing.T) {
	t.Parallel()

	const spec = `
openapi: 3.1.0
info:
  title: repro
  version: 1.0.0
paths:
  /picked:
    get:
      parameters:
        - name: rank
          in: query
          required: true
          schema:
            enum: [1, 2, 3]
        - name: flag
          in: query
          schema:
            const: true
      responses:
        '200':
          description: ok
`
	const rootURL = "http://example.com/picked"

	testcases := []struct {
		name        string
		url         string
		errContains string
	}{{
		name: "number enum and boolean const match",
		url:  "?rank=2&flag=true",
	}, {
		name:        "boolean const mismatch",
		url:         "?rank=2&flag=false",
		errContains: "parameter \"flag\" in query has an error: validation failed due to: at '': value must be true",
	}, {
		name:        "number enum mismatch",
		url:         "?rank=9&flag=true",
		errContains: "parameter \"rank\" in query has an error: validation failed due to: at '': value must be one of 1, 2, 3",
	}}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			testIssue1367Request(t, spec, rootURL+tc.url, tc.errContains)
		})
	}
}

// A typeless enum may mix kinds; a numeric string still decodes to the number.
func TestIssue1367TypelessEnumMixedKinds(t *testing.T) {
	t.Parallel()

	const spec = `
openapi: 3.1.0
info:
  title: repro
  version: 1.0.0
paths:
  /mid:
    get:
      parameters:
        - name: mid
          in: query
          required: true
          schema:
            enum: [42, "42", [42]]
      responses:
        '200':
          description: ok
`
	const rootURL = "http://example.com/mid"

	testcases := []struct {
		name        string
		url         string
		errContains string
	}{{
		name: "numeric string decodes to number",
		url:  "?mid=42",
	}, {
		name:        "no member matches",
		url:         "?mid=gadget",
		errContains: "parameter \"mid\" in query has an error: validation failed due to: at '': 'enum' failed",
	}, {
		name:        "array member never matches",
		url:         "?mid=%5B42%5D",
		errContains: "parameter \"mid\" in query has an error: validation failed due to: at '': 'enum' failed",
	}}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			testIssue1367Request(t, spec, rootURL+tc.url, tc.errContains)
		})
	}
}
