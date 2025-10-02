package openapi3filter

import (
	"net/http"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

func TestIssue1100(t *testing.T) {
	spec := `
openapi: 3.0.3
info:
  version: 1.0.0
  title: sample api
  description: api service paths to test the issue
paths:
  /api/path:
    post:
      summary: path
      tags:
        - api
      responses:
        '200':
          description: Ok
    `[1:]

	loader := openapi3.NewLoader()

	doc, err := loader.LoadFromData([]byte(spec))
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	router, err := gorillamux.NewRouter(doc)
	require.NoError(t, err)

	for _, testcase := range []struct {
		name       string
		endpoint   string
		ct         string
		data       string
		rejectBody bool
		shouldFail bool
	}{
		{
			name:       "json success",
			endpoint:   "/api/path",
			ct:         "application/json",
			data:       ``,
			rejectBody: false,
			shouldFail: false,
		},
		{
			name:       "json failure",
			endpoint:   "/api/path",
			ct:         "application/json",
			data:       `{"data":"some+unexpected+data"}`,
			rejectBody: false,
			shouldFail: false,
		},
		{
			name:       "json success",
			endpoint:   "/api/path",
			ct:         "application/json",
			data:       ``,
			rejectBody: true,
			shouldFail: false,
		},
		{
			name:       "json failure",
			endpoint:   "/api/path",
			ct:         "application/json",
			data:       `{"data":"some+unexpected+data"}`,
			rejectBody: true,
			shouldFail: true,
		},
	} {
		t.Run(
			testcase.name, func(t *testing.T) {
				data := strings.NewReader(testcase.data)
				req, err := http.NewRequest("POST", testcase.endpoint, data)
				require.NoError(t, err)
				req.Header.Add("Content-Type", testcase.ct)

				route, pathParams, err := router.FindRoute(req)
				require.NoError(t, err)

				validationInput := &openapi3filter.RequestValidationInput{
					Request:    req,
					PathParams: pathParams,
					Route:      route,
					Options:    &openapi3filter.Options{RejectWhenRequestBodyNotSpecified: testcase.rejectBody},
				}
				err = openapi3filter.ValidateRequest(loader.Context, validationInput)
				if testcase.shouldFail {
					require.Error(t, err, "This test case should fail "+testcase.data)
				} else {
					require.NoError(t, err, "This test case should pass "+testcase.data)
				}
			},
		)
	}
}
