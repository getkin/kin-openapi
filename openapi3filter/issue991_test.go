package openapi3filter

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateRequestDefault(t *testing.T) {
	const spec = `
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

	router := setupTestRouter(t, spec)

	type args struct {
		url      string
		expected []string
	}
	tests := []struct {
		name                 string
		args                 args
		expectedModification bool
		expectedErr          error
	}{
		{
			name: "Valid request without type parameters set",
			args: args{
				url:      "/category",
				expected: []string{"A", "B", "C"},
			},
			expectedModification: false,
			expectedErr:          nil,
		},
		{
			name: "Valid request with 1 type parameters set",
			args: args{
				url:      "/category?type=A",
				expected: []string{"A"},
			},
			expectedModification: false,
			expectedErr:          nil,
		},
		{
			name: "Valid request with 2 type parameters set",
			args: args{
				url:      "/category?type=A&type=C",
				expected: []string{"A", "C"},
			},
			expectedModification: false,
			expectedErr:          nil,
		},
		{
			name: "Valid request with 1 type parameters set out of enum",
			args: args{
				url:      "/category?type=X",
				expected: nil,
			},
			expectedModification: false,
			expectedErr:          &RequestError{},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			req, err := http.NewRequest(http.MethodGet, tc.args.url, nil)
			require.NoError(t, err)

			route, pathParams, err := router.FindRoute(req)
			require.NoError(t, err)

			validationInput := &RequestValidationInput{
				Request:    req,
				PathParams: pathParams,
				Route:      route,
			}
			err = ValidateRequest(context.Background(), validationInput)
			assert.IsType(t, tc.expectedErr, err, "ValidateRequest(): error = %v, expectedError %v", err, tc.expectedErr)
			if tc.expectedErr != nil {
				return
			}

			assert.Equal(t, tc.args.expected, req.URL.Query()["type"], "ValidateRequest(): query parameter type values do not match expected")
		})
	}
}
