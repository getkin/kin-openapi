package openapi3

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExamplesValidation(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name                  string
		schemaRequestExample  string
		schemaResponseExample string
		mediaTypeRequestField string
		examples              string
		errContains           string
	}

	testCases := []testCase{
		{
			name: "invalid_component_examples",
			schemaRequestExample: `
      example:
        username: good
        email: real@email
        password: validpassword
   `,
			mediaTypeRequestField: `
            examples:
              BadUser:
                $ref: '#/components/examples/BadUser'
   `,
			examples: `
  examples:
    BadUser:
      value:
        username: "]bad["
        email: bad
        password: short
   `,
		},
		{
			name: "invalid_mediatype_examples",
			schemaRequestExample: `
      example:
        username: good
        email: real@email
        password: validpassword
   `,
			mediaTypeRequestField: `
            example:
              username: "]bad["
              email: bad
              password: short
   `,
		},
		{
			name: "invalid_schema_request_example",
			schemaRequestExample: `
      example:
        username: good
        email: good@email.com
        # missing password
   `,
			mediaTypeRequestField: `
            example:
              username: good
              email: real@email
              password: validpassword
   `,
		},
		{
			name: "invalid_schema_response_example",
			schemaResponseExample: `
      example:
        user_id: 1
        # missing access_token
   `,
			mediaTypeRequestField: `
            example:
              username: good
              email: real@email
              password: validpassword
   `,
		},
		{
			name: "example_examples_mutually_exclusive",
			mediaTypeRequestField: `
            examples:
              BadUser:
                $ref: '#/components/examples/BadUser'
            example:
              username: good
              email: real@email
              password: validpassword
`,
			errContains: "example and examples are mutually exclusive",
			examples: `
  examples:
    BadUser:
      value:
        username: "]bad["
        email: bad
        password: short
`,
		},
		{
			name: "example_without_value",
			examples: `
  examples:
    BadUser:
      description: empty user example
`,
			errContains: "example has no value or externalValue field",
		},
		{
			name: "value_externalValue_mutual_exclusion",
			examples: `
  examples:
    BadUser:
      value:
        username: good
        email: real@email
        password: validpassword
      externalValue: 'http://example.com/examples/example'
`,
			errContains: "value and externalValue are mutually exclusive",
		},
	}

	testOptions := []struct {
		name                      string
		disableExamplesValidation bool
	}{
		{
			name:                      "examples_validation_disabled",
			disableExamplesValidation: true,
		},
		{
			name:                      "examples_validation_enabled",
			disableExamplesValidation: false,
		},
	}

	for _, testOption := range testOptions {
		testOption := testOption
		t.Run(testOption.name, func(t *testing.T) {
			t.Parallel()
			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					loader := NewLoader()

					spec := bytes.Buffer{}
					spec.WriteString(`
openapi: 3.0.3
info:
  title: An API
  version: 1.2.3.4
paths:
  /user:
    post:
      description: User creation.
      operationId: createUser
      requestBody:
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/CreateUserRequest"
`)
					spec.WriteString(tc.mediaTypeRequestField)
					spec.WriteString(`
        description: Created user object
        required: true
      responses:
        '204':
          description: "success"
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/CreateUserResponse"
components:
  schemas:
    CreateUserRequest:`)
					spec.WriteString(tc.schemaRequestExample)
					spec.WriteString(`
      required:
        - username
        - email
        - password
      properties:
        username:
          type: string
          pattern: "^[ a-zA-Z0-9_-]+$"
          minLength: 3
        email:
          type: string
          pattern: "^[A-Za-z0-9+_.-]+@(.+)$"
        password:
          type: string
          minLength: 7
      type: object
    CreateUserResponse:`)
					spec.WriteString(tc.schemaResponseExample)
					spec.WriteString(`
      description: represents the response to a User creation
      required:
        - access_token
        - user_id
      properties:
        access_token:
          type: string
        user_id:
          format: int64
          type: integer
      type: object
`)
					spec.WriteString(tc.examples)

					doc, err := loader.LoadFromData(spec.Bytes())
					require.NoError(t, err)

					if testOption.disableExamplesValidation {
						err = doc.Validate(loader.Context, DisableExamplesValidation())
						require.NoError(t, err)
					} else {
						err = doc.Validate(loader.Context)
						require.Error(t, err)
						require.Contains(t, err.Error(), tc.errContains)
					}
				})
			}
		})
	}
}
