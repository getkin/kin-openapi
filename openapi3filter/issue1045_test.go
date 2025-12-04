package openapi3filter

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

func TestIssue1045(t *testing.T) {
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
      requestBody:
        required: true
        content:
          application/json:
            schema: { $ref: '#/components/schemas/PathRequest' }
          application/x-www-form-urlencoded:
            schema: { $ref: '#/components/schemas/PathRequest' }
      responses:
        '200':
          description: Ok
          content:
            application/json:
              schema: { $ref: '#/components/schemas/PathResponse' }
components:
  schemas:
    Msg_Opt:
      properties:
        msg: { type: string }
    Msg:
      allOf:
        - $ref: '#/components/schemas/Msg_Opt'
        - required: [ msg ]
    Name:
      properties:
        name: { type: string }
      required: [ name ]
    Id:
      properties:
        id:
          type: string
          format: uint64
      required: [ id ]
    PathRequest:
      type: object
      allOf:
        - $ref: '#/components/schemas/Msg'
        - $ref: '#/components/schemas/Name'
    PathResponse:
      type: object
      allOf:
        - $ref: '#/components/schemas/PathRequest'
        - $ref: '#/components/schemas/Id'
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
		shouldFail bool
	}{
		{
			name:       "json success",
			endpoint:   "/api/path",
			ct:         "application/json",
			data:       `{"msg":"message", "name":"some+name"}`,
			shouldFail: false,
		},
		{
			name:       "json failure",
			endpoint:   "/api/path",
			ct:         "application/json",
			data:       `{"name":"some+name"}`,
			shouldFail: true,
		},

		// application/x-www-form-urlencoded
		{
			name:       "form success",
			endpoint:   "/api/path",
			ct:         "application/x-www-form-urlencoded",
			data:       "msg=message&name=some+name",
			shouldFail: false,
		},
		{
			name:       "form failure",
			endpoint:   "/api/path",
			ct:         "application/x-www-form-urlencoded",
			data:       "name=some+name",
			shouldFail: true,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			data := strings.NewReader(testcase.data)
			req, err := http.NewRequest("POST", testcase.endpoint, data)
			require.NoError(t, err)
			req.Header.Add("Content-Type", testcase.ct)

			route, pathParams, err := router.FindRoute(req)
			require.NoError(t, err)

			validationInput := &RequestValidationInput{
				Request:    req,
				PathParams: pathParams,
				Route:      route,
			}
			err = ValidateRequest(loader.Context, validationInput)
			if testcase.shouldFail {
				require.Error(t, err, "This test case should fail "+testcase.data)
			} else {
				require.NoError(t, err, "This test case should pass "+testcase.data)
			}
		})
	}
}
