package openapi3

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExamplesSchemaValidation(t *testing.T) {
	type testCase struct {
		name                                  string
		requestSchemaExample                  string
		responseSchemaExample                 string
		mediaTypeRequestExample               string
		mediaTypeResponseExample              string
		readWriteOnlyMediaTypeRequestExample  string
		readWriteOnlyMediaTypeResponseExample string
		parametersExample                     string
		componentExamples                     string
		errContains                           string
	}

	testCases := []testCase{
		{
			name: "invalid_parameter_examples",
			parametersExample: `
          examples:
            param1example:
              value: abcd
   `,
			errContains: `invalid paths: invalid path /user: invalid operation POST: invalid example: param1example`,
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
			errContains: `invalid path /user: invalid operation POST: invalid example`,
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
			errContains: `invalid paths: invalid path /user: invalid operation POST: invalid example: example BadUser`,
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
			errContains: `invalid path /user: invalid operation POST: invalid example`,
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
			errContains: `schema "CreateUserRequest": invalid example`,
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
			errContains: `schema "CreateUserResponse": invalid example`,
		},
		{
			name: "valid_schema_response_example",
			responseSchemaExample: `
      example:
        user_id: 1
        access_token: "abcd"
  `,
		},
		{
			name: "valid_readonly_writeonly_examples",
			readWriteOnlyMediaTypeRequestExample: `
            examples:
              ReadWriteOnlyRequest:
                $ref: '#/components/examples/ReadWriteOnlyRequestData'
`,
			readWriteOnlyMediaTypeResponseExample: `
              examples:
                ReadWriteOnlyResponse:
                  $ref: '#/components/examples/ReadWriteOnlyResponseData'
`,
			componentExamples: `
  examples:
    ReadWriteOnlyRequestData:
      value:
        username: user
        password: password
    ReadWriteOnlyResponseData:
      value:
        user_id: 4321
  `,
		},
		{
			name: "invalid_readonly_request_examples",
			readWriteOnlyMediaTypeRequestExample: `
            examples:
              ReadWriteOnlyRequest:
                $ref: '#/components/examples/ReadWriteOnlyRequestData'
`,
			componentExamples: `
  examples:
    ReadWriteOnlyRequestData:
      value:
        username: user
        password: password
        user_id: 4321
`,
			errContains: `invalid example: example ReadWriteOnlyRequest: Error at "/user_id": readOnly property "user_id" in request`,
		},
		{
			name: "invalid_writeonly_response_examples",
			readWriteOnlyMediaTypeResponseExample: `
              examples:
                ReadWriteOnlyResponse:
                  $ref: '#/components/examples/ReadWriteOnlyResponseData'
`,
			componentExamples: `
  examples:
    ReadWriteOnlyResponseData:
      value:
        password: password
        user_id: 4321
`,

			errContains: `invalid example: example ReadWriteOnlyResponse: Error at "/password": writeOnly property "password" in response`,
		},
	}

	testOptions := []struct {
		name                      string
		disableExamplesValidation bool
		useDefaultOptions         bool
	}{
		{
			name:              "examples_validation_default",
			useDefaultOptions: true,
		},
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
		t.Run(testOption.name, func(t *testing.T) {
			t.Parallel()
			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					spec := bytes.NewBufferString(`
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
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/CreateUserRequest"
`)
					spec.WriteString(tc.mediaTypeRequestExample)
					spec.WriteString(`
        description: Created user object
      responses:
        '204':
          description: "success"
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/CreateUserResponse"`)
					spec.WriteString(tc.mediaTypeResponseExample)
					spec.WriteString(`
  /readWriteOnly:
    post:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/ReadWriteOnlyData"
`)
					spec.WriteString(tc.readWriteOnlyMediaTypeRequestExample)
					spec.WriteString(`
      responses:
        '201':
          description: a response
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/ReadWriteOnlyData"`)
					spec.WriteString(tc.readWriteOnlyMediaTypeResponseExample)
					spec.WriteString(`
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
    ReadWriteOnlyData:
      required:
        # only required in request
        - username
        - password
        # only required in response
        - user_id
      properties:
        username:
          type: string
          default: default
          writeOnly: true # only sent in a request
        password:
          type: string
          default: default
          writeOnly: true # only sent in a request
        user_id:
          format: int64
          default: 1
          type: integer
          readOnly: true # only returned in a response
      type: object
`)
					spec.WriteString(tc.componentExamples)

					loader := NewLoader()
					doc, err := loader.LoadFromData(spec.Bytes())
					require.NoError(t, err)

					if testOption.useDefaultOptions {
						err = doc.Validate(loader.Context)
					} else if testOption.disableExamplesValidation {
						err = doc.Validate(loader.Context, DisableExamplesValidation())
					} else {
						err = doc.Validate(loader.Context, EnableExamplesValidation())
					}

					if tc.errContains != "" && !testOption.disableExamplesValidation {
						require.Error(t, err)
						require.ErrorContains(t, err, tc.errContains)
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
			errContains: `invalid path /user: invalid operation POST: example and examples are mutually exclusive`,
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
			errContains: `invalid components: example "BadUser": no value or externalValue field`,
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
			errContains: `invalid components: example "BadUser": value and externalValue are mutually exclusive`,
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
		t.Run(testOption.name, func(t *testing.T) {
			t.Parallel()
			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					spec := bytes.NewBufferString(`
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

					loader := NewLoader()
					doc, err := loader.LoadFromData(spec.Bytes())
					require.NoError(t, err)

					if testOption.disableExamplesValidation {
						err = doc.Validate(loader.Context, DisableExamplesValidation())
					} else {
						err = doc.Validate(loader.Context)
					}

					if tc.errContains != "" {
						require.Error(t, err)
						require.ErrorContains(t, err, tc.errContains)
					} else {
						require.NoError(t, err)
					}
				})
			}
		})
	}
}

// Request and response example direction must not affect later validation with
// the same parent context, including when body validation returns an error.
func TestBodyExampleValidationContext(t *testing.T) {
	for _, direction := range []string{"request", "response"} {
		for _, invalid := range []bool{false, true} {
			name := direction
			if invalid {
				name += "_invalid"
			}
			t.Run(name, func(t *testing.T) {
				readOnly := NewStringSchema()
				readOnly.ReadOnly = true
				writeOnly := NewStringSchema()
				writeOnly.WriteOnly = true
				schema := NewObjectSchema().
					WithProperty("id", readOnly).
					WithProperty("password", writeOnly).
					WithProperty("name", NewStringSchema()).
					WithRequired([]string{"id", "password", "name"})
				example := map[string]any{"name": "example"}
				if direction == "request" {
					example["password"] = "example-password"
				} else {
					example["id"] = "example-id"
				}
				if invalid {
					delete(example, "name")
				}
				content := NewContentWithJSONSchema(schema)
				content["application/json"].Example = example
				ctx := WithValidationOptions(t.Context(), EnableSchemaFormatValidation())
				var err error
				if direction == "request" {
					err = (&RequestBody{Content: content}).Validate(ctx)
				} else {
					response := NewResponse().WithDescription("success")
					response.Content = content
					err = response.Validate(ctx)
				}
				if invalid {
					require.ErrorContains(t, err, `property "name" is missing`)
				} else {
					require.NoError(t, err)
				}

				neutral := &MediaType{
					Schema: &SchemaRef{Value: schema},
					Example: map[string]any{
						"id": "example-id", "password": "example-password", "name": "example",
					},
				}
				require.NoError(t, neutral.Validate(ctx))
				delete(neutral.Example.(map[string]any), "id")
				require.ErrorContains(t, neutral.Validate(ctx), `property "id" is missing`)
			})
		}
	}
}

// Schema examples describe reusable schema values. Only examples on a media
// type describe a request or response payload.
func TestMediaTypeReusableSchemaExamples(t *testing.T) {
	for _, direction := range []string{"request", "response"} {
		for _, shape := range []string{"root", "property", "items"} {
			for _, explicit := range []bool{false, true} {
				name := direction + "/" + shape + "/default"
				if explicit {
					name = direction + "/" + shape + "/explicit"
				}
				t.Run(name, func(t *testing.T) {
					id := NewStringSchema()
					id.ReadOnly = true
					secret := NewStringSchema()
					secret.WriteOnly = true
					value := NewObjectSchema().WithProperty("id", id).WithProperty("secret", secret).
						WithRequired([]string{"id", "secret"})
					value.Example = map[string]any{"id": "id", "secret": "secret"}
					var payload any = map[string]any{"secret": "secret"}
					if direction == "response" {
						payload = map[string]any{"id": "id"}
					}
					schema := value
					switch shape {
					case "property":
						// An omitted response-only (or request-only) property may
						// contain an example of its own reusable object schema.
						value.ReadOnly = direction == "request"
						value.WriteOnly = direction == "response"
						schema = NewObjectSchema().WithProperty("details", value)
						payload = map[string]any{}
					case "items":
						items := NewArraySchema().WithItems(value)
						schema = NewObjectSchema().WithProperty("files", items)
						payload = map[string]any{"files": []any{payload}}
					}
					mediaType := &MediaType{Schema: &SchemaRef{Value: schema}, Example: payload}
					content := Content{"application/json": mediaType}
					ctx := t.Context()
					if explicit {
						ctx = WithValidationOptions(ctx, EnableExamplesValidation())
					}
					before := *getValidationOptions(ctx)
					if direction == "request" {
						require.NoError(t, (&RequestBody{Content: content}).Validate(ctx))
					} else {
						require.NoError(t, NewResponse().WithDescription("ok").WithContent(content).Validate(ctx))
					}
					require.Equal(t, before, *getValidationOptions(ctx))
				})
			}
		}
	}
}

func TestMediaTypePayloadExampleDirection(t *testing.T) {
	for _, direction := range []string{"request", "response"} {
		for _, named := range []bool{false, true} {
			name := direction + "/example"
			if named {
				name = direction + "/examples"
			}
			t.Run(name, func(t *testing.T) {
				property := NewStringSchema()
				property.ReadOnly = direction == "request"
				property.WriteOnly = direction == "response"
				schema := NewObjectSchema().WithProperty("forbidden", property)
				schema.Example = map[string]any{"forbidden": "schema annotation"}
				mediaType := &MediaType{Schema: &SchemaRef{Value: schema}}
				payload := map[string]any{"forbidden": "payload"}
				if named {
					mediaType.Examples = Examples{"payload": &ExampleRef{Value: &Example{Value: payload}}}
				} else {
					mediaType.Example = payload
				}
				content := Content{"application/json": mediaType}
				var err error
				if direction == "request" {
					err = (&RequestBody{Content: content}).Validate(t.Context())
					require.ErrorContains(t, err, `readOnly property "forbidden" in request`)
				} else {
					err = NewResponse().WithDescription("ok").WithContent(content).Validate(t.Context())
					require.ErrorContains(t, err, `writeOnly property "forbidden" in response`)
				}
				if named {
					require.ErrorContains(t, err, "example payload:")
				}
			})
		}
	}
}

func TestMediaTypeSchemaValidationOptions(t *testing.T) {
	for _, direction := range []string{"request", "response"} {
		t.Run(direction, func(t *testing.T) {
			validate := func(schema *Schema, opts ...ValidationOption) error {
				content := NewContentWithJSONSchema(schema)
				if direction == "request" {
					return (&RequestBody{Content: content}).Validate(t.Context(), opts...)
				}
				return NewResponse().WithDescription("ok").WithContent(content).Validate(t.Context(), opts...)
			}
			schema := NewStringSchema()
			schema.Example = 42
			require.ErrorContains(t, validate(schema), "invalid example")
			require.NoError(t, validate(schema, DisableExamplesValidation()))
			schema.Example = nil
			schema.Default = 42
			require.ErrorContains(t, validate(schema), "invalid default")
			require.NoError(t, validate(schema, DisableSchemaDefaultsValidation()))
			schema.Default = nil
			schema.Format = "unknown-format"
			require.NoError(t, validate(schema))
			require.Error(t, validate(schema, EnableSchemaFormatValidation()))
			schema.Type = &Types{"invalid-type"}
			require.Error(t, validate(schema, DisableExamplesValidation()))
		})
	}
}

func TestMediaTypeItemSchemaExamples(t *testing.T) {
	readOnly := NewStringSchema()
	readOnly.ReadOnly = true
	writeOnly := NewStringSchema()
	writeOnly.WriteOnly = true
	schema := NewObjectSchema().WithProperty("id", readOnly).WithProperty("secret", writeOnly)
	schema.Example = map[string]any{"id": "id", "secret": "secret"}
	mediaType := &MediaType{ItemSchema: &SchemaRef{Value: schema}}
	content := Content{"application/json-seq": mediaType}
	request := &RequestBody{Content: content}
	require.ErrorContains(t, request.Validate(t.Context()), "itemSchema")
	ctx := WithValidationOptions(t.Context(), IsOpenAPI32OrLater())
	before := *getValidationOptions(ctx)
	require.NoError(t, request.Validate(ctx))
	require.NoError(t, NewResponse().WithDescription("ok").WithContent(content).Validate(ctx))
	require.Equal(t, before, *getValidationOptions(ctx))
	schema.Example = 42
	require.ErrorContains(t, request.Validate(ctx), "invalid example")
	require.NoError(t, request.Validate(t.Context(), IsOpenAPI32OrLater(), DisableExamplesValidation()))
}
