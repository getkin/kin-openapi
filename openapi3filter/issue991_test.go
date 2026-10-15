package openapi3filter

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateSpec(explode bool) string {
	explodeStr := "false"
	if explode {
		explodeStr = "true"
	}
	return `
openapi: 3.0.0
info:
  title: 'Validator'
  version: 0.0.1
paths:
  /category:
    get:
      parameters:
        - $ref: "#/components/parameters/Type"
      responses:
        '200':
          description: Ok
components:
  parameters:
    Type:
      in: query
      name: type
      required: false
      explode: ` + explodeStr + `
      description: Type parameter
      schema:
        type: array
        default:
          - A
          - B
          - C
        items:
          type: string
          enum:
            - A
            - B
            - C
`
}

func TestValidateRequestDefault(t *testing.T) {
	t.Parallel()

	type args struct {
		url      string
		expected []string
	}
	tests := []struct {
		name                 string
		args                 args
		expectedModification bool
		expectedErr          error
		spec                 string
	}{
		{
			name: "Valid request without type parameters set and explode is false",
			args: args{
				url:      "/category",
				expected: []string{"A,B,C"},
			},
			expectedModification: false,
			expectedErr:          nil,
			spec:                 generateSpec(false),
		},
		{
			name: "Valid request with 1 type parameters set and explode is false",
			args: args{
				url:      "/category?type=A",
				expected: []string{"A"},
			},
			expectedModification: false,
			expectedErr:          nil,
			spec:                 generateSpec(false),
		},
		{
			name: "Valid request with 2 type parameters set and explode is false",
			args: args{
				url:      "/category?type=A&type=C",
				expected: []string{"A", "C"},
			},
			expectedModification: false,
			expectedErr:          nil,
			spec:                 generateSpec(false),
		},
		{
			name: "Valid request with 1 type parameters set out of enum and explode is false",
			args: args{
				url:      "/category?type=X",
				expected: nil,
			},
			expectedModification: false,
			expectedErr:          &RequestError{},
			spec:                 generateSpec(false),
		},
		{
			name: "Valid request without type parameters set and explode is true",
			args: args{
				url:      "/category",
				expected: []string{"A", "B", "C"},
			},
			expectedModification: false,
			expectedErr:          nil,
			spec:                 generateSpec(true),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, tc.args.url, nil)
			require.NoError(t, err)

			router := setupTestRouter(t, tc.spec)
			route, pathParams, err := router.FindRoute(req)
			require.NoError(t, err)

			validationInput := &RequestValidationInput{
				Request:    req,
				PathParams: pathParams,
				Route:      route,
			}
			err = ValidateRequest(t.Context(), validationInput)
			assert.IsType(t, tc.expectedErr, err, "ValidateRequest(): error = %v, expectedError %v", err, tc.expectedErr)
			if tc.expectedErr != nil {
				return
			}

			assert.Equal(t, tc.args.expected, req.URL.Query()["type"], "ValidateRequest(): query parameter type values do not match expected")
		})
	}
}

func TestValidateRequestDefaultSerializationMethod(t *testing.T) {
	t.Parallel()

	const spec = `
openapi: 3.0.0
info:
  title: 'Validator'
  version: 0.0.1
paths:
  /category:
    get:
      parameters:
        - in: query
          name: form
          schema:
            type: array
            default: [A, B, C]
            items:
              type: string
        - in: query
          name: spaces
          style: spaceDelimited
          schema:
            type: array
            default: [A, B, C]
            items:
              type: string
        - in: query
          name: pipes
          style: pipeDelimited
          schema:
            type: array
            default: [A, B, C]
            items:
              type: string
      responses:
        '200':
          description: Ok
`

	router := setupTestRouter(t, spec)

	req, err := http.NewRequest(http.MethodGet, "/category", nil)
	require.NoError(t, err)

	route, pathParams, err := router.FindRoute(req)
	require.NoError(t, err)

	validationInput := &RequestValidationInput{
		Request:    req,
		PathParams: pathParams,
		Route:      route,
	}
	err = ValidateRequest(t.Context(), validationInput)
	require.NoError(t, err)

	query := req.URL.Query()
	require.Equal(t, []string{"A", "B", "C"}, query["form"])
	require.Equal(t, []string{"A B C"}, query["spaces"])
	require.Equal(t, []string{"A|B|C"}, query["pipes"])

	// The populated query must decode back to the default values.
	for _, name := range []string{"form", "spaces", "pipes"} {
		param := route.Operation.Parameters.GetByInAndName("query", name)
		value, _, err := decodeStyledParameter(param, &RequestValidationInput{Request: req})
		require.NoError(t, err)
		require.Equal(t, []any{"A", "B", "C"}, value, name)
	}
}
