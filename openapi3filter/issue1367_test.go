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

func testIssue1367Request(t *testing.T, spec, url string, fail bool) {
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
	if fail {
		require.Error(t, err)
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
		url  string
		fail bool
	}{{
		url: "?kind=widget",
	}, {
		url: "?kind=widget&withenum=widget",
	}, {
		url:  "?kind=widget&withenum=gadget",
		fail: true,
	}, {
		url:  "?kind=gadget",
		fail: true,
	}, {
		url:  "?kind=widget&withenum=",
		fail: true,
	}}
	for _, tc := range testcases {
		t.Run(tc.url, func(t *testing.T) {
			testIssue1367Request(t, spec, rootURL+tc.url, tc.fail)
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
		url  string
		fail bool
	}{{
		url: "?rank=2&flag=true",
	}, {
		url:  "?rank=2&flag=false",
		fail: true,
	}, {
		url:  "?rank=9&flag=true",
		fail: true,
	}}
	for _, tc := range testcases {
		t.Run(tc.url, func(t *testing.T) {
			testIssue1367Request(t, spec, rootURL+tc.url, tc.fail)
		})
	}
}

// The accept set of a typeless `enum`/`const` may mix kinds, so the
// decoder works with a slice of types rather than a single one; a
// numeric-looking raw string still decodes to the enum's number even
// when a string member exists, and an array member can never be
// produced from a parameter value.
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
		url  string
		fail bool
	}{{
		url: "?mid=42",
	}, {
		url:  "?mid=gadget",
		fail: true,
	}, {
		url:  "?mid=%5B42%5D",
		fail: true,
	}}
	for _, tc := range testcases {
		t.Run(tc.url, func(t *testing.T) {
			testIssue1367Request(t, spec, rootURL+tc.url, tc.fail)
		})
	}
}
