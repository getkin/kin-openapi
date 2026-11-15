package openapi3filter_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

// rejectEverything stands in for a user's own regex engine, the reason
// Options.RegexCompiler exists: Go's RE2 cannot compile every ECMA 262 pattern.
type rejectEverything struct{}

func (rejectEverything) MatchString(string) bool { return false }

const regexCompilerSpec = `
openapi: 3.0.3
info:
  title: regex compiler
  version: "1"
paths:
  /items:
    get:
      parameters:
        - name: q
          in: query
          required: true
          schema:
            type: string
            pattern: '^[a-z]+$'
      responses:
        "200":
          description: ok
          content:
            application/json:
              schema:
                type: string
                pattern: '^[a-z]+$'
`

func TestRegexCompilerAppliesToParameters(t *testing.T) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(regexCompilerSpec))
	require.NoError(t, err)
	require.NoError(t, doc.Validate(context.Background()))

	router, err := gorillamux.NewRouter(doc)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodGet, "http://example.com/items?q=abc", nil)
	require.NoError(t, err)
	route, pathParams, err := router.FindRoute(req)
	require.NoError(t, err)

	calls := 0
	err = openapi3filter.ValidateRequest(context.Background(), &openapi3filter.RequestValidationInput{
		Request:    req,
		PathParams: pathParams,
		Route:      route,
		Options: &openapi3filter.Options{
			RegexCompiler: func(string) (openapi3.RegexMatcher, error) {
				calls++
				return rejectEverything{}, nil
			},
		},
	})

	require.Positive(t, calls)
	require.ErrorContains(t, err, `parameter "q"`)
}

func TestRegexCompilerAppliesToResponses(t *testing.T) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(regexCompilerSpec))
	require.NoError(t, err)
	require.NoError(t, doc.Validate(context.Background()))

	router, err := gorillamux.NewRouter(doc)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodGet, "http://example.com/items?q=abc", nil)
	require.NoError(t, err)
	route, pathParams, err := router.FindRoute(req)
	require.NoError(t, err)

	calls := 0
	opts := &openapi3filter.Options{
		RegexCompiler: func(string) (openapi3.RegexMatcher, error) {
			calls++
			return rejectEverything{}, nil
		},
	}
	err = openapi3filter.ValidateResponse(context.Background(), &openapi3filter.ResponseValidationInput{
		RequestValidationInput: &openapi3filter.RequestValidationInput{
			Request:    req,
			PathParams: pathParams,
			Route:      route,
			Options:    opts,
		},
		Status: http.StatusOK,
		Header: http.Header{"Content-Type": []string{"application/json"}},
		Body:   io.NopCloser(strings.NewReader(`"abc"`)),
		Options: &openapi3filter.Options{
			RegexCompiler: opts.RegexCompiler,
		},
	})

	require.Positive(t, calls)
	require.Error(t, err)
}
