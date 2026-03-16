package openapi3filter

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/oasdiff/kin-openapi/openapi3"
	"github.com/oasdiff/kin-openapi/routers/gorillamux"
)

func TestIssue1110(t *testing.T) {
	// Test case: POST with application/x-www-form-urlencoded
	// Schema has two optional properties (not required)
	// Sending only one param should be valid since no fields are required
	spec := `
openapi: 3.0.3
info:
  version: 1.0.0
  title: Test API
paths:
  /test:
    post:
      summary: Test endpoint with optional form parameters
      requestBody:
        content:
          application/x-www-form-urlencoded:
            schema:
              type: object
              properties:
                param1:
                  type: string
                param2:
                  type: string
      responses:
        '200':
          description: OK
`[1:]

	loader := openapi3.NewLoader()
	ctx := loader.Context

	doc, err := loader.LoadFromData([]byte(spec))
	require.NoError(t, err)

	err = doc.Validate(ctx)
	require.NoError(t, err)

	router, err := gorillamux.NewRouter(doc)
	require.NoError(t, err)

	for _, testcase := range []struct {
		name       string
		data       string
		shouldFail bool
	}{
		{
			name:       "empty body should be valid (no required fields)",
			data:       "",
			shouldFail: false,
		},
		{
			name:       "only param1 present, param2 absent - should be valid",
			data:       "param1=value1",
			shouldFail: false,
		},
		{
			name:       "only param2 present, param1 absent - should be valid",
			data:       "param2=value2",
			shouldFail: false,
		},
		{
			name:       "both params present - should be valid",
			data:       "param1=value1&param2=value2",
			shouldFail: false,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/test", strings.NewReader(testcase.data))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			route, pathParams, err := router.FindRoute(req)
			require.NoError(t, err)

			validationInput := &RequestValidationInput{
				Request:    req,
				PathParams: pathParams,
				Route:      route,
			}
			err = ValidateRequest(ctx, validationInput)
			if testcase.shouldFail {
				require.Error(t, err, "This test should fail: "+testcase.name)
			} else {
				require.NoError(t, err, "This test should pass: "+testcase.name)
			}
		})
	}
}
