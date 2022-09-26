package openapi3

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExamplesSchemaValidation(t *testing.T) {
	type testCase struct {
		name                    string
		requestSchemaExample    string
		responseSchemaExample   string
		mediaTypeRequestExample string
		parametersExample       string
		componentExamples       string
		errContains             string
	}

	testCases := []testCase{
		{
			name: "invalid_parameter_examples",
			parametersExample: `
          examples:
            param1example:
              value: abcd
   `,
			errContains: "invalid paths: invalid path /user: invalid operation POST: param1example",
		},
		{
			name: "valid_parameter_examples",
			parametersExample: `
          examples:
            param1example:
              value: 1
   `,
		},
		{
			name: "invalid_parameter_example",
			parametersExample: `
          example: abcd
   `,
			errContains: "invalid path /user: invalid operation POST: invalid example",
		},
		{
			name: "valid_parameter_example",
			parametersExample: `
          example: 1
   `,
		},
		{
			name: "invalid_component_examples",
			mediaTypeRequestExample: `
            examples:
              BadUser:
                $ref: '#/components/examples/BadUser'
   `,
			componentExamples: `
  examples:
    BadUser:
      value:
        username: "]bad["
        email: bad
        password: short
   `,
			errContains: "invalid paths: invalid path /user: invalid operation POST: example BadUser",
		},
		{
			name: "valid_component_examples",
			mediaTypeRequestExample: `
            examples:
              BadUser:
                $ref: '#/components/examples/BadUser'
   `,
			componentExamples: `
  examples:
    BadUser:
      value:
        username: good
        email: good@mail.com
        password: password
   `,
		},
		{
			name: "invalid_mediatype_examples",
			mediaTypeRequestExample: `
            example:
              username: "]bad["
              email: bad
              password: short
   `,
			errContains: "invalid path /user: invalid operation POST: invalid example",
		},
		{
			name: "valid_mediatype_examples",
			mediaTypeRequestExample: `
            example:
              username: good
              email: good@mail.com
              password: password
   `,
		},
		{
			name: "invalid_schema_request_example",
			requestSchemaExample: `
      example:
        username: good
        email: good@email.com
        # missing password
   `,
			errContains: "schema \"CreateUserRequest\": invalid example",
		},
		{
			name: "valid_schema_request_example",
			requestSchemaExample: `
      example:
        username: good
        email: good@email.com
        password: password
   `,
		},
		{
			name: "invalid_schema_response_example",
			responseSchemaExample: `
      example:
        user_id: 1
        # missing access_token
   `,
			errContains: "schema \"CreateUserResponse\": invalid example",
		},
		{
			name: "valid_schema_response_example",
			responseSchemaExample: `
      example:
        user_id: 1
        access_token: "abcd"
   `,
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

	t.Parallel()

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
      parameters:
        - name: param1
          in: 'query'
          schema:
            format: int64
            type: integer`)
					spec.WriteString(tc.parametersExample)
					spec.WriteString(`
      requestBody:
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/CreateUserRequest"
`)
					spec.WriteString(tc.mediaTypeRequestExample)
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
					spec.WriteString(tc.requestSchemaExample)
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
					spec.WriteString(tc.responseSchemaExample)
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
					spec.WriteString(tc.componentExamples)

					doc, err := loader.LoadFromData(spec.Bytes())
					require.NoError(t, err)

					if testOption.disableExamplesValidation {
						err = doc.Validate(loader.Context, DisableExamplesValidation())
					} else {
						err = doc.Validate(loader.Context)
					}

					if tc.errContains != "" && !testOption.disableExamplesValidation {
						require.Error(t, err)
						require.Contains(t, err.Error(), tc.errContains)
					} else {
						require.NoError(t, err)
					}
				})
			}
		})
	}
}

func TestExampleObjectValidation(t *testing.T) {
	type testCase struct {
		name                    string
		mediaTypeRequestExample string
		componentExamples       string
		errContains             string
	}

	testCases := []testCase{
		{
			name: "example_examples_mutually_exclusive",
			mediaTypeRequestExample: `
            examples:
              BadUser:
                $ref: '#/components/examples/BadUser'
            example:
              username: good
              email: real@email.com
              password: validpassword
`,
			errContains: "invalid path /user: invalid operation POST: example and examples are mutually exclusive",
			componentExamples: `
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
			componentExamples: `
  examples:
    BadUser:
      description: empty user example
`,
			errContains: "invalid components: example \"BadUser\": no value or externalValue field",
		},
		{
			name: "value_externalValue_mutual_exclusion",
			componentExamples: `
  examples:
    BadUser:
      value:
        username: good
        email: real@email.com
        password: validpassword
      externalValue: 'http://example.com/examples/example'
`,
			errContains: "invalid components: example \"BadUser\": value and externalValue are mutually exclusive",
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

	t.Parallel()

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
      parameters:
        - name: param1
          in: 'query'
          schema:
            format: int64
            type: integer
      requestBody:
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/CreateUserRequest"
`)
					spec.WriteString(tc.mediaTypeRequestExample)
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
    CreateUserRequest:
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
    CreateUserResponse:
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
					spec.WriteString(tc.componentExamples)

					doc, err := loader.LoadFromData(spec.Bytes())
					require.NoError(t, err)

					if testOption.disableExamplesValidation {
						err = doc.Validate(loader.Context, DisableExamplesValidation())
					} else {
						err = doc.Validate(loader.Context)
					}

					if tc.errContains != "" {
						require.Error(t, err)
						require.Contains(t, err.Error(), tc.errContains)
					} else {
						require.NoError(t, err)
					}
				})
			}
		})
	}
}
