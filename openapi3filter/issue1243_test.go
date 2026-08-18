package openapi3filter

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func issue1243Spec(version string) string {
	return `
openapi: ` + version + `
info:
  title: repro
  version: "1"
paths:
  /probe:
    get:
      operationId: probe
      responses:
        "200":
          description: ok
          content:
            application/json:
              schema:
                type: object
                required: [id, created_at]
                properties:
                  id: {type: string, format: uuid}
                  created_at: {type: string, format: date-time}
                  retries: {type: integer, format: int32}
`
}

func TestIssue1243StringFormatValidatorsApplyToOpenAPI31(t *testing.T) {
	for _, version := range []string{"3.0.3", "3.1.0"} {
		t.Run(version, func(t *testing.T) {
			router := setupTestRouter(t, issue1243Spec(version))

			req, err := http.NewRequest(http.MethodGet, "/probe", nil)
			require.NoError(t, err)

			route, pathParams, err := router.FindRoute(req)
			require.NoError(t, err)

			opts := &Options{
				SchemaValidationOptions: []openapi3.SchemaValidationOption{
					openapi3.EnableFormatValidation(),
					openapi3.WithStringFormatValidators(map[string]openapi3.StringFormatValidator{
						"uuid":      openapi3.NewRegexpFormatValidator(openapi3.FormatOfStringForUUIDOfRFC4122),
						"date-time": openapi3.NewRegexpFormatValidator(openapi3.FormatOfStringDateTime),
					}),
				},
			}

			validate := func(body string) error {
				return ValidateResponse(t.Context(), &ResponseValidationInput{
					RequestValidationInput: &RequestValidationInput{
						Request:    req,
						PathParams: pathParams,
						Route:      route,
						Options:    opts,
					},
					Status: http.StatusOK,
					Header: http.Header{"Content-Type": []string{"application/json"}},
					Body:   io.NopCloser(strings.NewReader(body)),
					Options: &Options{
						IncludeResponseStatus:   true,
						SchemaValidationOptions: opts.SchemaValidationOptions,
					},
				})
			}

			t.Run("valid", func(t *testing.T) {
				err := validate(`{"id": "9d1c8f74-4e5b-41f9-a2f5-6b1f0b6dbb9e", "created_at": "2017-07-21T17:32:28Z", "retries": 3}`)
				require.NoError(t, err)
			})

			t.Run("garbage uuid", func(t *testing.T) {
				err := validate(`{"id": "not-a-uuid", "created_at": "2017-07-21T17:32:28Z"}`)
				require.ErrorContains(t, err, "uuid")
			})

			t.Run("garbage date-time", func(t *testing.T) {
				err := validate(`{"id": "9d1c8f74-4e5b-41f9-a2f5-6b1f0b6dbb9e", "created_at": "yesterday"}`)
				require.ErrorContains(t, err, "date-time")
			})

			t.Run("out of range int32", func(t *testing.T) {
				err := validate(`{"id": "9d1c8f74-4e5b-41f9-a2f5-6b1f0b6dbb9e", "created_at": "2017-07-21T17:32:28Z", "retries": 2147483648}`)
				require.ErrorContains(t, err, "int32")
			})
		})
	}
}

func TestIssue1243UnregisteredFormatsStayAnnotationsUnderOpenAPI31(t *testing.T) {
	spec := `
openapi: 3.1.0
info:
  title: repro
  version: "1"
paths:
  /probe:
    get:
      operationId: probe
      responses:
        "200":
          description: ok
          content:
            application/json:
              schema:
                type: object
                properties:
                  contact: {type: string, format: email}
`
	router := setupTestRouter(t, spec)

	req, err := http.NewRequest(http.MethodGet, "/probe", nil)
	require.NoError(t, err)

	route, pathParams, err := router.FindRoute(req)
	require.NoError(t, err)

	// No validator is registered for "email" so, as with the built-in validator, the format is
	// only an annotation: enabling JSON Schema 2020-12 must not start asserting it.
	err = ValidateResponse(t.Context(), &ResponseValidationInput{
		RequestValidationInput: &RequestValidationInput{
			Request:    req,
			PathParams: pathParams,
			Route:      route,
		},
		Status:  http.StatusOK,
		Header:  http.Header{"Content-Type": []string{"application/json"}},
		Body:    io.NopCloser(strings.NewReader(`{"contact": "not an email"}`)),
		Options: &Options{IncludeResponseStatus: true},
	})
	require.NoError(t, err)
}
